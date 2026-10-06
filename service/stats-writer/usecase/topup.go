package usecase

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
)

func (u *statsUseCase) SaveTopupEvent(ctx context.Context, event events.TopupEvent) error {
	return u.repo.InsertTopupEvent(ctx, event)
}
