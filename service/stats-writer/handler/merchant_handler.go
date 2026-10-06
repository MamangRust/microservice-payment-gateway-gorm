package handler

import (
	"context"
	"encoding/json"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
	"go.uber.org/zap"
)

func (h *StatsHandler) handleMerchant(ctx context.Context, raw []byte) error {
	var event events.MerchantEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		h.log.Error("Failed to unmarshal merchant event", zap.Error(err))
		return err
	}
	if err := h.useCase.SaveMerchantEvent(ctx, event); err != nil {
		h.log.Error("Failed to save merchant event", zap.Error(err))
		return err
	}
	return nil
}
