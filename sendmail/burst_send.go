package sendmail

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/mail"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wtg42/hermes/utils"
)

const (
	defaultSafeBurstDomain    = "rd01.softnext.com.tw"
	defaultBurstSubjectPrefix = "HERMES-BURST"
	burstConfirmationLimit    = 1000
	maxBurstBodyKB            = 1024
	maxBurstWorkers           = 256
	maxBurstRate              = 10000
	burstProgressEvery        = 100
)

var (
	burstDomainPattern = regexp.MustCompile(`(?i)^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	burstRunIDPattern  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
)

// BurstOptions defines the SMTP target and address controls for a burst run.
type BurstOptions struct {
	Quantity       int
	Host           string
	Port           string
	From           string
	To             string
	Domains        []string
	AllowedDomains []string
	RunID          string
	SubjectPrefix  string
	BodyKB         int
	Workers        int
	RatePerSecond  int
	ConfirmBurst   bool
	Progress       func(BurstProgress)
}

type burstPlan struct {
	quantity      int
	address       string
	fixedFrom     string
	fixedTo       string
	domains       []string
	runID         string
	subjectPrefix string
	bodyKB        int
	workers       int
	ratePerSecond int
	progress      func(BurstProgress)
}

// BurstResult is the final outcome of a burst run. Succeeded means the SMTP
// server accepted the message from net/smtp.SendMail without returning an error.
type BurstResult struct {
	RunID     string
	Requested int
	Attempted int
	Succeeded int
	Failed    int
	Duration  time.Duration
}

// BurstProgress reports completed SMTP attempts during a burst run.
type BurstProgress struct {
	RunID     string
	Requested int
	Attempted int
	Succeeded int
	Failed    int
}

type burstRecorder struct {
	mu           sync.Mutex
	runID        string
	requested    int
	attempted    int
	succeeded    int
	failed       int
	firstErr     error
	lastReported int
	progress     func(BurstProgress)
}

type burstRateLimiter struct {
	ticker *time.Ticker
	once   sync.Once
}

// BurstModeSendMail is retained for callers that only need an error result.
func BurstModeSendMail(options BurstOptions) error {
	_, err := ExecuteBurst(options)
	return err
}

// ExecuteBurst validates the complete run before starting workers, sends every
// requested message, and returns success/failure statistics.
func ExecuteBurst(options BurstOptions) (BurstResult, error) {
	startedAt := time.Now()
	plan, err := buildBurstPlan(options)
	if err != nil {
		return BurstResult{Requested: options.Quantity}, err
	}

	mailPool := buildBurstMailPool(plan)
	tasks := make(chan int)
	limiter := newBurstRateLimiter(plan.ratePerSecond)
	defer limiter.stop()
	recorder := &burstRecorder{
		runID:     plan.runID,
		requested: plan.quantity,
		progress:  plan.progress,
	}
	var wg sync.WaitGroup

	for worker := range plan.workers {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(worker)))

			for sequence := range tasks {
				limiter.wait()

				from := plan.fixedFrom
				if from == "" {
					from = mailPool[r.Intn(len(mailPool))]
				}

				to := plan.fixedTo
				if to == "" {
					to = mailPool[r.Intn(len(mailPool))]
				}

				subject := fmt.Sprintf("%s-%s-%06d", plan.subjectPrefix, plan.runID, sequence)
				messageID := fmt.Sprintf("hermes-%s-%06d@%s", plan.runID, sequence, burstAddressDomain(from))
				contents := buildBurstBody(plan.runID, sequence, plan.bodyKB)
				email := new(bytes.Buffer)
				email.WriteString(buildBurstHeaders(from, to, subject, messageID, time.Now()))
				if err := buildMIMEContent(email, contents); err != nil {
					recorder.record(fmt.Errorf("build burst email %d: %w", sequence, err))
					continue
				}

				if err := SendMail(plan.address, nil, from, []string{to}, email.Bytes()); err != nil {
					recorder.record(fmt.Errorf("send burst email %d: %w", sequence, err))
					continue
				}
				recorder.record(nil)
			}
		}(worker)
	}

	for sequence := 1; sequence <= plan.quantity; sequence++ {
		tasks <- sequence
	}
	close(tasks)
	wg.Wait()
	result, firstErr := recorder.result(time.Since(startedAt))
	if result.Failed > 0 {
		return result, fmt.Errorf("burst completed with %d failed of %d: %w", result.Failed, result.Attempted, firstErr)
	}

	return result, nil
}

func buildBurstPlan(options BurstOptions) (burstPlan, error) {
	if options.Quantity <= 0 {
		return burstPlan{}, fmt.Errorf("quantity must be greater than zero")
	}
	if options.Quantity > burstConfirmationLimit && !options.ConfirmBurst {
		return burstPlan{}, fmt.Errorf("quantity %d exceeds safe limit %d; repeat with --confirm-burst", options.Quantity, burstConfirmationLimit)
	}
	if strings.TrimSpace(options.Host) == "" {
		return burstPlan{}, fmt.Errorf("host is required")
	}
	if strings.TrimSpace(options.Port) == "" {
		return burstPlan{}, fmt.Errorf("port is required")
	}

	runID, err := normalizeBurstRunID(options.RunID)
	if err != nil {
		return burstPlan{}, err
	}
	subjectPrefix, err := normalizeBurstSubjectPrefix(options.SubjectPrefix)
	if err != nil {
		return burstPlan{}, err
	}
	if options.BodyKB < 0 || options.BodyKB > maxBurstBodyKB {
		return burstPlan{}, fmt.Errorf("body-kb must be between 0 and %d", maxBurstBodyKB)
	}
	if options.Workers < 0 || options.Workers > maxBurstWorkers {
		return burstPlan{}, fmt.Errorf("workers must be between 0 and %d", maxBurstWorkers)
	}
	if options.RatePerSecond < 0 || options.RatePerSecond > maxBurstRate {
		return burstPlan{}, fmt.Errorf("rate must be between 0 and %d messages per second", maxBurstRate)
	}

	fixedFrom, fromDomain, err := parseBurstAddress("from", options.From)
	if err != nil {
		return burstPlan{}, err
	}
	fixedTo, toDomain, err := parseBurstAddress("to", options.To)
	if err != nil {
		return burstPlan{}, err
	}

	domains, err := normalizeBurstDomains("domain", options.Domains)
	if err != nil {
		return burstPlan{}, err
	}
	allowedDomains, err := normalizeBurstDomains("allow-domain", options.AllowedDomains)
	if err != nil {
		return burstPlan{}, err
	}

	needsRandomAddress := fixedFrom == "" || fixedTo == ""
	if needsRandomAddress && len(domains) == 0 {
		return burstPlan{}, fmt.Errorf("at least one --domain is required when from or to is random")
	}

	usedDomains := make(map[string]struct{})
	if fromDomain != "" {
		usedDomains[fromDomain] = struct{}{}
	}
	if toDomain != "" {
		usedDomains[toDomain] = struct{}{}
	}
	if needsRandomAddress {
		for _, domain := range domains {
			usedDomains[domain] = struct{}{}
		}
	}

	allowed := map[string]struct{}{defaultSafeBurstDomain: {}}
	for _, domain := range allowedDomains {
		allowed[domain] = struct{}{}
	}

	unauthorized := make([]string, 0)
	for domain := range usedDomains {
		if _, ok := allowed[domain]; !ok {
			unauthorized = append(unauthorized, domain)
		}
	}
	if len(unauthorized) > 0 {
		sort.Strings(unauthorized)
		return burstPlan{}, fmt.Errorf("unauthorized domains: %s; authorize each with --allow-domain", strings.Join(unauthorized, ", "))
	}

	workers := options.Workers
	if workers == 0 {
		workers = runtime.NumCPU()
	}
	workers = min(workers, options.Quantity)

	return burstPlan{
		quantity:      options.Quantity,
		address:       strings.TrimSpace(options.Host) + ":" + strings.TrimSpace(options.Port),
		fixedFrom:     fixedFrom,
		fixedTo:       fixedTo,
		domains:       domains,
		runID:         runID,
		subjectPrefix: subjectPrefix,
		bodyKB:        options.BodyKB,
		workers:       workers,
		ratePerSecond: options.RatePerSecond,
		progress:      options.Progress,
	}, nil
}

func normalizeBurstRunID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = time.Now().UTC().Format("20060102T150405.000000000Z") + "-" + utils.RandomString(6)
	}
	if !burstRunIDPattern.MatchString(value) {
		return "", fmt.Errorf("run-id must start with an alphanumeric character and contain at most 64 alphanumeric, dot, underscore, or hyphen characters")
	}
	return value, nil
}

func normalizeBurstSubjectPrefix(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = defaultBurstSubjectPrefix
	}
	if len(value) > 128 || strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf("subject-prefix must not contain line breaks and must be at most 128 bytes")
	}
	return value, nil
}

func parseBurstAddress(field string, value string) (string, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", nil
	}

	address, err := mail.ParseAddress(value)
	if err != nil {
		return "", "", fmt.Errorf("invalid --%s address %q: %w", field, value, err)
	}

	at := strings.LastIndex(address.Address, "@")
	if at <= 0 || at == len(address.Address)-1 {
		return "", "", fmt.Errorf("invalid --%s address %q", field, value)
	}

	domain := strings.ToLower(address.Address[at+1:])
	if !isValidBurstDomain(domain) {
		return "", "", fmt.Errorf("invalid --%s address %q: invalid domain %q", field, value, domain)
	}

	return address.Address, domain, nil
}

func normalizeBurstDomains(field string, values []string) ([]string, error) {
	domains := make([]string, 0, len(values))
	seen := make(map[string]struct{})

	for _, value := range values {
		domain := strings.ToLower(strings.TrimSpace(value))
		if domain == "" || !isValidBurstDomain(domain) {
			return nil, fmt.Errorf("invalid --%s domain %q", field, value)
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		domains = append(domains, domain)
	}

	return domains, nil
}

func isValidBurstDomain(domain string) bool {
	return len(domain) <= 253 && burstDomainPattern.MatchString(domain)
}

func buildBurstMailPool(plan burstPlan) []string {
	if plan.fixedFrom != "" && plan.fixedTo != "" {
		return nil
	}

	const totalEmailNum = 100
	mailPool := make([]string, totalEmailNum)
	for i := range mailPool {
		mailPool[i] = utils.RandomEmail(plan.domains)
	}

	return mailPool
}

func buildBurstHeaders(from string, to string, subject string, messageID string, sentAt time.Time) string {
	return fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMessage-ID: <%s>\r\n",
		from,
		to,
		encodeRFC2047(subject),
		sentAt.Format(time.RFC1123Z),
		messageID,
	)
}

func buildBurstBody(runID string, sequence int, bodyKB int) string {
	metadata := fmt.Sprintf("Hermes burst diagnostic\nRun-ID: %s\nSequence: %06d\n", runID, sequence)
	targetBytes := bodyKB * 1024
	if targetBytes == 0 {
		targetBytes = len(metadata) + 50
	}
	if targetBytes <= len(metadata) {
		return metadata
	}
	return metadata + utils.RandomString(targetBytes-len(metadata))
}

func burstAddressDomain(address string) string {
	at := strings.LastIndex(address, "@")
	if at < 0 || at == len(address)-1 {
		return defaultSafeBurstDomain
	}
	return strings.ToLower(address[at+1:])
}

func burstInterval(ratePerSecond int) time.Duration {
	if ratePerSecond <= 0 {
		return 0
	}
	return time.Second / time.Duration(ratePerSecond)
}

func newBurstRateLimiter(ratePerSecond int) *burstRateLimiter {
	interval := burstInterval(ratePerSecond)
	if interval == 0 {
		return &burstRateLimiter{}
	}
	return &burstRateLimiter{ticker: time.NewTicker(interval)}
}

func (limiter *burstRateLimiter) wait() {
	if limiter.ticker == nil {
		return
	}
	first := false
	limiter.once.Do(func() {
		first = true
	})
	if !first {
		<-limiter.ticker.C
	}
}

func (limiter *burstRateLimiter) stop() {
	if limiter.ticker != nil {
		limiter.ticker.Stop()
	}
}

func (recorder *burstRecorder) record(err error) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	recorder.attempted++
	if err == nil {
		recorder.succeeded++
	} else {
		recorder.failed++
		if recorder.firstErr == nil {
			recorder.firstErr = err
		}
	}

	if recorder.progress == nil {
		return
	}
	if recorder.attempted%burstProgressEvery != 0 && recorder.attempted != recorder.requested {
		return
	}
	if recorder.lastReported == recorder.attempted {
		return
	}
	recorder.lastReported = recorder.attempted
	recorder.progress(BurstProgress{
		RunID:     recorder.runID,
		Requested: recorder.requested,
		Attempted: recorder.attempted,
		Succeeded: recorder.succeeded,
		Failed:    recorder.failed,
	})
}

func (recorder *burstRecorder) result(duration time.Duration) (BurstResult, error) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	return BurstResult{
		RunID:     recorder.runID,
		Requested: recorder.requested,
		Attempted: recorder.attempted,
		Succeeded: recorder.succeeded,
		Failed:    recorder.failed,
		Duration:  duration,
	}, recorder.firstErr
}
