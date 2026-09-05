package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Publisher polls the outbox_events table and publishes pending
// events to Kafka, then marks them as published or failed.
type Publisher struct {
	db      *gorm.DB
	kafka   *kafka.Kafka
	logger  logger.LoggerInterface

	// PollInterval controls how often the publisher checks for new events.
	PollInterval time.Duration
	// BatchSize is the max number of rows claimed per poll cycle.
	BatchSize int32
	// MaxAttempts is the maximum number of retries before an event is abandoned.
	MaxAttempts int32

	publishedCounter metric.Int64Counter
	failedCounter    metric.Int64Counter
	pendingGauge     metric.Int64ObservableGauge
}

// NewPublisher creates a new outbox publisher.
func NewPublisher(db *gorm.DB, k *kafka.Kafka, log logger.LoggerInterface) *Publisher {
	meter := otel.Meter("shared/outbox")
	publishedCounter, err := meter.Int64Counter(
		"outbox_events_published",
		metric.WithDescription("Total number of outbox events published to Kafka"),
		metric.WithUnit("1"),
	)
	if err != nil {
		log.Error("outbox: failed to create published counter", zap.Error(err))
	}
	failedCounter, err := meter.Int64Counter(
		"outbox_events_failed",
		metric.WithDescription("Total number of outbox events that failed to publish"),
		metric.WithUnit("1"),
	)
	if err != nil {
		log.Error("outbox: failed to create failed counter", zap.Error(err))
	}

	pendingGauge, err := meter.Int64ObservableGauge(
		"outbox_events_pending",
		metric.WithDescription("Number of outbox events awaiting delivery (pending or retriable failed)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		log.Error("outbox: failed to create pending gauge", zap.Error(err))
	}

	p := &Publisher{
		db:               db,
		kafka:            k,
		logger:           log,
		PollInterval:     5 * time.Second,
		BatchSize:        50,
		MaxAttempts:      10,
		publishedCounter: publishedCounter,
		failedCounter:    failedCounter,
		pendingGauge:     pendingGauge,
	}

	if pendingGauge != nil && p.db != nil {
		if _, err := meter.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
			count, err := CountPendingOutbox(ctx, p.db)
			if err != nil {
				p.logger.Error("outbox: failed to read pending count", zap.Error(err))
				return nil
			}
			obs.ObserveInt64(pendingGauge, count)
			return nil
		}, pendingGauge); err != nil {
			log.Error("outbox: failed to register pending gauge callback", zap.Error(err))
		}
	}

	return p
}

// Start begins the publisher loop. It blocks until ctx is cancelled.
func (p *Publisher) Start(ctx context.Context) {
	p.logger.Info("outbox publisher started",
		zap.Duration("poll_interval", p.PollInterval),
		zap.Int32("batch_size", p.BatchSize),
		zap.Int32("max_attempts", p.MaxAttempts),
	)

	ticker := time.NewTicker(p.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("outbox publisher shutting down")
			return
		case <-ticker.C:
			p.pollAndPublish(ctx)
		}
	}
}

func (p *Publisher) pollAndPublish(ctx context.Context) {
	records, err := ClaimPendingOutbox(ctx, p.db, p.BatchSize, p.MaxAttempts)
	if err != nil {
		p.logger.Error("outbox: failed to claim pending events", zap.Error(err))
		return
	}

	if len(records) == 0 {
		return
	}

	p.logger.Info("outbox: processing batch", zap.Int("count", len(records)))

	for _, record := range records {
		topic := p.resolveTopic(record.EventType)
		if topic == "" {
			p.logger.Warn("outbox: unknown event type, skipping",
				zap.String("event_type", record.EventType),
				zap.String("event_id", record.EventID),
			)
			_ = CompleteOutbox(ctx, p.db, record.EventID)
			continue
		}

		key := record.AggregateID
		if key == "" {
			key = record.EventID
		}

		if err := p.kafka.SendMessage(topic, key, record.Payload); err != nil {
			p.logger.Error("outbox: failed to publish event",
				zap.String("event_id", record.EventID),
				zap.String("topic", topic),
				zap.Error(err),
			)
			if p.failedCounter != nil {
				p.failedCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topic)))
			}
			if failErr := FailOutbox(ctx, p.db, record.EventID, fmt.Sprintf("%.256s", err.Error())); failErr != nil {
				p.logger.Error("outbox: failed to mark event as failed",
					zap.String("event_id", record.EventID),
					zap.Error(failErr),
				)
			}
			continue
		}

		if err := CompleteOutbox(ctx, p.db, record.EventID); err != nil {
			p.logger.Error("outbox: failed to mark event as published",
				zap.String("event_id", record.EventID),
				zap.Error(err),
			)
		} else if p.publishedCounter != nil {
			p.publishedCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topic)))
		}

		p.logger.Debug("outbox: event published",
			zap.String("event_id", record.EventID),
			zap.String("topic", topic),
			zap.String("event_type", record.EventType),
		)
	}
}

// resolveTopic maps event types to Kafka topics.
func (p *Publisher) resolveTopic(eventType string) string {
	switch eventType {
	case "topup.created":
		return "email-service-topic-topup-create"
	case "topup.stats":
		return "stats-topic-topup-events"
	case "topup.saldo":
		return "stats-topic-saldo-events"

	case "transaction.created":
		return "email-service-topic-transaction-create"
	case "transaction.stats":
		return "stats-topic-transaction-events"
	case "transaction.saldo_user", "transaction.saldo_merchant":
		return "stats-topic-saldo-events"

	case "transfer.created":
		return "email-service-topic-transfer-create"
	case "transfer.stats":
		return "stats-topic-transfer-events"
	case "transfer.saldo_sender", "transfer.saldo_receiver":
		return "stats-topic-saldo-events"

	case "withdraw.created":
		return "email-service-topic-withdraw-create"
	case "withdraw.stats":
		return "stats-topic-withdraw-events"
	case "withdraw.saldo":
		return "stats-topic-saldo-events"

	case "saldo.changed":
		return "stats-topic-saldo-events"

	default:
		return ""
	}
}
