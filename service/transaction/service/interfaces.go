package service

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

// TransactionQueryService handles queries related to transactions.
type TransactionQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, *int, error)
	FindAllByCardNumber(ctx context.Context, req *requests.FindAllTransactionCardNumber) ([]*repository.TransactionByCardResult, *int, error)
	FindById(ctx context.Context, transactionID int) (*models.Transaction, error)
	FindByActive(ctx context.Context, req *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTransactions) ([]*repository.TransactionQueryResult, *int, error)
	FindTransactionByMerchantId(ctx context.Context, merchant_id int) ([]*models.Transaction, error)
}

type TransactionCommandService interface {
	Create(ctx context.Context, apiKey string, request *requests.CreateTransactionRequest) (*models.Transaction, error)
	Update(ctx context.Context, apiKey string, request *requests.UpdateTransactionRequest) (*models.Transaction, error)
	TrashedTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error)
	RestoreTransaction(ctx context.Context, transaction_id int) (*models.Transaction, error)
	DeleteTransactionPermanent(ctx context.Context, transaction_id int) (bool, error)

	RestoreAllTransaction(ctx context.Context) (bool, error)
	DeleteAllTransactionPermanent(ctx context.Context) (bool, error)
	RecoverStuckOperations(ctx context.Context, olderThan time.Duration, maxRows int32) error
	StartRecoveryWorker(ctx context.Context, interval, olderThan time.Duration, maxRows int32)
}
