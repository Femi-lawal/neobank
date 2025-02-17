package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer wraps kafka-go writer for producing messages
type Producer struct {
	writer *kafka.Writer
}

// Consumer wraps kafka-go reader for consuming messages
type Consumer struct {
	reader *kafka.Reader
}

// PaymentEvent represents a payment event message
type PaymentEvent struct {
	PaymentID     string `json:"payment_id"`
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	Timestamp     string `json:"timestamp"`
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}
	slog.Info("Kafka producer initialized", "brokers", brokers)
	return &Producer{writer: writer}
}

// Produce sends a message to the specified topic
func (p *Producer) Produce(ctx context.Context, topic string, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		slog.Error("Failed to produce message", "topic", topic, "error", err)
		return err
	}

	slog.Info("Message produced", "topic", topic, "key", key)
	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, groupID, topic string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})
	slog.Info("Kafka consumer initialized", "brokers", brokers, "group", groupID, "topic", topic)
	return &Consumer{reader: reader}
}

// Consume reads messages and calls the handler for each
func (c *Consumer) Consume(ctx context.Context, handler func(key string, value []byte) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				slog.Error("Failed to read message", "error", err)
				continue
			}

			if err := handler(string(msg.Key), msg.Value); err != nil {
				slog.Error("Failed to handle message", "key", string(msg.Key), "error", err)
				// Continue processing other messages
			}
		}
	}
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// Topics for payment events
const (
	TopicPaymentCreated   = "payment.created"
	TopicPaymentCompleted = "payment.completed"
	TopicPaymentFailed    = "payment.failed"
)
