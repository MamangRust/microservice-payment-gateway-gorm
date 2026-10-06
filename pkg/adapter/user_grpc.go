package adapter

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

// UserAdapter is the query-only surface used by consumers that just look up
// users (card, merchant, topup, ...).
type UserAdapter interface {
	FindById(ctx context.Context, userID int) (*models.User, error)
}

// AuthUserAdapter extends UserAdapter with the lookup/creation/update surface
// the auth service needs. It is satisfied by the gRPC adapter below (no local
// repository wrapper).
type AuthUserAdapter interface {
	UserAdapter
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error)
	FindByVerificationCode(ctx context.Context, code string) (*models.User, error)
	CreateUser(ctx context.Context, request *requests.RegisterRequest) (*models.User, error)
	UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) (*models.User, error)
	UpdateUserPassword(ctx context.Context, userID int, password string) (*models.User, error)
}

type userGRPCAdapter struct {
	queryClient user.UserQueryServiceClient
	guard       *resilience.DependencyGuard
}

func (a *userGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func NewUserAdapter(queryClient user.UserQueryServiceClient, opts ...GuardOption) UserAdapter {
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
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
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
		CreatedAt: parseTime(resp.Data.CreatedAt),
		UpdatedAt: parseTime(resp.Data.UpdatedAt),
	}, nil
}

// NewAuthUserAdapter builds the full user surface needed by the auth service
// (lookups by email/verification code plus user creation and status/password
// updates) on top of the user query and command clients.
func NewAuthUserAdapter(queryClient user.UserQueryServiceClient, commandClient user.UserCommandServiceClient, opts ...GuardOption) AuthUserAdapter {
	a := &authUserGRPCAdapter{
		userGRPCAdapter: userGRPCAdapter{queryClient: queryClient},
		commandClient:   commandClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

type authUserGRPCAdapter struct {
	userGRPCAdapter
	commandClient user.UserCommandServiceClient
}

func (a *authUserGRPCAdapter) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var resp *user.ApiResponseUser
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindByEmail(callCtx, &user.FindByEmailUserRequest{Email: email})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapUserResponse(resp.Data), nil
}

func (a *authUserGRPCAdapter) FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error) {
	return a.FindByEmail(ctx, email)
}

func (a *authUserGRPCAdapter) FindByVerificationCode(ctx context.Context, code string) (*models.User, error) {
	var resp *user.ApiResponseUser
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindByVerificationCode(callCtx, &user.FindByVerificationCodeUserRequest{VerificationCode: code})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapUserResponse(resp.Data), nil
}

func (a *authUserGRPCAdapter) CreateUser(ctx context.Context, request *requests.RegisterRequest) (*models.User, error) {
	var resp *user.ApiResponseUser
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.commandClient.Create(callCtx, &user.CreateUserRequest{
			Firstname:       request.FirstName,
			Lastname:        request.LastName,
			Email:           request.Email,
			Password:        request.Password,
			ConfirmPassword: request.Password,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapUserResponse(resp.Data), nil
}

func (a *authUserGRPCAdapter) UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) (*models.User, error) {
	var resp *user.ApiResponseUser
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.commandClient.UpdateIsVerified(callCtx, &user.UpdateUserIsVerifiedRequest{
			UserId:     int32(userID),
			IsVerified: isVerified,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapUserResponse(resp.GetData()), nil
}

func (a *authUserGRPCAdapter) UpdateUserPassword(ctx context.Context, userID int, password string) (*models.User, error) {
	var resp *user.ApiResponseUser
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.commandClient.UpdatePassword(callCtx, &user.UpdateUserPasswordRequest{
			UserId:   int32(userID),
			Password: password,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapUserResponse(resp.GetData()), nil
}

func mapUserResponse(u *user.UserResponse) *models.User {
	if u == nil {
		return nil
	}
	return &models.User{
		UserID:    u.Id,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: parseTime(u.CreatedAt),
		UpdatedAt: parseTime(u.UpdatedAt),
	}
}
