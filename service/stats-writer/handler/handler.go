package handler

import (
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-writer/usecase"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/idempotent_consumer"
	"go.uber.org/zap"
)

// statEnvelope mirrors the outbox envelope for dedup.
type statEnvelope struct {
	EventID string          `json:"event_id"`
	Payload json.RawMessage `json:"payload"`
}

type StatsHandler struct {
	useCase usecase.UseCase
	log     logger.LoggerInterface
	dedup   *idempotent_consumer.Dedup
}

func NewStatsHandler(useCase usecase.UseCase, log logger.LoggerInterface) *StatsHandler {
	return &StatsHandler{
		useCase: useCase,
		log:     log,
		dedup:   idempotent_consumer.New(48 * time.Hour),
	}
}

func (h *StatsHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *StatsHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return h.useCase.Close()
}

func (h *StatsHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// Unwrap Phase 3 envelope and check dedup before processing.
		raw := msg.Value
		if env := h.tryUnwrap(raw); env != nil {
			if h.dedup.IsDuplicate(env.EventID) {
				session.MarkMessage(msg, "")
				continue
			}
			raw = env.Payload
		}

		switch msg.Topic {
		case "payment.transaction.created", "stats-topic-transaction-events":
			var event events.TransactionEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				h.log.Error("Failed to unmarshal transaction event", zap.Error(err))
				continue
			}
			if err := h.useCase.SaveTransactionEvent(session.Context(), event); err != nil {
				h.log.Error("Failed to save transaction event", zap.Error(err))
				continue
			}
		case "stats-topic-topup-events":
			var event events.TopupEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				h.log.Error("Failed to unmarshal topup event", zap.Error(err))
				continue
			}
			if err := h.useCase.SaveTopupEvent(session.Context(), event); err != nil {
				h.log.Error("Failed to save topup event", zap.Error(err))
				continue
			}
		case "stats-topic-transfer-events":
			var event events.TransferEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				h.log.Error("Failed to unmarshal transfer event", zap.Error(err))
				continue
			}
			if err := h.useCase.SaveTransferEvent(session.Context(), event); err != nil {
				h.log.Error("Failed to save transfer event", zap.Error(err))
				continue
			}
		case "stats-topic-withdraw-events":
			var event events.WithdrawEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				h.log.Error("Failed to unmarshal withdraw event", zap.Error(err))
				continue
			}
			if err := h.useCase.SaveWithdrawEvent(session.Context(), event); err != nil {
				h.log.Error("Failed to save withdraw event", zap.Error(err))
				continue
			}
		case "stats-topic-saldo-events":
			var event events.SaldoEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				h.log.Error("Failed to unmarshal saldo event", zap.Error(err))
				continue
			}
			if err := h.useCase.SaveSaldoEvent(session.Context(), event); err != nil {
				h.log.Error("Failed to save saldo event", zap.Error(err))
				continue
			}
		case "stats-topic-merchant-events":
			var event events.MerchantEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				h.log.Error("Failed to unmarshal merchant event", zap.Error(err))
				continue
			}
			if err := h.useCase.SaveMerchantEvent(session.Context(), event); err != nil {
				h.log.Error("Failed to save merchant event", zap.Error(err))
				continue
			}
		case "stats-topic-card-events":
			var event events.CardEvent
			if err := json.Unmarshal(raw, &event); err != nil {
				h.log.Error("Failed to unmarshal card event", zap.Error(err))
				continue
			}
			if err := h.useCase.SaveCardEvent(session.Context(), event); err != nil {
				h.log.Error("Failed to save card event", zap.Error(err))
				continue
			}
		}

		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *StatsHandler) tryUnwrap(raw []byte) *statEnvelope {
	var env statEnvelope
	if err := json.Unmarshal(raw, &env); err != nil || env.EventID == "" {
		return nil
	}
	return &env
}
