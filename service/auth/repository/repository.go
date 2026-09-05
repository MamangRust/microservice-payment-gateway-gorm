package repository

import (
	pb_role "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	pb_user "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"gorm.io/gorm"
)

type Repositories struct {
	User         UserRepository
	RefreshToken RefreshTokenRepository
	UserRole     UserRoleRepository
	Role         RoleRepository
	ResetToken   ResetTokenRepository
}

type RepositoriesDeps struct {
	DB                *gorm.DB
	UserQueryClient   pb_user.UserQueryServiceClient
	UserCommandClient pb_user.UserCommandServiceClient
	RoleQueryClient   pb_role.RoleQueryServiceClient
	RoleCommandClient pb_role.RoleCommandServiceClient
}

func NewRepositories(deps *RepositoriesDeps) *Repositories {
	return &Repositories{
		User:         NewUserRepository(deps.UserQueryClient, deps.UserCommandClient),
		UserRole:     NewUserRoleRepository(deps.RoleCommandClient),
		RefreshToken: NewRefreshTokenRepository(deps.DB),
		Role:         NewRoleRepository(deps.RoleQueryClient),
		ResetToken:   NewResetTokenRepository(deps.DB),
	}
}
