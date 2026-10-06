package usecase

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
)

func (u *statsUseCase) SaveWithdrawEvent(ctx context.Context, event events.WithdrawEvent) error {
	return u.repo.InsertWithdrawEvent(ctx, event)
}
