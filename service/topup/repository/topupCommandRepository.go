package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type topupCommandRepository struct {
	db *gorm.DB
}

func NewTopupCommandRepository(db *gorm.DB) TopupCommandRepository {
	return &topupCommandRepository{db: db}
}

func (r *topupCommandRepository) CreateTopup(ctx context.Context, request *requests.CreateTopupRequest) (*models.Topup, error) {
	now := time.Now()
	topup := &models.Topup{
		TopupNo:     uuid.New().String(),
		CardNumber:  request.CardNumber,
		TopupAmount: int64(request.TopupAmount),
		TopupMethod: request.TopupMethod,
		TopupTime:   now,
		Status:      "pending",
		OperationID: uuid.New().String(),
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	if err := r.db.WithContext(ctx).Create(topup).Error; err != nil {
		return nil, sharedErrors.ErrFailed("create topup").WithInternal(err)
	}

	return topup, nil
}

func (r *topupCommandRepository) UpdateTopup(ctx context.Context, request *requests.UpdateTopupRequest) (*models.Topup, error) {
	var topup models.Topup
	if err := r.db.WithContext(ctx).First(&topup, int32(*request.TopupID)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "update topup")
	}

	now := time.Now()
	topup.CardNumber = request.CardNumber
	topup.TopupAmount = int64(request.TopupAmount)
	topup.TopupMethod = request.TopupMethod
	topup.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&topup).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "update topup")
	}

	return &topup, nil
}

func (r *topupCommandRepository) UpdateTopupAmount(ctx context.Context, request *requests.UpdateTopupAmount) (*models.Topup, error) {
	var topup models.Topup
	if err := r.db.WithContext(ctx).First(&topup, int32(request.TopupID)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "update topup amount")
	}

	now := time.Now()
	topup.TopupAmount = int64(request.TopupAmount)
	topup.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&topup).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "update topup amount")
	}

	return &topup, nil
}

func (r *topupCommandRepository) UpdateTopupStatus(ctx context.Context, request *requests.UpdateTopupStatus) (*models.Topup, error) {
	var topup models.Topup
	if err := r.db.WithContext(ctx).First(&topup, int32(request.TopupID)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "update topup status")
	}

	now := time.Now()
	topup.Status = request.Status
	topup.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&topup).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "update topup status")
	}

	return &topup, nil
}

func (r *topupCommandRepository) TrashedTopup(ctx context.Context, topup_id int) (*models.Topup, error) {
	var topup models.Topup
	if err := r.db.WithContext(ctx).First(&topup, int32(topup_id)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "trash topup")
	}

	now := time.Now()
	topup.DeletedAt = &now

	if err := r.db.WithContext(ctx).Save(&topup).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "trash topup")
	}

	return &topup, nil
}

func (r *topupCommandRepository) RestoreTopup(ctx context.Context, topup_id int) (*models.Topup, error) {
	var topup models.Topup
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").First(&topup, int32(topup_id)).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "restore topup")
	}

	topup.DeletedAt = nil

	if err := r.db.WithContext(ctx).Unscoped().Save(&topup).Error; err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "topup", "restore topup")
	}

	return &topup, nil
}

func (r *topupCommandRepository) DeleteTopupPermanent(ctx context.Context, topup_id int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("topup_id = ?", int32(topup_id)).Delete(&models.Topup{})
	if result.Error != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(result.Error, "topup", "delete topup permanently")
	}
	return true, nil
}

func (r *topupCommandRepository) RestoreAllTopup(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Model(&models.Topup{}).Where("deleted_at IS NOT NULL").Update("deleted_at", nil)
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("restore all topups").WithInternal(result.Error)
	}
	return true, nil
}

func (r *topupCommandRepository) DeleteAllTopupPermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Topup{})
	if result.Error != nil {
		return false, sharedErrors.ErrFailed("delete all topups permanently").WithInternal(result.Error)
	}
	return true, nil
}
