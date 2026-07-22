package sendmail

import (
	"bufio"
	"bytes"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wtg42/hermes/mail"
)

// HistoryRecordVersion is the JSON schema version supported by Hermes.
const HistoryRecordVersion = 2

// RecordedSendRequest is the resolved, replayable subset of one structured send.
type RecordedSendRequest struct {
	Server        string   `json:"server"`
	Port          string   `json:"port"`
	From          string   `json:"from"`
	To            []string `json:"to"`
	CC            []string `json:"cc,omitempty"`
	BCC           []string `json:"bcc,omitempty"`
	Subject       string   `json:"subject"`
	Body          string   `json:"body"`
	Attachments   []string `json:"attachments,omitempty"`
	TLSMode       string   `json:"tls_mode"`
	TLSServerName string   `json:"tls_server_name,omitempty"`
	AuthMode      string   `json:"auth_mode"`
	AuthUsername  string   `json:"auth_username,omitempty"`
}

// RecordedSendResult describes the outcome of the single Mailer attempt.
type RecordedSendResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// HistoryRecord is one versioned JSONL entry for a structured send attempt.
type HistoryRecord struct {
	Version   int                 `json:"version"`
	ID        string              `json:"id"`
	CreatedAt time.Time           `json:"created_at"`
	Request   RecordedSendRequest `json:"request"`
	Result    RecordedSendResult  `json:"result"`
}

// HistorySummary is the non-sensitive projection returned by history list.
type HistorySummary struct {
	ID        string             `json:"id"`
	CreatedAt time.Time          `json:"created_at"`
	Server    string             `json:"server"`
	Port      string             `json:"port"`
	From      string             `json:"from"`
	To        []string           `json:"to"`
	Subject   string             `json:"subject"`
	Result    RecordedSendResult `json:"result"`
}

var historyAppendMu sync.Mutex

// NewHistoryRecord creates one history record from a resolved compose and Mailer result.
func NewHistoryRecord(plan StructuredSendPlan, sendErr error, now time.Time, entropy io.Reader) (HistoryRecord, error) {
	if entropy == nil {
		entropy = cryptorand.Reader
	}
	random := make([]byte, 8)
	if _, err := io.ReadFull(entropy, random); err != nil {
		return HistoryRecord{}, fmt.Errorf("generate history id: %w", err)
	}
	result := RecordedSendResult{Success: sendErr == nil}
	transport := plan.Transport
	if sendErr != nil {
		result.Error = sendErr.Error()
		if transport.Password != "" {
			result.Error = strings.ReplaceAll(result.Error, transport.Password, "[REDACTED]")
		}
	}
	record := HistoryRecord{
		Version:   HistoryRecordVersion,
		ID:        hex.EncodeToString(random),
		CreatedAt: now.UTC(),
		Request:   recordedSendRequest(plan.Compose, transport),
		Result:    result,
	}
	if err := validateHistoryRecord(record); err != nil {
		return HistoryRecord{}, err
	}
	return record, nil
}

func recordedSendRequest(compose mail.MailCompose, transport SMTPTransportConfig) RecordedSendRequest {
	request := RecordedSendRequestFromCompose(compose)
	request.TLSMode = transport.TLSMode
	request.TLSServerName = transport.TLSServerName
	request.AuthMode = transport.AuthMode
	request.AuthUsername = transport.AuthUsername
	return request
}

// RecordedSendRequestFromCompose copies replayable resolved message fields.
func RecordedSendRequestFromCompose(compose mail.MailCompose) RecordedSendRequest {
	attachments := normalizeAttachmentPaths(compose.Attachment, compose.Attachments)
	return RecordedSendRequest{
		Server:      compose.Host,
		Port:        compose.Port,
		From:        compose.From,
		To:          append([]string(nil), compose.To...),
		CC:          append([]string(nil), compose.CC...),
		BCC:         append([]string(nil), compose.BCC...),
		Subject:     compose.Subject,
		Body:        compose.Body,
		Attachments: append([]string(nil), attachments...),
		TLSMode:     TLSModeNone,
		AuthMode:    AuthModeNone,
	}
}

// ValidateHistoryRecords validates schema fields and record ID uniqueness.
func ValidateHistoryRecords(records []HistoryRecord) error {
	seen := make(map[string]struct{}, len(records))
	for index, record := range records {
		if err := validateHistoryRecord(record); err != nil {
			return fmt.Errorf("record %d: %w", index+1, err)
		}
		if _, exists := seen[record.ID]; exists {
			return fmt.Errorf("record %d: duplicate history id %q", index+1, record.ID)
		}
		seen[record.ID] = struct{}{}
	}
	return nil
}

