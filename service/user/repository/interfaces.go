package repository

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type UserQueryRepository interface {
	FindAllUsers(ctx context.Context, req *requests.FindAllUsers) ([]*models.User, error)
	FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*models.User, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*models.User, error)
	FindById(ctx context.Context, user_id int) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByVerificationCode(ctx context.Context, code string) (*models.User, error)
	CountAllUsers(ctx context.Context, search string) (int64, error)
	CountActiveUsers(ctx context.Context, search string) (int64, error)
	CountTrashedUsers(ctx context.Context, search string) (int64, error)
}

type UserCommandRepository interface {
	CreateUser(ctx context.Context, request *requests.CreateUserRequest) (*models.User, error)
	UpdateUser(ctx context.Context, request *requests.UpdateUserRequest) (*models.User, error)
	UpdateIsVerified(ctx context.Context, userID int, isVerified bool) (*models.User, error)
	UpdatePassword(ctx context.Context, userID int, password string) (*models.User, error)
	TrashedUser(ctx context.Context, user_id int) (*models.User, error)
	RestoreUser(ctx context.Context, user_id int) (*models.User, error)
	DeleteUserPermanent(ctx context.Context, user_id int) (bool, error)
	RestoreAllUser(ctx context.Context) (bool, error)
	DeleteAllUserPermanent(ctx context.Context) (bool, error)
}

type RoleRepository interface {
	FindById(ctx context.Context, role_id int) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
}
