package sendmail

import (
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"io"
	mailaddress "net/mail"
	"strconv"
	"strings"
	"time"

	mailmodel "github.com/wtg42/hermes/mail"
)

// StructuredSendOptions contains the values accepted by the non-interactive send command.
type StructuredSendOptions struct {
	Server                  string
	Port                    string
	From                    string
	To                      []string
	CC                      []string
	BCC                     []string
	Subject                 string
	Body                    string
	Attachments             []string
	ConfirmOutsideWhitelist bool
	NoHistory               bool
	TLSMode                 string
	TLSServerName           string
	AuthMode                string
	AuthUsername            string
	AuthPasswordStdin       bool
	AuthPassword            string
}

// StructuredSendPlan is the complete immutable input for one send attempt.
type StructuredSendPlan struct {
	Compose   mailmodel.MailCompose
	Transport SMTPTransportConfig
}

// StructuredSendAttempt performs one I/O attempt using a complete plan.
type StructuredSendAttempt func(StructuredSendPlan) error

// StructuredSender validates and resolves a structured send before invoking a Mailer.
type StructuredSender struct {
	attempt StructuredSendAttempt
	now     func() time.Time
	entropy io.Reader
}

// StructuredSendExecution exposes the resolved message and one Mailer attempt.
type StructuredSendExecution struct {
	Plan            StructuredSendPlan
	MailerAttempted bool
	MailerError     error
}

// NewStructuredSender creates a sender backed by one plan-aware attempt.
func NewStructuredSender(attempt StructuredSendAttempt) *StructuredSender {
	return &StructuredSender{
		attempt: attempt,
		now:     time.Now,
		entropy: cryptorand.Reader,
	}
}

// Send resolves defaults, validates safety policy and attachments, then sends once.
func (s *StructuredSender) Send(options StructuredSendOptions) error {
	_, err := s.Execute(options)
	return err
}

// Execute resolves and validates one send, then reports its concrete Mailer attempt.
func (s *StructuredSender) Execute(options StructuredSendOptions) (StructuredSendExecution, error) {
	plan, err := s.BuildPlan(options)
	if err != nil {
		return StructuredSendExecution{}, err
	}
	if s.attempt == nil {
		return StructuredSendExecution{}, fmt.Errorf("structured send attempt is required")
	}
	execution := StructuredSendExecution{Plan: plan, MailerAttempted: true}
	execution.MailerError = s.attempt(plan)
	return execution, execution.MailerError
}

// BuildPlan resolves and validates all data before any send attempt.
func (s *StructuredSender) BuildPlan(options StructuredSendOptions) (StructuredSendPlan, error) {
	resolved, err := resolveStructuredSend(options)
	if err != nil {
		return StructuredSendPlan{}, err
	}

	if err := validateStructuredSafety(resolved); err != nil {
		return StructuredSendPlan{}, err
	}
	transport := resolved.transportConfig()
	if err := transport.validateReady(); err != nil {
		return StructuredSendPlan{}, fmt.Errorf("transport preflight: %w", err)
	}

	resolved.Attachments = normalizeAttachmentPaths("", resolved.Attachments)
	if _, err := loadAttachments(resolved.Attachments); err != nil {
		return StructuredSendPlan{}, err
	}

	if resolved.Subject == "" || resolved.Body == "" {
		randomSubject, randomBody, err := generateRandomContent(s.now(), s.entropy)
		if err != nil {
			return StructuredSendPlan{}, fmt.Errorf("failed to generate random mail content: %w", err)
		}
		if resolved.Subject == "" {
			resolved.Subject = randomSubject
		}
		if resolved.Body == "" {
			resolved.Body = randomBody
		}
	}

	compose := mailmodel.MailCompose{
		From:        resolved.From,
		To:          append([]string(nil), resolved.To...),
		CC:          append([]string(nil), resolved.CC...),
		BCC:         append([]string(nil), resolved.BCC...),
		Subject:     resolved.Subject,
		Body:        resolved.Body,
		Attachments: append([]string(nil), resolved.Attachments...),
		Host:        resolved.Server,
		Port:        resolved.Port,
	}
	return StructuredSendPlan{Compose: compose, Transport: transport}, nil
}

// SendStructuredWithHistory sends once and records only actual Mailer attempts.
func SendStructuredWithHistory(sender *StructuredSender, options StructuredSendOptions, historyPath string) error {
	if sender == nil {
		return fmt.Errorf("structured sender is required")
	}
	if options.NoHistory {
		return sender.Send(options)
	}

	execution, sendErr := sender.Execute(options)
	if !execution.MailerAttempted {
		return sendErr
	}
	record, recordErr := NewHistoryRecord(execution.Plan, execution.MailerError, time.Now(), cryptorand.Reader)
	if recordErr == nil {
		recordErr = AppendHistory(historyPath, record)
	}
	if recordErr == nil {
		return sendErr
	}
	historyErr := fmt.Errorf("history write failed: %w", recordErr)
	if sendErr == nil {
		return fmt.Errorf("mail was sent but %w", historyErr)
	}
	return errors.Join(sendErr, historyErr)
}

