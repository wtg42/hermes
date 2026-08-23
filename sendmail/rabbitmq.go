package sendmail

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DefaultDeliveryQueue is the durable RabbitMQ queue used when no queue is specified.
const DefaultDeliveryQueue = "hermes.delivery"

// RabbitQueueConfig identifies the RabbitMQ broker and durable queue used for delivery jobs.
type RabbitQueueConfig struct {
	URL   string
	Queue string
}

// Validate rejects an incomplete RabbitMQ configuration.
func (config RabbitQueueConfig) Validate() error {
	if strings.TrimSpace(config.URL) == "" {
		return fmt.Errorf("RabbitMQ URL is required")
	}
	if strings.TrimSpace(config.Queue) == "" {
		return fmt.Errorf("RabbitMQ queue is required")
	}
	return nil
}

// EnqueueBurstToRabbitMQ publishes a durable job for every requested burst message.
func EnqueueBurstToRabbitMQ(ctx context.Context, options BurstOptions, config RabbitQueueConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}
	connection, err := amqp.Dial(config.URL)
	if err != nil {
		return fmt.Errorf("connect to RabbitMQ: %w", err)
	}
	defer connection.Close()
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("open RabbitMQ channel: %w", err)
	}
	defer channel.Close()
	if _, err := declareDeliveryQueue(channel, config.Queue); err != nil {
		return err
	}

	return EnqueueBurst(ctx, options, func(ctx context.Context, job DeliveryJob) error {
		body, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("encode delivery job: %w", err)
		}
		return channel.PublishWithContext(ctx, "", config.Queue, false, false, amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		})
	})
}

// RunRabbitMQWorker consumes delivery jobs with manual acknowledgements. A failed SMTP attempt is
// negatively acknowledged and requeued; retry delay and dead-letter routing are added in a later phase.
func RunRabbitMQWorker(ctx context.Context, config WorkerConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}
	connection, err := amqp.Dial(config.URL)
	if err != nil {
		return fmt.Errorf("connect to RabbitMQ: %w", err)
	}
	defer connection.Close()

	var workers sync.WaitGroup
	errs := make(chan error, config.Concurrency)
	for index := range config.Concurrency {
		workers.Add(1)
		go func(index int) {
			defer workers.Done()
			if err := consumeRabbitMQWorker(ctx, connection, config.Queue, index); err != nil && ctx.Err() == nil {
				errs <- err
			}
		}(index)
	}
	workers.Wait()
	close(errs)
	if err := <-errs; err != nil {
		return err
	}
	return ctx.Err()
}

func declareDeliveryQueue(channel *amqp.Channel, queue string) (amqp.Queue, error) {
	queue = strings.TrimSpace(queue)
	declared, err := channel.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("declare RabbitMQ queue %q: %w", queue, err)
	}
	return declared, nil
}

func consumeRabbitMQWorker(ctx context.Context, connection *amqp.Connection, queue string, index int) error {
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("worker %d open RabbitMQ channel: %w", index, err)
	}
	defer channel.Close()
	if _, err := declareDeliveryQueue(channel, queue); err != nil {
		return err
	}
	if err := channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("worker %d configure prefetch: %w", index, err)
	}
	deliveries, err := channel.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("worker %d consume RabbitMQ queue: %w", index, err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("worker %d RabbitMQ delivery channel closed", index)
			}
			var job DeliveryJob
			if err := json.Unmarshal(delivery.Body, &job); err != nil {
				if nackErr := delivery.Nack(false, false); nackErr != nil {
					return fmt.Errorf("worker %d reject malformed job: %w", index, nackErr)
				}
				continue
			}
			if err := deliverJobWithSMTP(job); err != nil {
				if nackErr := delivery.Nack(false, true); nackErr != nil {
					return fmt.Errorf("worker %d requeue failed delivery: %w", index, nackErr)
				}
				continue
			}
			if err := delivery.Ack(false); err != nil {
				return fmt.Errorf("worker %d acknowledge delivery: %w", index, err)
			}
		}
	}
}
