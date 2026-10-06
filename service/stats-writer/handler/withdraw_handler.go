package handler

import (
	"context"
	"encoding/json"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
	"go.uber.org/zap"
)

func (h *StatsHandler) handleWithdraw(ctx context.Context, raw []byte) error {
	var event events.WithdrawEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		h.log.Error("Failed to unmarshal withdraw event", zap.Error(err))
		return err
	}
	if err := h.useCase.SaveWithdrawEvent(ctx, event); err != nil {
		h.log.Error("Failed to save withdraw event", zap.Error(err))
		return err
	}
	return nil
}
