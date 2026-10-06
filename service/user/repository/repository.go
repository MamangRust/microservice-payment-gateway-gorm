package repository

import (
	pbroles "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	pbuserroles "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user_role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	roleadapter "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter/role"
	userroleadapter "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter/user_role"
	"gorm.io/gorm"
)

type GuardOptions struct {
	UserRole []adapter.GuardOption
	Role     []adapter.GuardOption
}

type Repositories struct {
	UserCommand UserCommandRepository
	UserQuery   UserQueryRepository
	Role        RoleRepository
	UserRole    UserRoleRepository
}

type Deps struct {
	Db              *gorm.DB
	RoleQueryClient pbroles.RoleQueryServiceClient
	UserRoleClient  pbuserroles.UserRoleServiceClient
	Guard           GuardOptions
}

func NewRepositories(deps *Deps) *Repositories {
	return &Repositories{
		UserCommand: NewUserCommandRepository(deps.Db),
		UserQuery:   NewUserQueryRepository(deps.Db),
		Role:        NewRoleRepository(roleadapter.New(deps.RoleQueryClient, deps.Guard.Role...)),
		UserRole:    userroleadapter.New(deps.UserRoleClient, deps.Guard.UserRole...),
	}
}
