package repository

import (
	"context"
	"strings"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userCommandRepository struct {
	db *gorm.DB
}

func NewUserCommandRepository(db *gorm.DB) UserCommandRepository {
	return &userCommandRepository{db: db}
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "unique constraint") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}

func (r *userCommandRepository) CreateUser(ctx context.Context, request *requests.CreateUserRequest) (*models.User, error) {
	verified := false
	verifyCode := uuid.New().String()

	user := &models.User{
		Firstname:        request.FirstName,
		Lastname:         request.LastName,
		Email:            request.Email,
		Password:         request.Password,
		VerificationCode: verifyCode,
		IsVerified:       &verified,
	}

	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		if isUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("email already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create user").WithInternal(err)
	}
	return user, nil
}

func (r *userCommandRepository) UpdateUser(ctx context.Context, request *requests.UpdateUserRequest) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("user_id = ?", *request.UserID).First(&user).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "user", "update user")
	}

	err = r.db.WithContext(ctx).Model(&user).Updates(map[string]interface{}{
		"firstname": request.FirstName,
		"lastname":  request.LastName,
		"email":     request.Email,
		"password":  request.Password,
		"updated_at": time.Now(),
	}).Error
	if err != nil {
		if isUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("email already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("update user").WithInternal(err)
	}

	// Re-read to get updated timestamp
	err = r.db.WithContext(ctx).Where("user_id = ?", *request.UserID).First(&user).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("re-read user after update").WithInternal(err)
	}
	return &user, nil
}

func (r *userCommandRepository) UpdateIsVerified(ctx context.Context, userID int, isVerified bool) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "user", "update user")
	}

	err = r.db.WithContext(ctx).Model(&user).Updates(map[string]interface{}{
		"is_verified": isVerified,
		"updated_at":  time.Now(),
	}).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("update user").WithInternal(err)
	}

	err = r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("re-read user after update").WithInternal(err)
	}
	return &user, nil
}

func (r *userCommandRepository) UpdatePassword(ctx context.Context, userID int, password string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "user", "update user")
	}

	err = r.db.WithContext(ctx).Model(&user).Updates(map[string]interface{}{
		"password":   password,
		"updated_at": time.Now(),
	}).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("update user").WithInternal(err)
	}

	err = r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("re-read user after update").WithInternal(err)
	}
	return &user, nil
}

func (r *userCommandRepository) TrashedUser(ctx context.Context, id int) (*models.User, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("user_id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return nil, sharedErrors.ErrFailed("trash user").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, sharedErrors.ErrNotFoundResponse("user")
	}

	var user models.User
	if err := r.db.WithContext(ctx).Where("user_id = ?", id).First(&user).Error; err != nil {
		return nil, sharedErrors.ErrFailed("get trashed user").WithInternal(err)
	}
	return &user, nil
}

func (r *userCommandRepository) RestoreUser(ctx context.Context, id int) (*models.User, error) {
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("user_id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil)
	if result.Error != nil {
		return nil, sharedErrors.ErrFailed("restore user").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, sharedErrors.ErrNotFoundResponse("user")
	}

	var user models.User
	if err := r.db.WithContext(ctx).Where("user_id = ?", id).First(&user).Error; err != nil {
		return nil, sharedErrors.ErrFailed("get restored user").WithInternal(err)
	}
	return &user, nil
}

func (r *userCommandRepository) DeleteUserPermanent(ctx context.Context, id int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("user_id = ?", id).Delete(&models.User{})
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("delete user permanently").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return false, sharedErrors.ErrNotFoundResponse("user")
	}
	return true, nil
}

func (r *userCommandRepository) RestoreAllUser(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("deleted_at IS NOT NULL").
		Update("deleted_at", nil)
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("restore all users").WithInternal(result.Error)
	}
	return true, nil
}

func (r *userCommandRepository) DeleteAllUserPermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.User{})
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("delete all users permanently").WithInternal(result.Error)
	}
	return true, nil
}
