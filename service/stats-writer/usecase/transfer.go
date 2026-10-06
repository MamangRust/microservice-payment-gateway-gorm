package usecase

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
)

func (u *statsUseCase) SaveTransferEvent(ctx context.Context, event events.TransferEvent) error {
	return u.repo.InsertTransferEvent(ctx, event)
}
