package usecase

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
)

func (u *statsUseCase) SaveCardEvent(ctx context.Context, event events.CardEvent) error {
	return u.repo.InsertCardEvent(ctx, event)
}
