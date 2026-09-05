package repository

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	saldoRepository "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/idempotency"
)

type IdempotencyRepository interface {
	idempotency.Store
}

type OutboxRepository interface {
	Insert(ctx context.Context, record OutboxRecord) error
}

type SaldoRepository interface {
	FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error)
	UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*saldoRepository.SaldoMutationResult, error)
	UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*saldoRepository.SaldoMutationResult, error)
}

type WithdrawQueryResult struct {
	Withdraw   *models.Withdraw
	TotalCount int64
}

type WithdrawQueryRepository interface {
	FindAll(ctx context.Context, req *requests.FindAllWithdraws) ([]*WithdrawQueryResult, error)
	FindByActive(ctx context.Context, req *requests.FindAllWithdraws) ([]*WithdrawQueryResult, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllWithdraws) ([]*WithdrawQueryResult, error)
	FindAllByCardNumber(ctx context.Context, req *requests.FindAllWithdrawCardNumber) ([]*WithdrawQueryResult, error)
	FindById(ctx context.Context, id int) (*models.Withdraw, error)
	GetTodayWithdrawSumByCardNumber(ctx context.Context, cardNumber string) (int64, error)
}

type WithdrawCommandRepository interface {
	CreateWithdraw(ctx context.Context, request *requests.CreateWithdrawRequest) (*models.Withdraw, error)
	UpdateWithdraw(ctx context.Context, request *requests.UpdateWithdrawRequest) (*models.Withdraw, error)
	UpdateWithdrawStatus(ctx context.Context, request *requests.UpdateWithdrawStatus) (*models.Withdraw, error)

	TrashedWithdraw(ctx context.Context, withdrawID int) (*models.Withdraw, error)
	RestoreWithdraw(ctx context.Context, withdrawID int) (*models.Withdraw, error)
	DeleteWithdrawPermanent(ctx context.Context, withdrawID int) (bool, error)

	RestoreAllWithdraw(ctx context.Context) (bool, error)
	DeleteAllWithdrawPermanent(ctx context.Context) (bool, error)
}

type CardRepository interface {
	FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.Card, error)
}
