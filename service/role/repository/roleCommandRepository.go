package repository

import (
	"context"
	"strings"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type roleCommandRepository struct {
	db *gorm.DB
}

func NewRoleCommandRepository(db *gorm.DB) RoleCommandRepository {
	return &roleCommandRepository{db: db}
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "unique constraint") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}

func (r *roleCommandRepository) CreateRole(ctx context.Context, req *requests.CreateRoleRequest) (*models.Role, error) {
	role := &models.Role{RoleName: req.Name}
	err := r.db.WithContext(ctx).Create(role).Error
	if err != nil {
		if isUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("role name already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create role").WithInternal(err)
	}
	return role, nil
}

func (r *roleCommandRepository) UpdateRole(ctx context.Context, req *requests.UpdateRoleRequest) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("role_id = ?", *req.ID).First(&role).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "role", "update role")
	}

	err = r.db.WithContext(ctx).Model(&role).Update("role_name", req.Name).Error
	if err != nil {
		if isUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("role name already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("update role").WithInternal(err)
	}

	// Re-read to get updated timestamp
	err = r.db.WithContext(ctx).Where("role_id = ?", *req.ID).First(&role).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("re-read role after update").WithInternal(err)
	}
	return &role, nil
}

func (r *roleCommandRepository) TrashedRole(ctx context.Context, id int) (*models.Role, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.Role{}).
		Where("role_id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return nil, sharedErrors.ErrFailed("trash role").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, sharedErrors.ErrNotFoundResponse("role")
	}

	var role models.Role
	if err := r.db.WithContext(ctx).Where("role_id = ?", id).First(&role).Error; err != nil {
		return nil, sharedErrors.ErrFailed("get trashed role").WithInternal(err)
	}
	return &role, nil
}

func (r *roleCommandRepository) RestoreRole(ctx context.Context, id int) (*models.Role, error) {
	result := r.db.WithContext(ctx).Model(&models.Role{}).
		Where("role_id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil)
	if result.Error != nil {
		return nil, sharedErrors.ErrFailed("restore role").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, sharedErrors.ErrNotFoundResponse("role")
	}

	var role models.Role
	if err := r.db.WithContext(ctx).Where("role_id = ?", id).First(&role).Error; err != nil {
		return nil, sharedErrors.ErrFailed("get restored role").WithInternal(err)
	}
	return &role, nil
}

func (r *roleCommandRepository) DeleteRolePermanent(ctx context.Context, id int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("role_id = ?", id).Delete(&models.Role{})
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("delete role permanently").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return false, sharedErrors.ErrNotFoundResponse("role")
	}
	return true, nil
}

func (r *roleCommandRepository) RestoreAllRole(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Model(&models.Role{}).
		Where("deleted_at IS NOT NULL").
		Update("deleted_at", nil)
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("restore all roles").WithInternal(result.Error)
	}
	return true, nil
}

func (r *roleCommandRepository) DeleteAllRolePermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Role{})
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("delete all roles permanently").WithInternal(result.Error)
	}
	return true, nil
}

func (r *roleCommandRepository) CreateUserRole(ctx context.Context, userID, roleID int) (*models.Role, error) {
	userRole := &models.UserRole{
		UserID: int32(userID),
		RoleID: int32(roleID),
	}
	err := r.db.WithContext(ctx).Create(userRole).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("create user role").WithInternal(err)
	}

	var role models.Role
	if err := r.db.WithContext(ctx).Where("role_id = ?", roleID).First(&role).Error; err != nil {
		return nil, sharedErrors.ErrFailed("get role for user role").WithInternal(err)
	}
	return &role, nil
}

func (r *roleCommandRepository) DeleteUserRole(ctx context.Context, userID, roleID int) (bool, error) {
	result := r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&models.UserRole{})
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("delete user role").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return false, sharedErrors.ErrNotFoundResponse("user role")
	}
	return true, nil
}
