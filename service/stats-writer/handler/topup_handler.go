package handler

import (
	"context"
	"encoding/json"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
	"go.uber.org/zap"
)

func (h *StatsHandler) handleTopup(ctx context.Context, raw []byte) error {
	var event events.TopupEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		h.log.Error("Failed to unmarshal topup event", zap.Error(err))
		return err
	}
	if err := h.useCase.SaveTopupEvent(ctx, event); err != nil {
		h.log.Error("Failed to save topup event", zap.Error(err))
		return err
	}
	return nil
}
