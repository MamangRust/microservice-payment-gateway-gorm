package repository

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter/role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

// roleRepository adapts the role service's gRPC adapter to the user service's
// repository contract.
type roleRepository struct {
	adapter role.QueryRepository
}

func NewRoleRepository(a role.QueryRepository) *roleRepository {
	return &roleRepository{adapter: a}
}

func (r *roleRepository) FindById(ctx context.Context, id int) (*models.Role, error) {
	res, err := r.adapter.FindById(ctx, id)
	if err != nil {
		return nil, sharedErrors.ErrRoleNotFound.WithInternal(err)
	}
	if res == nil {
		return nil, sharedErrors.ErrRoleNotFound
	}
	return res, nil
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	res, err := r.adapter.FindByName(ctx, name)
	if err != nil {
		return nil, sharedErrors.ErrRoleNotFound.WithInternal(err)
	}
	if res == nil {
		return nil, sharedErrors.ErrRoleNotFound
	}
	return res, nil
}
