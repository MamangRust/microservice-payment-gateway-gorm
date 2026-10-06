package usecase

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
)

func (u *statsUseCase) SaveSaldoEvent(ctx context.Context, event events.SaldoEvent) error {
	return u.repo.InsertSaldoEvent(ctx, event)
}
