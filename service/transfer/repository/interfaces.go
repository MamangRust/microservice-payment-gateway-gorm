package repository

import (
	"context"
	"time"

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
}

type CardRepository interface {
	FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.Card, error)
	FindCardByCardNumber(ctx context.Context, card_number string) (*models.Card, error)
}

type TransferQueryResult struct {
	Transfer   *models.Transfer
	TotalCount int64
}

type TransferQueryRepository interface {
	FindAll(ctx context.Context, req *requests.FindAllTransfers) ([]*TransferQueryResult, error)
	FindByActive(ctx context.Context, req *requests.FindAllTransfers) ([]*TransferQueryResult, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTransfers) ([]*TransferQueryResult, error)
	FindById(ctx context.Context, id int) (*models.Transfer, error)
	FindTransferByTransferFrom(ctx context.Context, transferFrom string) ([]*models.Transfer, error)
	FindTransferByTransferTo(ctx context.Context, transferTo string) ([]*models.Transfer, error)
}

type TransferCommandRepository interface {
	CreateTransfer(ctx context.Context, request *requests.CreateTransferRequest) (*models.Transfer, error)
	UpdateTransfer(ctx context.Context, request *requests.UpdateTransferRequest) (*models.Transfer, error)
	UpdateTransferAmount(ctx context.Context, request *requests.UpdateTransferAmountRequest) (*models.Transfer, error)
	UpdateTransferStatus(ctx context.Context, request *requests.UpdateTransferStatus) (*models.Transfer, error)
	TransitionStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (*models.Transfer, error)
	GuardStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (bool, error)
	ListStuck(ctx context.Context, olderThan time.Duration, maxRows int32) ([]*models.Transfer, error)

	TrashedTransfer(ctx context.Context, transferID int) (*models.Transfer, error)
	RestoreTransfer(ctx context.Context, transferID int) (*models.Transfer, error)
	DeleteTransferPermanent(ctx context.Context, transferID int) (bool, error)

	RestoreAllTransfer(ctx context.Context) (bool, error)
	DeleteAllTransferPermanent(ctx context.Context) (bool, error)
}
