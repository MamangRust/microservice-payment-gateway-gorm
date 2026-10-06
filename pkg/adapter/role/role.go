// Package role adapts the role service's query gRPC API into the shared domain
// model. Role assignment lives in the user_role service, so this adapter only
// exposes the read path.
package role

import (
	"context"

	pbrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
)

// QueryRepository reads roles from the role service over gRPC.
type QueryRepository interface {
	FindById(ctx context.Context, id int) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
}

type roleGRPCAdapter struct {
	queryClient pbrole.RoleQueryServiceClient
	guard       *resilience.DependencyGuard
}

func (a *roleGRPCAdapter) SetGuard(g *resilience.DependencyGuard) { a.guard = g }

// New builds a role query adapter. Passing zero options leaves the guard nil,
// which makes CallGuarded a plain passthrough.
func New(queryClient pbrole.RoleQueryServiceClient, opts ...adapter.GuardOption) QueryRepository {
	a := &roleGRPCAdapter{queryClient: queryClient}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *roleGRPCAdapter) FindById(ctx context.Context, id int) (*models.Role, error) {
	var resp *pbrole.ApiResponseRole
	err := adapter.CallGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindByIdRole(callCtx, &pbrole.FindByIdRoleRequest{RoleId: int32(id)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, nil
	}
	return &models.Role{RoleID: resp.Data.Id, RoleName: resp.Data.Name}, nil
}

func (a *roleGRPCAdapter) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var resp *pbrole.ApiResponseRole
	err := adapter.CallGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindByNameRole(callCtx, &pbrole.FindByNameRoleRequest{Name: name})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, nil
	}
	return &models.Role{RoleID: resp.Data.Id, RoleName: resp.Data.Name}, nil
}
