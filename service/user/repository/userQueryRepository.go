package repository

import (
	"context"
	"errors"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type userQueryRepository struct {
	db *gorm.DB
}

func NewUserQueryRepository(db *gorm.DB) UserQueryRepository {
	return &userQueryRepository{db: db}
}

func (r *userQueryRepository) FindAllUsers(ctx context.Context, req *requests.FindAllUsers) ([]*models.User, error) {
	offset := (req.Page - 1) * req.PageSize

	var users []*models.User
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		query = query.Where("firstname ILIKE ? OR lastname ILIKE ? OR email ILIKE ?",
			"%"+req.Search+"%", "%"+req.Search+"%", "%"+req.Search+"%")
	}

	err := query.Offset(offset).Limit(req.PageSize).Order("user_id ASC").Find(&users).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("find all users").WithInternal(err)
	}
	return users, nil
}

func (r *userQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*models.User, error) {
	offset := (req.Page - 1) * req.PageSize

	var users []*models.User
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		query = query.Where("firstname ILIKE ? OR lastname ILIKE ? OR email ILIKE ?",
			"%"+req.Search+"%", "%"+req.Search+"%", "%"+req.Search+"%")
	}

	err := query.Offset(offset).Limit(req.PageSize).Order("user_id ASC").Find(&users).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("find active users").WithInternal(err)
	}
	return users, nil
}

func (r *userQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*models.User, error) {
	offset := (req.Page - 1) * req.PageSize

	var users []*models.User
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		query = query.Where("firstname ILIKE ? OR lastname ILIKE ? OR email ILIKE ?",
			"%"+req.Search+"%", "%"+req.Search+"%", "%"+req.Search+"%")
	}

	err := query.Offset(offset).Limit(req.PageSize).Order("user_id ASC").Find(&users).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed users").WithInternal(err)
	}
	return users, nil
}

func (r *userQueryRepository) FindById(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("user_id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.ErrNotFound.WithMessage("user not found").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

func (r *userQueryRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.ErrNotFound.WithMessage("user not found").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

func (r *userQueryRepository) FindByVerificationCode(ctx context.Context, code string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("verification_code = ?", code).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.ErrNotFound.WithMessage("user not found").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

func (r *userQueryRepository) CountAllUsers(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("firstname ILIKE ? OR lastname ILIKE ? OR email ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *userQueryRepository) CountActiveUsers(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("firstname ILIKE ? OR lastname ILIKE ? OR email ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *userQueryRepository) CountTrashedUsers(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("deleted_at IS NOT NULL")
	if search != "" {
		query = query.Where("firstname ILIKE ? OR lastname ILIKE ? OR email ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	err := query.Count(&count).Error
	return count, err
}
