package service

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type TransferQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, error)
	FindById(ctx context.Context, transferId int) (*models.Transfer, error)
	FindByActive(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, error)
	FindTransferByTransferFrom(ctx context.Context, transfer_from string) ([]*models.Transfer, error)
	FindTransferByTransferTo(ctx context.Context, transfer_to string) ([]*models.Transfer, error)
}

type TransferCommandService interface {
	CreateTransaction(ctx context.Context, request *requests.CreateTransferRequest) (*models.Transfer, error)
	UpdateTransaction(ctx context.Context, request *requests.UpdateTransferRequest) (*models.Transfer, error)
	TrashedTransfer(ctx context.Context, transfer_id int) (*models.Transfer, error)
	RestoreTransfer(ctx context.Context, transfer_id int) (*models.Transfer, error)
	DeleteTransferPermanent(ctx context.Context, transfer_id int) (bool, error)

	RestoreAllTransfer(ctx context.Context) (bool, error)
	DeleteAllTransferPermanent(ctx context.Context) (bool, error)
	RecoverStuckOperations(ctx context.Context, olderThan time.Duration, maxRows int32) error
	StartRecoveryWorker(ctx context.Context, interval, olderThan time.Duration, maxRows int32)
}
