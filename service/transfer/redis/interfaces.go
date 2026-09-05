package mencache

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type TransferQueryCache interface {
	GetCachedTransfersCache(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, bool)
	SetCachedTransfersCache(ctx context.Context, req *requests.FindAllTransfers, data []*repository.TransferQueryResult, total *int)

	GetCachedTransferActiveCache(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, bool)
	SetCachedTransferActiveCache(ctx context.Context, req *requests.FindAllTransfers, data []*repository.TransferQueryResult, total *int)

	GetCachedTransferTrashedCache(ctx context.Context, req *requests.FindAllTransfers) ([]*repository.TransferQueryResult, *int, bool)
	SetCachedTransferTrashedCache(ctx context.Context, req *requests.FindAllTransfers, data []*repository.TransferQueryResult, total *int)

	GetCachedTransferCache(ctx context.Context, id int) (*models.Transfer, bool)
	SetCachedTransferCache(ctx context.Context, data *models.Transfer)

	GetCachedTransferByFrom(ctx context.Context, from string) ([]*models.Transfer, bool)
	SetCachedTransferByFrom(ctx context.Context, from string, data []*models.Transfer)

	GetCachedTransferByTo(ctx context.Context, to string) ([]*models.Transfer, bool)
	SetCachedTransferByTo(ctx context.Context, to string, data []*models.Transfer)
}

type TransferCommandCache interface {
	DeleteTransferCache(ctx context.Context, id int)
}
