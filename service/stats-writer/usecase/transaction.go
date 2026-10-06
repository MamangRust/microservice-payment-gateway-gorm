package usecase

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
)

func (u *statsUseCase) SaveTransactionEvent(ctx context.Context, event events.TransactionEvent) error {
	return u.repo.InsertTransactionEvent(ctx, event)
}
