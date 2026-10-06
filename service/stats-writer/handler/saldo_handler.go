package handler

import (
	"context"
	"encoding/json"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
	"go.uber.org/zap"
)

func (h *StatsHandler) handleSaldo(ctx context.Context, raw []byte) error {
	var event events.SaldoEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		h.log.Error("Failed to unmarshal saldo event", zap.Error(err))
		return err
	}
	if err := h.useCase.SaveSaldoEvent(ctx, event); err != nil {
		h.log.Error("Failed to save saldo event", zap.Error(err))
		return err
	}
	return nil
}
