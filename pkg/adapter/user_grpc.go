package adapter

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
)

type UserAdapter interface {
	FindById(ctx context.Context, userID int) (*models.User, error)
}

type userGRPCAdapter struct {
	queryClient user.UserQueryServiceClient
	guard       *resilience.DependencyGuard
}

func (a *userGRPCAdapter) setGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func NewUserAdapter(queryClient user.UserQueryServiceClient, opts ...func(guardSetter)) UserAdapter {
	a := &userGRPCAdapter{
		queryClient: queryClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *userGRPCAdapter) FindById(ctx context.Context, userID int) (*models.User, error) {
	var resp *user.ApiResponseUser
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindById(callCtx, &user.FindByIdUserRequest{
			Id: int32(userID),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}

	return &models.User{
		UserID:    resp.Data.Id,
		Email:     resp.Data.Email,
		Firstname: resp.Data.Firstname,
		Lastname:  resp.Data.Lastname,
	}, nil
}

type localUserAdapter struct {
	repo repository.UserQueryRepository
}

func NewLocalUserAdapter(repo repository.UserQueryRepository) UserAdapter {
	return &localUserAdapter{
		repo: repo,
	}
}

func (a *localUserAdapter) FindById(ctx context.Context, userID int) (*models.User, error) {
	res, err := a.repo.FindById(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.User{
		UserID:    res.UserID,
		Email:     res.Email,
		Firstname: res.Firstname,
		Lastname:  res.Lastname,
	}, nil
}