func validateHistoryRecord(record HistoryRecord) error {
	if record.Version != HistoryRecordVersion {
		return fmt.Errorf("unsupported history version %d", record.Version)
	}
	if strings.TrimSpace(record.ID) == "" {
		return errors.New("history id is required")
	}
	if record.CreatedAt.IsZero() {
		return errors.New("created_at is required")
	}
	if strings.TrimSpace(record.Request.Server) == "" {
		return errors.New("request server is required")
	}
	if strings.TrimSpace(record.Request.Port) == "" {
		return errors.New("request port is required")
	}
	if strings.TrimSpace(record.Request.From) == "" {
		return errors.New("request from is required")
	}
	if len(record.Request.To) == 0 {
		return errors.New("request to is required")
	}
	transport := SMTPTransportConfig{
		Server: record.Request.Server, Port: record.Request.Port,
		TLSMode: record.Request.TLSMode, TLSServerName: record.Request.TLSServerName,
		AuthMode: record.Request.AuthMode, AuthUsername: record.Request.AuthUsername,
		PasswordSource: record.Request.AuthMode == AuthModePlain,
	}
	if err := transport.ValidateStatic(); err != nil {
		return fmt.Errorf("invalid request transport: %w", err)
	}
	if record.Result.Success && record.Result.Error != "" {
		return errors.New("invalid successful result with error")
	}
	if !record.Result.Success && record.Result.Error == "" {
		return errors.New("invalid failed result without error")
	}
	return nil
}

// RecentHistory returns a newest-first copy limited to count records.
func RecentHistory(records []HistoryRecord, count int) []HistoryRecord {
	if count <= 0 {
		return []HistoryRecord{}
	}
	ordered := append([]HistoryRecord(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].CreatedAt.After(ordered[j].CreatedAt)
	})
	if len(ordered) > count {
		ordered = ordered[:count]
	}
	return ordered
}

// FindHistory returns the record matching id.
func FindHistory(records []HistoryRecord, id string) (HistoryRecord, bool) {
	for _, record := range records {
		if record.ID == id {
			return record, true
		}
	}
	return HistoryRecord{}, false
}

// SummarizeHistory removes Body, BCC and attachment paths from list output.
func SummarizeHistory(record HistoryRecord) HistorySummary {
	return HistorySummary{
		ID:        record.ID,
		CreatedAt: record.CreatedAt,
		Server:    record.Request.Server,
		Port:      record.Request.Port,
		From:      record.Request.From,
		To:        append([]string(nil), record.Request.To...),
		Subject:   record.Request.Subject,
		Result:    record.Result,
	}
}

// ReplayOptions reconstructs structured values without reusing authorization controls.
func ReplayOptions(record HistoryRecord) StructuredSendOptions {
	return StructuredSendOptions{
		Server:      record.Request.Server,
		Port:        record.Request.Port,
		From:        record.Request.From,
		To:          append([]string(nil), record.Request.To...),
		CC:          append([]string(nil), record.Request.CC...),
		BCC:         append([]string(nil), record.Request.BCC...),
		Subject:     record.Request.Subject,
		Body:        record.Request.Body,
		Attachments: append([]string(nil), record.Request.Attachments...),
		TLSMode:     record.Request.TLSMode, TLSServerName: record.Request.TLSServerName,
		AuthMode: record.Request.AuthMode, AuthUsername: record.Request.AuthUsername,
	}
}

// ResolveHistoryPath returns the XDG state path or its home-directory fallback.
func ResolveHistoryPath(xdgStateHome, home string) (string, error) {
	if xdgStateHome != "" {
		return filepath.Join(xdgStateHome, "hermes", "history.jsonl"), nil
	}
	if home == "" {
		return "", errors.New("cannot resolve history path without XDG_STATE_HOME or home directory")
	}
	return filepath.Join(home, ".local", "state", "hermes", "history.jsonl"), nil
}

// DefaultHistoryPath resolves the history path from the current process environment.
func DefaultHistoryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil && os.Getenv("XDG_STATE_HOME") == "" {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return ResolveHistoryPath(os.Getenv("XDG_STATE_HOME"), home)
}

// AppendHistory appends one validated record as a single JSON line.
func AppendHistory(path string, record HistoryRecord) error {
	if err := ValidateHistoryRecords([]HistoryRecord{record}); err != nil {
		return err
	}
	line, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode history record: %w", err)
	}
	line = append(line, '\n')

	historyAppendMu.Lock()
	defer historyAppendMu.Unlock()

	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create history directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open history file: %w", err)
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return fmt.Errorf("secure history file: %w", err)
	}
	if _, err := file.Write(line); err != nil {
		_ = file.Close()
		return fmt.Errorf("append history record: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close history file: %w", err)
	}
	return nil
}

// ReadHistory reads and validates every JSONL record, failing on partial data.
func ReadHistory(path string) ([]HistoryRecord, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return []HistoryRecord{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open history file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	records := make([]HistoryRecord, 0)
	seen := make(map[string]struct{})
	for lineNumber := 1; ; lineNumber++ {
		line, readErr := reader.ReadBytes('\n')
		if len(line) == 0 && errors.Is(readErr, io.EOF) {
			break
		}
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			return nil, fmt.Errorf("history line %d is empty", lineNumber)
		}
		var record HistoryRecord
		if err := json.Unmarshal(trimmed, &record); err != nil {
			return nil, fmt.Errorf("history line %d: decode record: %w", lineNumber, err)
		}
		if err := validateHistoryRecord(record); err != nil {
			return nil, fmt.Errorf("history line %d: %w", lineNumber, err)
		}
		if _, exists := seen[record.ID]; exists {
			return nil, fmt.Errorf("history line %d: duplicate history id %q", lineNumber, record.ID)
		}
		seen[record.ID] = struct{}{}
		records = append(records, record)
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read history line %d: %w", lineNumber, readErr)
		}
	}
	return records, nil
}
