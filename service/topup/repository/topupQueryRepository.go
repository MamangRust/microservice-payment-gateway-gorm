package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type topupQueryRepository struct {
	db *gorm.DB
}

func NewTopupQueryRepository(db *gorm.DB) TopupQueryRepository {
	return &topupQueryRepository{db: db}
}

func (r *topupQueryRepository) FindAllTopups(ctx context.Context, req *requests.FindAllTopups) ([]*TopupQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Topup{})
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where(
			"card_number LIKE ? OR topup_method LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count topups").WithInternal(err)
	}

	var topups []*models.Topup
	query2 := r.db.WithContext(ctx)
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where(
			"card_number LIKE ? OR topup_method LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query2.Order("topup_time DESC").Offset(offset).Limit(req.PageSize).Find(&topups).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find all topups").WithInternal(err)
	}

	results := make([]*TopupQueryResult, len(topups))
	for i, t := range topups {
		results[i] = &TopupQueryResult{Topup: t, TotalCount: totalCount}
	}
	return results, nil
}

func (r *topupQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllTopups) ([]*TopupQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Topup{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where(
			"card_number LIKE ? OR topup_method LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count active topups").WithInternal(err)
	}

	var topups []*models.Topup
	query2 := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where(
			"card_number LIKE ? OR topup_method LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query2.Order("topup_time DESC").Offset(offset).Limit(req.PageSize).Find(&topups).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find active topups").WithInternal(err)
	}

	results := make([]*TopupQueryResult, len(topups))
	for i, t := range topups {
		results[i] = &TopupQueryResult{Topup: t, TotalCount: totalCount}
	}
	return results, nil
}

func (r *topupQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllTopups) ([]*TopupQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Topup{}).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where(
			"card_number LIKE ? OR topup_method LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count trashed topups").WithInternal(err)
	}

	var topups []*models.Topup
	query2 := r.db.WithContext(ctx).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where(
			"card_number LIKE ? OR topup_method LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query2.Order("topup_time DESC").Offset(offset).Limit(req.PageSize).Find(&topups).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find trashed topups").WithInternal(err)
	}

	results := make([]*TopupQueryResult, len(topups))
	for i, t := range topups {
		results[i] = &TopupQueryResult{Topup: t, TotalCount: totalCount}
	}
	return results, nil
}

func (r *topupQueryRepository) FindAllTopupByCardNumber(ctx context.Context, req *requests.FindAllTopupsByCardNumber) ([]*TopupByCardResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Topup{}).Where("card_number = ?", req.CardNumber)
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where("topup_method LIKE ? OR status LIKE ?", like, like)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count topups by card number").WithInternal(err)
	}

	var topups []*models.Topup
	query2 := r.db.WithContext(ctx).Where("card_number = ?", req.CardNumber)
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where("topup_method LIKE ? OR status LIKE ?", like, like)
	}
	if err := query2.Order("topup_time DESC").Offset(offset).Limit(req.PageSize).Find(&topups).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find topups by card number").WithInternal(err)
	}

	results := make([]*TopupByCardResult, len(topups))
	for i, t := range topups {
		results[i] = &TopupByCardResult{Topup: t, TotalCount: totalCount}
	}
	return results, nil
}

func (r *topupQueryRepository) FindById(ctx context.Context, topup_id int) (*models.Topup, error) {
	var topup models.Topup
	err := r.db.WithContext(ctx).First(&topup, int32(topup_id)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.ErrNotFoundResponse("topup").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &topup, nil
}
