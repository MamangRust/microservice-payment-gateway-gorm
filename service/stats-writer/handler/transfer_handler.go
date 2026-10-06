package handler

import (
	"context"
	"encoding/json"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
	"go.uber.org/zap"
)

func (h *StatsHandler) handleTransfer(ctx context.Context, raw []byte) error {
	var event events.TransferEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		h.log.Error("Failed to unmarshal transfer event", zap.Error(err))
		return err
	}
	if err := h.useCase.SaveTransferEvent(ctx, event); err != nil {
		h.log.Error("Failed to save transfer event", zap.Error(err))
		return err
	}
	return nil
}
