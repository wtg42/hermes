package sendmail

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wtg42/hermes/mail"
)

func TestNewHistoryRecordShape(t *testing.T) {
	now := time.Date(2026, 7, 22, 10, 30, 0, 0, time.FixedZone("Asia/Taipei", 8*60*60))
	compose := historyTestCompose()

	tests := []struct {
		name    string
		sendErr error
		success bool
		message string
	}{
		{name: "success", success: true},
		{name: "failure", sendErr: errors.New("smtp unavailable"), message: "smtp unavailable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := StructuredSendPlan{Compose: compose, Transport: DefaultSMTPTransportConfig(compose.Host, compose.Port)}
			record, err := NewHistoryRecord(plan, tt.sendErr, now, strings.NewReader("12345678"))
			if err != nil {
				t.Fatalf("NewHistoryRecord() error = %v", err)
			}
			if record.Version != HistoryRecordVersion || record.ID == "" || record.CreatedAt != now.UTC() {
				t.Fatalf("record identity = %+v", record)
			}
			if record.Request.Server != compose.Host || record.Request.Port != compose.Port || record.Request.From != compose.From {
				t.Fatalf("record request = %+v", record.Request)
			}
			if record.Result.Success != tt.success || record.Result.Error != tt.message {
				t.Fatalf("record result = %+v", record.Result)
			}

			encoded, err := json.Marshal(record)
			if err != nil {
				t.Fatal(err)
			}
			for _, forbidden := range []string{"confirm_outside_whitelist", "no_history", "password", "token"} {
				if strings.Contains(string(encoded), forbidden) {
					t.Fatalf("record contains forbidden field %q: %s", forbidden, encoded)
				}
			}
		})
	}
}

func TestHistoryPersistsTransportPolicyWithoutSecret(t *testing.T) {
	transport := SMTPTransportConfig{
		Server: "192.0.2.10", Port: "587", TLSMode: TLSModeRequired,
		TLSServerName: "smtp.test", AuthMode: AuthModePlain,
		AuthUsername: "agent", PasswordSource: true, Password: "one-shot-secret",
	}
	record, err := NewHistoryRecord(StructuredSendPlan{Compose: historyTestCompose(), Transport: transport}, errors.New("auth failed for one-shot-secret"), time.Now(), strings.NewReader("12345678"))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	if strings.Contains(text, transport.Password) || !strings.Contains(text, "[REDACTED]") {
		t.Fatalf("history secret handling = %s", text)
	}
	if record.Request.TLSMode != TLSModeRequired || record.Request.AuthMode != AuthModePlain || record.Request.AuthUsername != "agent" {
		t.Fatalf("transport metadata = %+v", record.Request)
	}
	options := ReplayOptions(record)
	if options.AuthPassword != "" || options.AuthPasswordStdin || options.AuthUsername != "agent" {
		t.Fatalf("replay options = %+v", options)
	}
}

func TestHistoryDataOperations(t *testing.T) {
	older := historyTestRecord("older", time.Date(2026, 7, 22, 1, 0, 0, 0, time.UTC))
	newer := historyTestRecord("newer", time.Date(2026, 7, 22, 2, 0, 0, 0, time.UTC))
	recent := RecentHistory([]HistoryRecord{older, newer}, 1)
	if len(recent) != 1 || recent[0].ID != "newer" {
		t.Fatalf("RecentHistory() = %+v", recent)
	}
	if record, ok := FindHistory([]HistoryRecord{older, newer}, "older"); !ok || record.ID != "older" {
		t.Fatalf("FindHistory() = %+v, %v", record, ok)
	}
	if _, ok := FindHistory([]HistoryRecord{older}, "missing"); ok {
		t.Fatal("FindHistory() found missing record")
	}

	summary := SummarizeHistory(newer)
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	for _, hidden := range []string{"body", "bcc", "attachments"} {
		if strings.Contains(string(encoded), `"`+hidden+`"`) {
			t.Fatalf("summary leaks %s: %s", hidden, encoded)
		}
	}

	options := ReplayOptions(newer)
	if options.Server != newer.Request.Server || options.Body != newer.Request.Body || options.ConfirmOutsideWhitelist || options.NoHistory {
		t.Fatalf("ReplayOptions() = %+v", options)
	}
}

