package repository

import (
	"context"
	"errors"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type merchantQueryRepository struct {
	db *gorm.DB
}

func NewMerchantQueryRepository(db *gorm.DB) MerchantQueryRepository {
	return &merchantQueryRepository{db: db}
}

func (r *merchantQueryRepository) FindAllMerchants(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, error) {
	offset := (req.Page - 1) * req.PageSize
	var merchants []*models.Merchant
	query := r.db.WithContext(ctx).Model(&models.Merchant{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		query = query.Where("name ILIKE ?", "%"+req.Search+"%")
	}
	err := query.Offset(offset).Limit(req.PageSize).Order("merchant_id ASC").Find(&merchants).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("find all merchants").WithInternal(err)
	}
	return merchants, nil
}

func (r *merchantQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, error) {
	offset := (req.Page - 1) * req.PageSize
	var merchants []*models.Merchant
	query := r.db.WithContext(ctx).Model(&models.Merchant{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		query = query.Where("name ILIKE ?", "%"+req.Search+"%")
	}
	err := query.Offset(offset).Limit(req.PageSize).Order("merchant_id ASC").Find(&merchants).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("find active merchants").WithInternal(err)
	}
	return merchants, nil
}

func (r *merchantQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*models.Merchant, error) {
	offset := (req.Page - 1) * req.PageSize
	var merchants []*models.Merchant
	query := r.db.WithContext(ctx).Model(&models.Merchant{}).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		query = query.Where("name ILIKE ?", "%"+req.Search+"%")
	}
	err := query.Offset(offset).Limit(req.PageSize).Order("merchant_id ASC").Find(&merchants).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("find trashed merchants").WithInternal(err)
	}
	return merchants, nil
}

func (r *merchantQueryRepository) FindByMerchantId(ctx context.Context, id int) (*models.Merchant, error) {
	var merchant models.Merchant
	err := r.db.WithContext(ctx).Where("merchant_id = ?", id).First(&merchant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantQueryRepository) FindByApiKey(ctx context.Context, apiKey string) (*models.Merchant, error) {
	var merchant models.Merchant
	err := r.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&merchant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantQueryRepository) FindByName(ctx context.Context, name string) (*models.Merchant, error) {
	var merchant models.Merchant
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&merchant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.ErrNotFoundResponse("merchant").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantQueryRepository) FindByMerchantUserId(ctx context.Context, userID int) ([]*models.Merchant, error) {
	var merchants []*models.Merchant
	err := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).Find(&merchants).Error
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return merchants, nil
}

func (r *merchantQueryRepository) CountAllMerchants(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Merchant{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *merchantQueryRepository) CountActiveMerchants(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Merchant{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *merchantQueryRepository) CountTrashedMerchants(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Merchant{}).Where("deleted_at IS NOT NULL")
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	err := query.Count(&count).Error
	return count, err
}
