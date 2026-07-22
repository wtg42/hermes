package sendmail

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestStructuredSenderAttemptReceivesCompletePlan(t *testing.T) {
	var got StructuredSendPlan
	calls := 0
	sender := &StructuredSender{
		attempt: func(plan StructuredSendPlan) error { calls++; got = plan; return nil },
		now:     func() time.Time { return time.Date(2026, 7, 22, 1, 2, 3, 0, time.UTC) },
		entropy: deterministicEntropy(),
	}
	options := historySafeOptions()
	options.TLSMode = TLSModeRequired
	options.TLSServerName = "smtp.test"
	options.AuthMode = AuthModePlain
	options.AuthUsername = "agent"
	options.AuthPasswordStdin = true
	options.AuthPassword = "one-shot-secret"
	if err := sender.Send(options); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || got.Transport.TLSMode != TLSModeRequired || got.Transport.AuthUsername != "agent" || got.Transport.Password != "one-shot-secret" {
		t.Fatalf("calls=%d plan=%+v", calls, got)
	}
	if got.Compose.Subject != options.Subject || got.Compose.Host != options.Server {
		t.Fatalf("compose=%+v", got.Compose)
	}
}

func TestStructuredPlanCopiesInputSlices(t *testing.T) {
	to := []string{"adam@rd01.softnext.com.tw"}
	attachments := []string{}
	sender := &StructuredSender{now: time.Now, entropy: deterministicEntropy()}
	plan, err := sender.BuildPlan(StructuredSendOptions{Server: "192.0.2.10", To: to, Attachments: attachments})
	if err != nil {
		t.Fatal(err)
	}
	to[0] = "changed@example.com"
	if plan.Compose.To[0] == to[0] {
		t.Fatal("plan aliases input recipient slice")
	}
}

func TestStructuredAttemptErrorUsesSamePlanOnce(t *testing.T) {
	want := errors.New("attempt failed")
	calls := 0
	var attempted StructuredSendPlan
	sender := NewStructuredSender(func(plan StructuredSendPlan) error { calls++; attempted = plan; return want })
	execution, err := sender.Execute(historySafeOptions())
	if !errors.Is(err, want) || calls != 1 || !execution.MailerAttempted || !reflect.DeepEqual(execution.Plan, attempted) {
		t.Fatalf("execution=%+v error=%v calls=%d", execution, err, calls)
	}
}

func deterministicEntropy() *repeatReader { return &repeatReader{} }

type repeatReader struct{ offset int }

func (reader *repeatReader) Read(buffer []byte) (int, error) {
	pattern := []byte{1, 2, 3, 4, 0xde, 0xad, 0xbe, 0xef}
	for index := range buffer {
		buffer[index] = pattern[reader.offset%len(pattern)]
		reader.offset++
	}
	return len(buffer), nil
}
