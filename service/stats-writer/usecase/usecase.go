package usecase

import (
	"context"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-writer/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
)

type UseCase interface {
	SaveTransactionEvent(ctx context.Context, event events.TransactionEvent) error
	SaveTopupEvent(ctx context.Context, event events.TopupEvent) error
	SaveTransferEvent(ctx context.Context, event events.TransferEvent) error
	SaveWithdrawEvent(ctx context.Context, event events.WithdrawEvent) error
	SaveSaldoEvent(ctx context.Context, event events.SaldoEvent) error
	SaveMerchantEvent(ctx context.Context, event events.MerchantEvent) error
	SaveCardEvent(ctx context.Context, event events.CardEvent) error

	Close() error
}

type statsUseCase struct {
	repo repository.Repository
}

func NewStatsUseCase(repo repository.Repository) UseCase {
	return &statsUseCase{
		repo: repo,
	}
}

func (u *statsUseCase) SaveTransactionEvent(ctx context.Context, event events.TransactionEvent) error {
	return u.repo.InsertTransactionEvent(ctx, event)
}

func (u *statsUseCase) SaveTopupEvent(ctx context.Context, event events.TopupEvent) error {
	return u.repo.InsertTopupEvent(ctx, event)
}

func (u *statsUseCase) SaveTransferEvent(ctx context.Context, event events.TransferEvent) error {
	return u.repo.InsertTransferEvent(ctx, event)
}

func (u *statsUseCase) SaveWithdrawEvent(ctx context.Context, event events.WithdrawEvent) error {
	return u.repo.InsertWithdrawEvent(ctx, event)
}

func (u *statsUseCase) SaveSaldoEvent(ctx context.Context, event events.SaldoEvent) error {
	return u.repo.InsertSaldoEvent(ctx, event)
}

func (u *statsUseCase) SaveMerchantEvent(ctx context.Context, event events.MerchantEvent) error {
	return u.repo.InsertMerchantEvent(ctx, event)
}

func (u *statsUseCase) SaveCardEvent(ctx context.Context, event events.CardEvent) error {
	return u.repo.InsertCardEvent(ctx, event)
}

func (u *statsUseCase) Close() error {
	return u.repo.Close()
}
