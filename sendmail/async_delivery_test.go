package sendmail

import (
	"context"
	"strings"
	"testing"
)

func TestEnqueueBurstBuildsOneDeliveryJobPerMessage(t *testing.T) {
	var jobs []DeliveryJob
	err := EnqueueBurst(context.Background(), BurstOptions{
		Quantity: 3,
		Host:     "smtp.example.com",
		Port:     "25",
		From:     "sender@" + safeBurstDomain,
		To:       "recipient@" + safeBurstDomain,
	}, func(_ context.Context, job DeliveryJob) error {
		jobs = append(jobs, job)
		return nil
	})
	if err != nil {
		t.Fatalf("EnqueueBurst() error = %v", err)
	}
	if len(jobs) != 3 {
		t.Fatalf("jobs = %d, want 3", len(jobs))
	}
	for _, job := range jobs {
		if job.SMTPAddress != "smtp.example.com:25" {
			t.Errorf("SMTPAddress = %q", job.SMTPAddress)
		}
		if job.From != "sender@"+safeBurstDomain || len(job.To) != 1 || job.To[0] != "recipient@"+safeBurstDomain {
			t.Errorf("unexpected envelope: %+v", job)
		}
		if !strings.Contains(string(job.Message), "From: sender@"+safeBurstDomain+"\r\n") {
			t.Errorf("message lacks sender header: %q", job.Message)
		}
	}
}

func TestEnqueueBurstStopsOnPublishFailure(t *testing.T) {
	published := 0
	err := EnqueueBurst(context.Background(), BurstOptions{
		Quantity: 3, Host: "smtp.example.com", Port: "25",
		From: "sender@" + safeBurstDomain, To: "recipient@" + safeBurstDomain,
	}, func(_ context.Context, job DeliveryJob) error {
		published++
		return context.DeadlineExceeded
	})
	if err == nil || !strings.Contains(err.Error(), "enqueue burst email") {
		t.Fatalf("error = %v, want enqueue error", err)
	}
	if published != 1 {
		t.Fatalf("published = %d, want 1", published)
	}
}

func TestDeliverJobPassesEnvelopeAndMessageToSMTP(t *testing.T) {
	job := DeliveryJob{SMTPAddress: "smtp.example.com:25", From: "from@example.com", To: []string{"to@example.com"}, Message: []byte("body")}
	var got DeliveryJob
	err := DeliverJob(job, func(address, from string, to []string, message []byte) error {
		got = DeliveryJob{SMTPAddress: address, From: from, To: to, Message: message}
		return nil
	})
	if err != nil {
		t.Fatalf("DeliverJob() error = %v", err)
	}
	if got.SMTPAddress != job.SMTPAddress || got.From != job.From || len(got.To) != 1 || got.To[0] != job.To[0] || string(got.Message) != string(job.Message) {
		t.Errorf("got %+v, want %+v", got, job)
	}
}

func TestWorkerConfigValidation(t *testing.T) {
	for _, config := range []WorkerConfig{
		{URL: "", Queue: "hermes.delivery", Concurrency: 1},
		{URL: "amqp://guest:guest@localhost:5672/", Queue: "", Concurrency: 1},
		{URL: "amqp://guest:guest@localhost:5672/", Queue: "hermes.delivery", Concurrency: 0},
	} {
		if err := config.Validate(); err == nil {
			t.Errorf("Validate(%+v) succeeded, want error", config)
		}
	}
	if err := (WorkerConfig{URL: "amqp://guest:guest@localhost:5672/", Queue: "hermes.delivery", Concurrency: 2}).Validate(); err != nil {
		t.Fatalf("valid config error = %v", err)
	}
}
