package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

type userRepository struct {
	userQueryClient pb.UserQueryServiceClient
}

func NewUserRepository(userQueryClient pb.UserQueryServiceClient) UserRepository {
	return &userRepository{
		userQueryClient: userQueryClient,
	}
}

func (r *userRepository) FindById(ctx context.Context, user_id int) (*models.User, error) {
	resp, err := r.userQueryClient.FindById(ctx, &pb.FindByIdUserRequest{
		Id: int32(user_id),
	})

	if err != nil {
		return nil, sharedErrors.ErrNotFound.WithMessage("user not found").WithInternal(err)
	}

	if resp == nil || resp.Data == nil {
		return nil, sharedErrors.ErrNotFound.WithMessage("user not found")
	}

	parseTime := func(ts string) *time.Time {
		if ts == "" {
			return nil
		}
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			return nil
		}
		return &t
	}

	return &models.User{
		UserID:    resp.Data.Id,
		Firstname: resp.Data.Firstname,
		Lastname:  resp.Data.Lastname,
		Email:     resp.Data.Email,
		CreatedAt: parseTime(resp.Data.CreatedAt),
		UpdatedAt: parseTime(resp.Data.UpdatedAt),
	}, nil
}