func TestValidateHistoryRecords(t *testing.T) {
	valid := historyTestRecord("valid", time.Now().UTC())
	tests := []struct {
		name    string
		records []HistoryRecord
		want    string
	}{
		{name: "valid", records: []HistoryRecord{valid}},
		{name: "empty id", records: []HistoryRecord{func() HistoryRecord { r := valid; r.ID = ""; return r }()}, want: "id"},
		{name: "unsupported version", records: []HistoryRecord{func() HistoryRecord { r := valid; r.Version = 99; return r }()}, want: "version"},
		{name: "missing server", records: []HistoryRecord{func() HistoryRecord { r := valid; r.Request.Server = ""; return r }()}, want: "server"},
		{name: "missing recipient", records: []HistoryRecord{func() HistoryRecord { r := valid; r.Request.To = nil; return r }()}, want: "to"},
		{name: "invalid success result", records: []HistoryRecord{func() HistoryRecord { r := valid; r.Result.Error = "unexpected"; return r }()}, want: "result"},
		{name: "invalid failure result", records: []HistoryRecord{func() HistoryRecord { r := valid; r.Result.Success = false; r.Result.Error = ""; return r }()}, want: "result"},
		{name: "duplicate id", records: []HistoryRecord{valid, valid}, want: "duplicate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHistoryRecords(tt.records)
			if tt.want == "" && err != nil {
				t.Fatalf("ValidateHistoryRecords() error = %v", err)
			}
			if tt.want != "" && (err == nil || !strings.Contains(strings.ToLower(err.Error()), tt.want)) {
				t.Fatalf("ValidateHistoryRecords() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestResolveHistoryPath(t *testing.T) {
	tests := []struct {
		name string
		xdg  string
		home string
		want string
		err  bool
	}{
		{name: "xdg", xdg: "/state", home: "/home/test", want: "/state/hermes/history.jsonl"},
		{name: "home fallback", home: "/home/test", want: "/home/test/.local/state/hermes/history.jsonl"},
		{name: "missing home", err: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveHistoryPath(tt.xdg, tt.home)
			if (err != nil) != tt.err || got != tt.want {
				t.Fatalf("ResolveHistoryPath() = %q, %v; want %q, error %v", got, err, tt.want, tt.err)
			}
		})
	}
}

func TestAppendAndReadHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "hermes", "history.jsonl")
	records := []HistoryRecord{
		historyTestRecord("one", time.Date(2026, 7, 22, 1, 0, 0, 0, time.UTC)),
		historyTestRecord("two", time.Date(2026, 7, 22, 2, 0, 0, 0, time.UTC)),
	}
	records[0].Request.Body = "中文\nEnglish"
	for _, record := range records {
		if err := AppendHistory(path, record); err != nil {
			t.Fatalf("AppendHistory() error = %v", err)
		}
	}

	got, err := ReadHistory(path)
	if err != nil {
		t.Fatalf("ReadHistory() error = %v", err)
	}
	if len(got) != 2 || got[0].ID != "one" || got[1].ID != "two" || got[0].Request.Body != records[0].Request.Body {
		t.Fatalf("ReadHistory() = %+v", got)
	}

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if dirInfo.Mode().Perm() != 0o700 || fileInfo.Mode().Perm() != 0o600 {
		t.Fatalf("permissions = dir %o file %o", dirInfo.Mode().Perm(), fileInfo.Mode().Perm())
	}
}

func TestReadHistoryMissingAndEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.jsonl")
	for _, setup := range []func() error{
		func() error { return nil },
		func() error { return os.WriteFile(path, nil, 0o600) },
	} {
		if err := setup(); err != nil {
			t.Fatal(err)
		}
		records, err := ReadHistory(path)
		if err != nil || len(records) != 0 {
			t.Fatalf("ReadHistory() = %+v, %v", records, err)
		}
	}
}

func TestAppendHistoryConcurrent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	const count = 20
	var wg sync.WaitGroup
	errs := make(chan error, count)
	for i := range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			record := historyTestRecord(string(rune('a'+i)), time.Now().UTC())
			errs <- AppendHistory(path, record)
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("AppendHistory() error = %v", err)
		}
	}
	records, err := ReadHistory(path)
	if err != nil || len(records) != count {
		t.Fatalf("ReadHistory() count = %d, error = %v", len(records), err)
	}
}

func TestHistoryIOFailsClosed(t *testing.T) {
	dir := t.TempDir()
	blockedParent := filepath.Join(dir, "file")
	if err := os.WriteFile(blockedParent, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := AppendHistory(filepath.Join(blockedParent, "history.jsonl"), historyTestRecord("one", time.Now().UTC())); err == nil {
		t.Fatal("AppendHistory() accepted unwritable path")
	}

	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "truncated", content: `{"version":1`, want: "line 1"},
		{name: "unsupported", content: `{"version":99,"id":"one","created_at":"2026-07-22T00:00:00Z","request":{},"result":{}}` + "\n", want: "version"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, tt.name+".jsonl")
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := ReadHistory(path); err == nil || !strings.Contains(strings.ToLower(err.Error()), tt.want) {
				t.Fatalf("ReadHistory() error = %v, want %q", err, tt.want)
			}
		})
	}

	duplicatePath := filepath.Join(dir, "duplicate.jsonl")
	data, _ := json.Marshal(historyTestRecord("same", time.Now().UTC()))
	if err := os.WriteFile(duplicatePath, append(append(data, '\n'), append(data, '\n')...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadHistory(duplicatePath); err == nil || !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		t.Fatalf("ReadHistory() duplicate error = %v", err)
	}
}

func historyTestCompose() mail.MailCompose {
	return mail.MailCompose{
		From:        "weitingshih@rd01.softnext.com.tw",
		To:          []string{"adam@rd01.softnext.com.tw"},
		CC:          []string{"jllee@rd01.softnext.com.tw"},
		BCC:         []string{"audit@rd01.softnext.com.tw"},
		Subject:     "History 中文 📨",
		Body:        "body",
		Attachments: []string{"/tmp/one.txt", "/tmp/two.txt"},
		Host:        "192.0.2.10",
		Port:        "25",
	}
}

func historyTestRecord(id string, createdAt time.Time) HistoryRecord {
	compose := historyTestCompose()
	return HistoryRecord{
		Version:   HistoryRecordVersion,
		ID:        id,
		CreatedAt: createdAt,
		Request:   RecordedSendRequestFromCompose(compose),
		Result:    RecordedSendResult{Success: true},
	}
}
