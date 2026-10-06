package repository

import (
	pbrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	pbuser "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	pbuserrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user_role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	roleadapter "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter/role"
	userroleadapter "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter/user_role"
	"gorm.io/gorm"
)

type Repositories struct {
	User         UserRepository
	RefreshToken RefreshTokenRepository
	UserRole     UserRoleRepository
	Role         RoleRepository
	ResetToken   ResetTokenRepository
}

type GuardOptions struct {
	User     []adapter.GuardOption
	UserRole []adapter.GuardOption
	Role     []adapter.GuardOption
}

func NewRepositories(
	db *gorm.DB,
	userQueryClient pbuser.UserQueryServiceClient,
	userCommandClient pbuser.UserCommandServiceClient,
	roleQueryClient pbrole.RoleQueryServiceClient,
	userRoleClient pbuserrole.UserRoleServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		User:         adapter.NewAuthUserAdapter(userQueryClient, userCommandClient, g.User...),
		UserRole:     userroleadapter.New(userRoleClient, g.UserRole...),
		RefreshToken: NewRefreshTokenRepository(db),
		Role:         roleadapter.New(roleQueryClient, g.Role...),
		ResetToken:   NewResetTokenRepository(db),
	}
}
