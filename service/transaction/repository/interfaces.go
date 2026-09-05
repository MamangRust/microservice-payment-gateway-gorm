package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	saldoRepository "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/idempotency"
)

// TransactionQueryResult is the paginated transaction list response.
type TransactionQueryResult struct {
	Transaction *models.Transaction
	TotalCount  int64
}

// TransactionByCardResult is the paginated transaction list by card number.
type TransactionByCardResult struct {
	Transaction *models.Transaction
	TotalCount  int64
}

type IdempotencyRepository interface {
	idempotency.Store
}

type OutboxRepository interface {
	Insert(ctx context.Context, record OutboxRecord) error
}

type MerchantRepository interface {
	FindByApiKey(ctx context.Context, api_key string) (*models.Merchant, error)
}

type SaldoRepository interface {
	FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error)
	UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*saldoRepository.SaldoMutationResult, error)
}

type CardRepository interface {
	FindCardByUserId(ctx context.Context, user_id int) (*models.Card, error)
	FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.Card, error)
	FindCardByCardNumber(ctx context.Context, card_number string) (*models.Card, error)
	UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*models.Card, error)
}

type TransactionQueryRepository interface {
	FindAllTransactions(ctx context.Context, req *requests.FindAllTransactions) ([]*TransactionQueryResult, error)
	FindByActive(ctx context.Context, req *requests.FindAllTransactions) ([]*TransactionQueryResult, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTransactions) ([]*TransactionQueryResult, error)
	FindAllTransactionByCardNumber(ctx context.Context, req *requests.FindAllTransactionCardNumber) ([]*TransactionByCardResult, error)
	FindById(ctx context.Context, transaction_id int) (*models.Transaction, error)
	FindTransactionByMerchantId(ctx context.Context, merchant_id int) ([]*models.Transaction, error)
}

type TransactionCommandRepository interface {
	CreateTransaction(ctx context.Context, request *requests.CreateTransactionRequest) (*models.Transaction, error)
	UpdateTransaction(ctx context.Context, request *requests.UpdateTransactionRequest) (*models.Transaction, error)
	UpdateTransactionStatus(ctx context.Context, request *requests.UpdateTransactionStatus) (*models.Transaction, error)
	TransitionStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (*models.Transaction, error)
	GuardStatus(ctx context.Context, id int, fromStatus, toStatus, reason string) (bool, error)
	ListStuck(ctx context.Context, olderThan time.Duration, maxRows int32) ([]*models.Transaction, error)
	TrashedTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error)
	RestoreTransaction(ctx context.Context, topup_id int) (*models.Transaction, error)
	DeleteTransactionPermanent(ctx context.Context, topup_id int) (bool, error)

	RestoreAllTransaction(ctx context.Context) (bool, error)
	DeleteAllTransactionPermanent(ctx context.Context) (bool, error)
}
