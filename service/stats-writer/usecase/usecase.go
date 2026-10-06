package usecase

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-writer/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
)

// UseCase maps each Kafka event type to a repository insert.
//
// Implementasinya dipecah per domain — satu file per event type (transaction.go,
// topup.go, dst.) supaya menambah event baru tidak menumpuk di satu file.
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

func (u *statsUseCase) Close() error {
	return u.repo.Close()
}
