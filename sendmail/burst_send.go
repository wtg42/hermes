package sendmail

import (
	"bytes"
	"fmt"
	"net/mail"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"math/rand"

	"github.com/wtg42/hermes/utils"
)

const defaultSafeBurstDomain = "rd01.softnext.com.tw"

var burstDomainPattern = regexp.MustCompile(`(?i)^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

// BurstOptions defines the SMTP target and address controls for a burst run.
type BurstOptions struct {
	Quantity       int
	Host           string
	Port           string
	From           string
	To             string
	Domains        []string
	AllowedDomains []string
}

type burstPlan struct {
	quantity  int
	address   string
	fixedFrom string
	fixedTo   string
	domains   []string
}

// BurstModeSendMail 瘋狂發送郵件 - 用於壓力測試
// It validates every address and domain before starting any goroutine.
func BurstModeSendMail(options BurstOptions) error {
	plan, err := buildBurstPlan(options)
	if err != nil {
		return err
	}

	mailPool := buildBurstMailPool(plan)
	workerCount := min(runtime.NumCPU(), plan.quantity)
	tasksPerWorker := plan.quantity / workerCount
	remainder := plan.quantity % workerCount
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	reportError := func(err error) {
		select {
		case errCh <- err:
		default:
		}
	}

	for worker := range workerCount {
		total := tasksPerWorker
		if worker == 0 {
			total += remainder
		}

		wg.Add(1)
		go func(worker int, total int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(worker)))

			for range total {
				from := plan.fixedFrom
				if from == "" {
					from = mailPool[r.Intn(len(mailPool))]
				}

				to := plan.fixedTo
				if to == "" {
					to = mailPool[r.Intn(len(mailPool))]
				}

				data := EmailData{
					Host:     options.Host,
					Port:     options.Port,
					From:     from,
					To:       []string{to},
					Cc:       []string{},
					Bcc:      []string{},
					Subject:  utils.RandomString(10),
					Contents: utils.RandomString(50),
				}

				email := new(bytes.Buffer)
				email.WriteString(buildEmailHeaders(data))
				if err := buildMIMEContent(email, data.Contents); err != nil {
					reportError(fmt.Errorf("build burst email: %w", err))
					continue
				}

				if err := SendMail(plan.address, nil, from, []string{to}, email.Bytes()); err != nil {
					reportError(fmt.Errorf("send burst email: %w", err))
				}
			}
		}(worker, total)
	}

	wg.Wait()
	close(errCh)

	if err := <-errCh; err != nil {
		return err
	}

	return nil
}

func buildBurstPlan(options BurstOptions) (burstPlan, error) {
	if options.Quantity <= 0 {
		return burstPlan{}, fmt.Errorf("quantity must be greater than zero")
	}
	if strings.TrimSpace(options.Host) == "" {
		return burstPlan{}, fmt.Errorf("host is required")
	}
	if strings.TrimSpace(options.Port) == "" {
		return burstPlan{}, fmt.Errorf("port is required")
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

	return burstPlan{
		quantity:  options.Quantity,
		address:   strings.TrimSpace(options.Host) + ":" + strings.TrimSpace(options.Port),
		fixedFrom: fixedFrom,
		fixedTo:   fixedTo,
		domains:   domains,
	}, nil
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
