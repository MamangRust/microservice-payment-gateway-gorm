package handler

import (
	"context"
	"encoding/json"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
	"go.uber.org/zap"
)

func (h *StatsHandler) handleTransaction(ctx context.Context, raw []byte) error {
	var event events.TransactionEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		h.log.Error("Failed to unmarshal transaction event", zap.Error(err))
		return err
	}
	if err := h.useCase.SaveTransactionEvent(ctx, event); err != nil {
		h.log.Error("Failed to save transaction event", zap.Error(err))
		return err
	}
	return nil
}
