//go:build ignore

// REFERENCE ONLY — real Kafka (segmentio/kafka-go) + RabbitMQ (amqp091-go)
// producer/consumer code. Build-tagged `ignore` so this folder builds
// without the broker client deps.
//
// To run for real:
//   docker compose up -d   (see docker-compose.yml: starts Kafka + RabbitMQ)
//   remove the build tag, then:
//   go get github.com/segmentio/kafka-go github.com/rabbitmq/amqp091-go
//   go run .
package main

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/segmentio/kafka-go"
)

// ---- Kafka producer ------------------------------------------------------
func kafkaProduce(ctx context.Context, broker, topic string) error {
	w := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.Hash{}, // key-based partitioning -> per-key ordering
	}
	defer w.Close()
	return w.WriteMessages(ctx, kafka.Message{
		Key:   []byte("order-42"), // key decides partition (ordering per order)
		Value: []byte(`{"event_id":"evt-1","order_id":"42","qty":3}`),
	})
}

// ---- Kafka consumer (consumer group, at-least-once) ----------------------
func kafkaConsume(ctx context.Context, broker, topic, group string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: group, // consumer group: partitions shared across members
	})
	defer r.Close()
	for {
		// FetchMessage does NOT auto-commit -> we commit AFTER processing
		// (at-least-once: a crash before commit => redelivery => need idempotency)
		m, err := r.FetchMessage(ctx)
		if err != nil {
			return
		}
		if err := process(m.Value); err != nil {
			log.Printf("process failed, will be redelivered: %v", err)
			continue // do NOT commit -> redelivered
		}
		if err := r.CommitMessages(ctx, m); err != nil {
			log.Printf("commit failed: %v", err)
		}
	}
}

func process(payload []byte) error {
	// idempotency guard goes here (dedup on event_id) — see lesson
	log.Printf("processed: %s", payload)
	return nil
}

// ---- RabbitMQ producer + consumer ---------------------------------------
func rabbitDemo(url string) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("orders", true /*durable*/, false, false, false, nil)
	if err != nil {
		return err
	}

	// publish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ch.PublishWithContext(ctx, "", q.Name, false, false,
		amqp.Publishing{ContentType: "application/json", Body: []byte(`{"order_id":"42"}`)}); err != nil {
		return err
	}

	// consume (manual ack = at-least-once)
	msgs, err := ch.Consume(q.Name, "", false /*autoAck=false*/, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range msgs {
			if err := process(d.Body); err != nil {
				d.Nack(false, true) // requeue on failure
				continue
			}
			d.Ack(false) // ack AFTER successful processing
		}
	}()
	return nil
}
