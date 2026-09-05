package service

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type WithdrawQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, error)
	FindAllByCardNumber(ctx context.Context, req *requests.FindAllWithdrawCardNumber) ([]*repository.WithdrawQueryResult, *int, error)
	FindById(ctx context.Context, withdrawID int) (*models.Withdraw, error)
	FindByActive(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllWithdraws) ([]*repository.WithdrawQueryResult, *int, error)
}

type WithdrawCommandService interface {
	Create(ctx context.Context, request *requests.CreateWithdrawRequest) (*models.Withdraw, error)
	Update(ctx context.Context, request *requests.UpdateWithdrawRequest) (*models.Withdraw, error)
	TrashedWithdraw(ctx context.Context, withdraw_id int) (*models.Withdraw, error)
	RestoreWithdraw(ctx context.Context, withdraw_id int) (*models.Withdraw, error)
	DeleteWithdrawPermanent(ctx context.Context, withdraw_id int) (bool, error)

	RestoreAllWithdraw(ctx context.Context) (bool, error)
	DeleteAllWithdrawPermanent(ctx context.Context) (bool, error)
}
