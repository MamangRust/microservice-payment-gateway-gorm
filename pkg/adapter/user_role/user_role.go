// Package user_role adapts the user_role service's gRPC API into the shared
// domain model. It owns the only place that talks to pb/user_role's
// UserRoleService.
package user_role

import (
	"context"

	pbuserrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user_role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

// CommandRepository is the write path consumers use to (un)assign roles.
type CommandRepository interface {
	AssignRoleToUser(ctx context.Context, request *requests.CreateUserRoleRequest) (*models.UserRole, error)
	RemoveRoleFromUser(ctx context.Context, request *requests.RemoveUserRoleRequest) error
}

type userRoleGRPCAdapter struct {
	commandClient pbuserrole.UserRoleServiceClient
	guard         *resilience.DependencyGuard
}

func (a *userRoleGRPCAdapter) SetGuard(g *resilience.DependencyGuard) { a.guard = g }

// New builds a user-role command adapter. Passing zero options leaves the guard
// nil, which makes CallGuarded a plain passthrough.
func New(commandClient pbuserrole.UserRoleServiceClient, opts ...adapter.GuardOption) CommandRepository {
	a := &userRoleGRPCAdapter{commandClient: commandClient}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *userRoleGRPCAdapter) AssignRoleToUser(ctx context.Context, request *requests.CreateUserRoleRequest) (*models.UserRole, error) {
	err := adapter.CallGuarded(ctx, a.guard, func(callCtx context.Context) error {
		_, callErr := a.commandClient.CreateUserRole(callCtx, &pbuserrole.CreateUserRoleRequest{
			UserId: int32(request.UserId),
			RoleId: int32(request.RoleId),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return &models.UserRole{UserID: int32(request.UserId), RoleID: int32(request.RoleId)}, nil
}

func (a *userRoleGRPCAdapter) RemoveRoleFromUser(ctx context.Context, request *requests.RemoveUserRoleRequest) error {
	return adapter.CallGuarded(ctx, a.guard, func(callCtx context.Context) error {
		_, callErr := a.commandClient.DeleteUserRole(callCtx, &pbuserrole.DeleteUserRoleRequest{
			UserId: int32(request.UserId),
			RoleId: int32(request.RoleId),
		})
		return callErr
	})
}
