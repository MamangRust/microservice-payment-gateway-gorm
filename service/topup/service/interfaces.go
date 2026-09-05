package service

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

// TopupQueryService defines the read-only operations for querying topup data.
type TopupQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, error)
	FindAllByCardNumber(ctx context.Context, req *requests.FindAllTopupsByCardNumber) ([]*repository.TopupByCardResult, *int, error)
	FindById(ctx context.Context, topupID int) (*models.Topup, error)
	FindByActive(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTopups) ([]*repository.TopupQueryResult, *int, error)
}

type TopupCommandService interface {
	CreateTopup(ctx context.Context, request *requests.CreateTopupRequest) (*models.Topup, error)
	UpdateTopup(ctx context.Context, request *requests.UpdateTopupRequest) (*models.Topup, error)
	TrashedTopup(ctx context.Context, topup_id int) (*models.Topup, error)
	RestoreTopup(ctx context.Context, topup_id int) (*models.Topup, error)
	DeleteTopupPermanent(ctx context.Context, topup_id int) (bool, error)

	RestoreAllTopup(ctx context.Context) (bool, error)
	DeleteAllTopupPermanent(ctx context.Context) (bool, error)
}
