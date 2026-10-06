package usecase

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
)

func (u *statsUseCase) SaveMerchantEvent(ctx context.Context, event events.MerchantEvent) error {
	return u.repo.InsertMerchantEvent(ctx, event)
}
