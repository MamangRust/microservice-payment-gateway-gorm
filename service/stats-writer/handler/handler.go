package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-writer/usecase"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/idempotent_consumer"
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

		// Pesan yang gagal diproses sengaja TIDAK di-MarkMessage, supaya Kafka
		// mengirim ulang (at-least-once) alih-alih event hilang diam-diam.
		if err := h.dispatch(session.Context(), msg.Topic, raw); err != nil {
			continue
		}

		session.MarkMessage(msg, "")
	}
	return nil
}

// dispatch mengarahkan satu pesan Kafka ke handler pemilik topic-nya dan
// mengembalikan error pemrosesan. Tiap handler tinggal di file domainnya sendiri
// (transaction_handler.go, topup_handler.go, dst.) supaya penambahan event baru
// tidak menumpuk di satu switch raksasa. Topic yang tidak dikenal bukan error.
func (h *StatsHandler) dispatch(ctx context.Context, topic string, raw []byte) error {
	switch topic {
	case "payment.transaction.created", "stats-topic-transaction-events":
		return h.handleTransaction(ctx, raw)
	case "stats-topic-topup-events":
		return h.handleTopup(ctx, raw)
	case "stats-topic-transfer-events":
		return h.handleTransfer(ctx, raw)
	case "stats-topic-withdraw-events":
		return h.handleWithdraw(ctx, raw)
	case "stats-topic-saldo-events":
		return h.handleSaldo(ctx, raw)
	case "stats-topic-merchant-events":
		return h.handleMerchant(ctx, raw)
	case "stats-topic-card-events":
		return h.handleCard(ctx, raw)
	default:
		return nil
	}
}

func (h *StatsHandler) tryUnwrap(raw []byte) *statEnvelope {
	var env statEnvelope
	if err := json.Unmarshal(raw, &env); err != nil || env.EventID == "" {
		return nil
	}
	return &env
}
