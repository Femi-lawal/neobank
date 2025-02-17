package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/femi-lawal/new_bank/backend/ledger-service/internal/service"
	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/kafka"
)

// PaymentConsumer consumes payment events from Kafka
type PaymentConsumer struct {
	consumer  *kafka.Consumer
	ledgerSvc *service.LedgerService
	producer  *kafka.Producer // For publishing completion events
}

// NewPaymentConsumer creates a new payment event consumer
func NewPaymentConsumer(brokers []string, ledgerSvc *service.LedgerService, producer *kafka.Producer) *PaymentConsumer {
	consumer := kafka.NewConsumer(brokers, "ledger-service", kafka.TopicPaymentCreated)
	return &PaymentConsumer{
		consumer:  consumer,
		ledgerSvc: ledgerSvc,
		producer:  producer,
	}
}

// Start begins consuming payment events
func (c *PaymentConsumer) Start(ctx context.Context) error {
	slog.Info("Starting payment event consumer", "topic", kafka.TopicPaymentCreated)

	return c.consumer.Consume(ctx, func(key string, value []byte) error {
		var event kafka.PaymentEvent
		if err := json.Unmarshal(value, &event); err != nil {
			slog.Error("Failed to unmarshal payment event", "error", err)
			return err
		}

		slog.Info("Processing payment event", "payment_id", event.PaymentID, "amount", event.Amount)

		// Process the transfer
		err := c.processPayment(ctx, event)
		if err != nil {
			slog.Error("Failed to process payment", "payment_id", event.PaymentID, "error", err)
			// Publish failure event
			c.publishResult(ctx, event.PaymentID, kafka.TopicPaymentFailed, event)
			return nil // Don't retry, just log
		}

		// Publish success event
		event.Status = "COMPLETED"
		c.publishResult(ctx, event.PaymentID, kafka.TopicPaymentCompleted, event)

		slog.Info("Payment processed successfully", "payment_id", event.PaymentID)
		return nil
	})
}

// processPayment executes the ledger transaction
func (c *PaymentConsumer) processPayment(ctx context.Context, event kafka.PaymentEvent) error {
	// Create journal entry with postings using the convenience method
	_, err := c.ledgerSvc.PostTransfer(
		event.FromAccountID,
		event.ToAccountID,
		event.Amount,
		"Payment: "+event.Description,
	)
	return err
}

// publishResult publishes the payment result event
func (c *PaymentConsumer) publishResult(ctx context.Context, paymentID, topic string, event kafka.PaymentEvent) {
	if c.producer == nil {
		return
	}

	if err := c.producer.Produce(ctx, topic, paymentID, event); err != nil {
		slog.Error("Failed to publish payment result", "payment_id", paymentID, "topic", topic, "error", err)
	}
}

// Close closes the consumer
func (c *PaymentConsumer) Close() error {
	return c.consumer.Close()
}