func resolveStructuredSend(options StructuredSendOptions) (StructuredSendOptions, error) {
	resolved := options
	if !isDottedIPv4(resolved.Server) {
		return StructuredSendOptions{}, fmt.Errorf("server must be an explicit dotted IPv4 address")
	}

	if resolved.Port == "" {
		resolved.Port = "25"
	}
	if resolved.TLSMode == "" {
		resolved.TLSMode = TLSModeNone
	}
	if resolved.AuthMode == "" {
		resolved.AuthMode = AuthModeNone
	}
	if err := resolved.transportConfig().ValidateStatic(); err != nil {
		return StructuredSendOptions{}, fmt.Errorf("transport preflight: %w", err)
	}
	port, err := strconv.Atoi(resolved.Port)
	if err != nil || port < 1 || port > 65535 {
		return StructuredSendOptions{}, fmt.Errorf("port must be an integer between 1 and 65535")
	}

	if resolved.From == "" {
		resolved.From = safeSingleSenders[0]
	}
	resolved.From, err = validateMailbox("from", resolved.From)
	if err != nil {
		return StructuredSendOptions{}, err
	}

	if len(resolved.To) == 0 {
		resolved.To = []string{defaultStructuredRecipient(resolved.From)}
	}
	resolved.To, err = validateMailboxList("to", resolved.To)
	if err != nil {
		return StructuredSendOptions{}, err
	}
	resolved.CC, err = validateMailboxList("cc", resolved.CC)
	if err != nil {
		return StructuredSendOptions{}, err
	}
	resolved.BCC, err = validateMailboxList("bcc", resolved.BCC)
	if err != nil {
		return StructuredSendOptions{}, err
	}

	return resolved, nil
}

func (options StructuredSendOptions) transportConfig() SMTPTransportConfig {
	return SMTPTransportConfig{
		Server: options.Server, Port: options.Port, TLSMode: options.TLSMode,
		TLSServerName: options.TLSServerName, AuthMode: options.AuthMode,
		AuthUsername: options.AuthUsername, PasswordSource: options.AuthPasswordStdin,
		Password: options.AuthPassword,
	}
}

// ValidateStructuredSendStatic validates options without reading secrets or dialing.
func ValidateStructuredSendStatic(options StructuredSendOptions) error {
	resolved, err := resolveStructuredSend(options)
	if err != nil {
		return err
	}
	if err := validateStructuredSafety(resolved); err != nil {
		return err
	}
	if _, err := loadAttachments(normalizeAttachmentPaths("", resolved.Attachments)); err != nil {
		return err
	}
	return nil
}

// ValidateStructuredTransportStatic validates resolved fields and transport combinations.
func ValidateStructuredTransportStatic(options StructuredSendOptions) error {
	_, err := resolveStructuredSend(options)
	return err
}

func isDottedIPv4(server string) bool {
	parts := strings.Split(server, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
		value, err := strconv.Atoi(part)
		if err != nil || value > 255 {
			return false
		}
	}
	return true
}

func validateMailbox(field, value string) (string, error) {
	address, err := mailaddress.ParseAddress(strings.TrimSpace(value))
	if err != nil || address.Address == "" {
		return "", fmt.Errorf("invalid %s email address %q", field, value)
	}
	return address.Address, nil
}

func validateMailboxList(field string, values []string) ([]string, error) {
	validated := make([]string, 0, len(values))
	for _, value := range values {
		address, err := validateMailbox(field, value)
		if err != nil {
			return nil, err
		}
		validated = append(validated, address)
	}
	return validated, nil
}

func defaultStructuredRecipient(from string) string {
	for i := len(safeSingleSenders) - 1; i >= 0; i-- {
		if !strings.EqualFold(safeSingleSenders[i], from) {
			return safeSingleSenders[i]
		}
	}
	return safeSingleSenders[0]
}

func validateStructuredSafety(options StructuredSendOptions) error {
	assessment, err := AssessSingleSendSafety(SingleSendSafetyInput{
		From: options.From,
		To:   options.To,
		CC:   options.CC,
		BCC:  options.BCC,
		Port: options.Port,
	})
	if err != nil {
		return err
	}

	if assessment.RequiresConfirmation() && !options.ConfirmOutsideWhitelist {
		return fmt.Errorf("send requires --confirm-outside-whitelist: %s", strings.Join(assessment.Reasons, "; "))
	}
	return nil
}

func generateRandomContent(now time.Time, entropy io.Reader) (string, string, error) {
	random := make([]byte, 8)
	if _, err := io.ReadFull(entropy, random); err != nil {
		return "", "", err
	}

	emojis := []string{"🚀", "📨", "🧪", "✨"}
	activities := []string{"delivery check", "routing probe", "MIME test", "SMTP check"}
	subject := fmt.Sprintf("%s 中文 English %s", emojis[int(random[0])%len(emojis)], activities[int(random[1])%len(activities)])
	taipei := time.FixedZone("Asia/Taipei", 8*60*60)
	timestamp := now.In(taipei).Format("2006-01-02 15:04:05 -07:00")
	traceID := fmt.Sprintf("%x", random[4:8])
	body := fmt.Sprintf("Hermes 中文測試郵件 / English test message\nTaipei-Time: %s\nTrace-ID: %s", timestamp, traceID)

	return subject, body, nil
}
