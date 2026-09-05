package repository

import (
	"context"
	"net/http"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

type userRepository struct {
	userQueryClient   pb.UserQueryServiceClient
	userCommandClient pb.UserCommandServiceClient
}

func NewUserRepository(queryClient pb.UserQueryServiceClient, commandClient pb.UserCommandServiceClient) *userRepository {
	return &userRepository{
		userQueryClient:   queryClient,
		userCommandClient: commandClient,
	}
}

func (r *userRepository) FindById(ctx context.Context, user_id int) (*models.User, error) {
	resp, err := r.userQueryClient.FindById(ctx, &pb.FindByIdUserRequest{
		Id: int32(user_id),
	})
	if err != nil {
		return nil, sharedErrors.ErrUserNotFound.WithInternal(err)
	}

	user := resp.GetData()
	createdAt, _ := time.Parse(time.RFC3339, user.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, user.UpdatedAt)

	return &models.User{
		UserID:    int32(user.Id),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	resp, err := r.userQueryClient.FindByEmail(ctx, &pb.FindByEmailUserRequest{
		Email: email,
	})
	if err != nil {
		return nil, sharedErrors.ErrUserNotFound.WithInternal(err)
	}

	user := resp.GetData()

	return &models.User{
		UserID:   int32(user.Id),
		Email:    user.Email,
		Password: user.Password,
	}, nil
}

func (r *userRepository) FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error) {
	resp, err := r.userQueryClient.FindByEmail(ctx, &pb.FindByEmailUserRequest{
		Email: email,
	})
	if err != nil {
		return nil, sharedErrors.ErrUserNotFound.WithInternal(err)
	}

	user := resp.GetData()
	createdAt, _ := time.Parse(time.RFC3339, user.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, user.UpdatedAt)

	return &models.User{
		UserID:    int32(user.Id),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		Password:  user.Password,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}, nil
}

func (r *userRepository) FindByVerificationCode(ctx context.Context, verification_code string) (*models.User, error) {
	resp, err := r.userQueryClient.FindByVerificationCode(ctx, &pb.FindByVerificationCodeUserRequest{
		VerificationCode: verification_code,
	})
	if err != nil {
		return nil, sharedErrors.ErrUserNotFound.WithInternal(err)
	}

	user := resp.GetData()
	createdAt, _ := time.Parse(time.RFC3339, user.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, user.UpdatedAt)

	return &models.User{
		UserID:    int32(user.Id),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}, nil
}

func (r *userRepository) CreateUser(ctx context.Context, request *requests.RegisterRequest) (*models.User, error) {
	resp, err := r.userCommandClient.Create(ctx, &pb.CreateUserRequest{
		Firstname:       request.FirstName,
		Lastname:        request.LastName,
		Email:           request.Email,
		Password:        request.Password,
		ConfirmPassword: request.Password,
	})
	if err != nil {
		if appErr := sharedErrors.ParseGrpcError(err); appErr != nil && appErr.Code == http.StatusConflict {
			return nil, sharedErrors.NewConflictError("email already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create user").WithInternal(err)
	}

	user := resp.GetData()
	return &models.User{
		UserID:    int32(user.Id),
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
	}, nil
}

func (r *userRepository) UpdateUserIsVerified(ctx context.Context, user_id int, is_verified bool) (*models.User, error) {
	resp, err := r.userCommandClient.UpdateIsVerified(ctx, &pb.UpdateUserIsVerifiedRequest{
		UserId:     int32(user_id),
		IsVerified: is_verified,
	})
	if err != nil {
		return nil, sharedErrors.ErrFailed("update user verification code").WithInternal(err)
	}

	user := resp.GetData()
	return &models.User{
		UserID: int32(user.Id),
	}, nil
}

func (r *userRepository) UpdateUserPassword(ctx context.Context, user_id int, password string) (*models.User, error) {
	resp, err := r.userCommandClient.UpdatePassword(ctx, &pb.UpdateUserPasswordRequest{
		UserId:   int32(user_id),
		Password: password,
	})
	if err != nil {
		return nil, sharedErrors.ErrFailed("update user password").WithInternal(err)
	}

	user := resp.GetData()
	return &models.User{
		UserID: int32(user.Id),
	}, nil
}
