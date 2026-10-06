package handler

import (
	"context"
	"encoding/json"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
	"go.uber.org/zap"
)

func (h *StatsHandler) handleCard(ctx context.Context, raw []byte) error {
	var event events.CardEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		h.log.Error("Failed to unmarshal card event", zap.Error(err))
		return err
	}
	if err := h.useCase.SaveCardEvent(ctx, event); err != nil {
		h.log.Error("Failed to save card event", zap.Error(err))
		return err
	}
	return nil
}
