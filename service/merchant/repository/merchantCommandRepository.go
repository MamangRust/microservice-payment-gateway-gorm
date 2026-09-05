package repository

import (
	"context"
	"time"

	apikey "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/api-key"
	database "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type merchantCommandRepository struct {
	db *gorm.DB
}

func NewMerchantCommandRepository(db *gorm.DB) MerchantCommandRepository {
	return &merchantCommandRepository{db: db}
}

func (r *merchantCommandRepository) CreateMerchant(ctx context.Context, request *requests.CreateMerchantRequest) (*models.Merchant, error) {
	apiKey, err := apikey.GenerateApiKey()
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	merchant := &models.Merchant{
		Name:   request.Name,
		ApiKey: apiKey,
		UserID: int32(request.UserID),
		Status: "inactive",
	}

	res := r.db.WithContext(ctx).Create(merchant)
	if res.Error != nil {
		if database.IsUniqueViolation(res.Error) {
			return nil, sharedErrors.NewConflictError("merchant api key already exists").WithInternal(res.Error)
		}
		return nil, sharedErrors.ErrFailed("create merchant").WithInternal(res.Error)
	}

	return merchant, nil
}

func (r *merchantCommandRepository) UpdateMerchant(ctx context.Context, request *requests.UpdateMerchantRequest) (*models.Merchant, error) {
	var merchant models.Merchant
	err := r.db.WithContext(ctx).Where("merchant_id = ? AND deleted_at IS NULL", *request.MerchantID).First(&merchant).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant", "update merchant")
	}

	merchant.Name = request.Name
	merchant.UserID = int32(request.UserID)
	merchant.Status = request.Status

	if err := r.db.WithContext(ctx).Save(&merchant).Error; err != nil {
		return nil, sharedErrors.ErrFailed("update merchant").WithInternal(err)
	}

	return &merchant, nil
}

func (r *merchantCommandRepository) UpdateMerchantStatus(ctx context.Context, request *requests.UpdateMerchantStatusRequest) (*models.Merchant, error) {
	var merchant models.Merchant
	err := r.db.WithContext(ctx).Where("merchant_id = ? AND deleted_at IS NULL", *request.MerchantID).First(&merchant).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant", "update merchant status")
	}

	merchant.Status = request.Status

	if err := r.db.WithContext(ctx).Save(&merchant).Error; err != nil {
		return nil, sharedErrors.ErrFailed("update merchant status").WithInternal(err)
	}

	return &merchant, nil
}

func (r *merchantCommandRepository) TrashedMerchant(ctx context.Context, merchantID int) (*models.Merchant, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.Merchant{}).
		Where("merchant_id = ? AND deleted_at IS NULL", merchantID).
		Update("deleted_at", now)
	if result.Error != nil {
		return nil, sharedErrors.ErrFailed("trash merchant").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, sharedErrors.ErrNotFoundResponse("merchant")
	}

	var merchant models.Merchant
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&merchant).Error; err != nil {
		return nil, sharedErrors.ErrFailed("get trashed merchant").WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) RestoreMerchant(ctx context.Context, merchantID int) (*models.Merchant, error) {
	result := r.db.WithContext(ctx).Unscoped().Model(&models.Merchant{}).
		Where("merchant_id = ? AND deleted_at IS NOT NULL", merchantID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return nil, sharedErrors.ErrFailed("restore merchant").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, sharedErrors.ErrNotFoundResponse("merchant")
	}

	var merchant models.Merchant
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&merchant).Error; err != nil {
		return nil, sharedErrors.ErrFailed("get restored merchant").WithInternal(err)
	}
	return &merchant, nil
}

func (r *merchantCommandRepository) DeleteMerchantPermanent(ctx context.Context, merchantID int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("merchant_id = ?", merchantID).Delete(&models.Merchant{})
	if result.Error != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(result.Error, "merchant", "delete merchant permanently")
	}
	if result.RowsAffected == 0 {
		return false, sharedErrors.ErrNotFoundResponse("merchant")
	}
	return true, nil
}

func (r *merchantCommandRepository) RestoreAllMerchant(ctx context.Context) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Model(&models.Merchant{}).Where("deleted_at IS NOT NULL").Update("deleted_at", nil).Error; err != nil {
		return false, sharedErrors.ErrFailed("restore all merchants").WithInternal(err)
	}
	return true, nil
}

func (r *merchantCommandRepository) DeleteAllMerchantPermanent(ctx context.Context) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Merchant{}).Error; err != nil {
		return false, sharedErrors.ErrFailed("delete all merchants permanently").WithInternal(err)
	}
	return true, nil
}
