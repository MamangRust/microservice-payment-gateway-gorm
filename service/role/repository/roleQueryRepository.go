package repository

import (
	"context"
	"errors"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type roleQueryRepository struct {
	db *gorm.DB
}

func NewRoleQueryRepository(db *gorm.DB) RoleQueryRepository {
	return &roleQueryRepository{db: db}
}

func (r *roleQueryRepository) FindAllRoles(ctx context.Context, req *requests.FindAllRoles) ([]*models.Role, error) {
	offset := (req.Page - 1) * req.PageSize

	var roles []*models.Role
	query := r.db.WithContext(ctx).Model(&models.Role{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		query = query.Where("role_name ILIKE ?", "%"+req.Search+"%")
	}

	err := query.Offset(offset).Limit(req.PageSize).Order("role_id ASC").Find(&roles).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("find all roles").WithInternal(err)
	}
	return roles, nil
}

func (r *roleQueryRepository) FindByActiveRole(ctx context.Context, req *requests.FindAllRoles) ([]*models.Role, error) {
	offset := (req.Page - 1) * req.PageSize

	var roles []*models.Role
	query := r.db.WithContext(ctx).Model(&models.Role{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		query = query.Where("role_name ILIKE ?", "%"+req.Search+"%")
	}

	err := query.Offset(offset).Limit(req.PageSize).Order("role_id ASC").Find(&roles).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("find active roles").WithInternal(err)
	}
	return roles, nil
}

func (r *roleQueryRepository) FindByTrashedRole(ctx context.Context, req *requests.FindAllRoles) ([]*models.Role, error) {
	offset := (req.Page - 1) * req.PageSize

	var roles []*models.Role
	query := r.db.WithContext(ctx).Model(&models.Role{}).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		query = query.Where("role_name ILIKE ?", "%"+req.Search+"%")
	}

	err := query.Offset(offset).Limit(req.PageSize).Order("role_id ASC").Find(&roles).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed roles").WithInternal(err)
	}
	return roles, nil
}

func (r *roleQueryRepository) FindById(ctx context.Context, id int) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("role_id = ?", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.ErrRoleNotFound.WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &role, nil
}

func (r *roleQueryRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("role_name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.ErrRoleNotFound.WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &role, nil
}

func (r *roleQueryRepository) FindByUserId(ctx context.Context, userID int) ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.role_id").
		Where("user_roles.user_id = ? AND roles.deleted_at IS NULL", userID).
		Find(&roles).Error
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return roles, nil
}

func (r *roleQueryRepository) CountAllRoles(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Role{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("role_name ILIKE ?", "%"+search+"%")
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *roleQueryRepository) CountActiveRoles(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Role{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("role_name ILIKE ?", "%"+search+"%")
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *roleQueryRepository) CountTrashedRoles(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Role{}).Where("deleted_at IS NOT NULL")
	if search != "" {
		query = query.Where("role_name ILIKE ?", "%"+search+"%")
	}
	err := query.Count(&count).Error
	return count, err
}
