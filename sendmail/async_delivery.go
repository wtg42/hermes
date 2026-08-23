package sendmail

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// DeliveryJob is the complete SMTP envelope and RFC 5322 payload for one deferred delivery.
// It intentionally contains no credentials; worker SMTP authentication is a future transport policy concern.
type DeliveryJob struct {
	SMTPAddress string   `json:"smtp_address"`
	From        string   `json:"from"`
	To          []string `json:"to"`
	Message     []byte   `json:"message"`
}

// WorkerConfig configures a RabbitMQ delivery worker.
type WorkerConfig struct {
	URL         string
	Queue       string
	Concurrency int
}

// Validate rejects incomplete worker configuration before the worker connects to RabbitMQ.
func (config WorkerConfig) Validate() error {
	if strings.TrimSpace(config.URL) == "" {
		return fmt.Errorf("RabbitMQ URL is required")
	}
	if strings.TrimSpace(config.Queue) == "" {
		return fmt.Errorf("RabbitMQ queue is required")
	}
	if config.Concurrency <= 0 {
		return fmt.Errorf("worker concurrency must be greater than zero")
	}
	return nil
}

// EnqueueBurst validates a burst request and creates one delivery job at a time.
// publish is a narrow I/O boundary so callers can use RabbitMQ while tests can verify job generation.
func EnqueueBurst(ctx context.Context, options BurstOptions, publish func(context.Context, DeliveryJob) error) error {
	if publish == nil {
		return fmt.Errorf("delivery job publisher is required")
	}
	plan, err := buildBurstPlan(options)
	if err != nil {
		return err
	}

	mailPool := buildBurstMailPool(plan)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for sequence := 1; sequence <= plan.quantity; sequence++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		job, err := buildBurstDeliveryJob(plan, mailPool, r, sequence)
		if err != nil {
			return err
		}
		if err := publish(ctx, job); err != nil {
			return fmt.Errorf("enqueue burst email: %w", err)
		}
	}
	return nil
}

func buildBurstDeliveryJob(plan burstPlan, mailPool []string, r *rand.Rand, sequence int) (DeliveryJob, error) {
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
	email := new(bytes.Buffer)
	email.WriteString(buildBurstHeaders(from, to, subject, messageID, time.Now()))
	if err := buildMIMEContent(email, buildBurstBody(plan.runID, sequence, plan.bodyKB)); err != nil {
		return DeliveryJob{}, fmt.Errorf("build burst email: %w", err)
	}
	return DeliveryJob{SMTPAddress: plan.address, From: from, To: []string{to}, Message: email.Bytes()}, nil
}

// DeliverJob sends a previously queued job. Callers ACK the broker message only after it succeeds.
func DeliverJob(job DeliveryJob, deliver func(address, from string, to []string, message []byte) error) error {
	if strings.TrimSpace(job.SMTPAddress) == "" || strings.TrimSpace(job.From) == "" || len(job.To) == 0 || len(job.Message) == 0 {
		return fmt.Errorf("invalid delivery job")
	}
	if deliver == nil {
		return fmt.Errorf("SMTP deliverer is required")
	}
	if err := deliver(job.SMTPAddress, job.From, job.To, job.Message); err != nil {
		return fmt.Errorf("SMTP delivery: %w", err)
	}
	return nil
}

func deliverJobWithSMTP(job DeliveryJob) error {
	return DeliverJob(job, func(address, from string, to []string, message []byte) error {
		return SendMail(address, nil, from, to, message)
	})
}
