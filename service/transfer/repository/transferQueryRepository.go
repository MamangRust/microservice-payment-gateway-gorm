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

type transferQueryRepository struct {
	db *gorm.DB
}

func NewTransferQueryRepository(db *gorm.DB) TransferQueryRepository {
	return &transferQueryRepository{db: db}
}

func (r *transferQueryRepository) FindAll(ctx context.Context, req *requests.FindAllTransfers) ([]*TransferQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Transfer{})
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where(
			"transfer_from LIKE ? OR transfer_to LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count transfers").WithInternal(err)
	}

	var transfers []*models.Transfer
	query2 := r.db.WithContext(ctx)
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where(
			"transfer_from LIKE ? OR transfer_to LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query2.Order("transfer_id ASC").Offset(offset).Limit(req.PageSize).Find(&transfers).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find all transfers").WithInternal(err)
	}

	results := make([]*TransferQueryResult, len(transfers))
	for i, t := range transfers {
		results[i] = &TransferQueryResult{Transfer: t, TotalCount: totalCount}
	}
	return results, nil
}

func (r *transferQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllTransfers) ([]*TransferQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Transfer{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where(
			"transfer_from LIKE ? OR transfer_to LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count active transfers").WithInternal(err)
	}

	var transfers []*models.Transfer
	query2 := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where(
			"transfer_from LIKE ? OR transfer_to LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query2.Order("transfer_id ASC").Offset(offset).Limit(req.PageSize).Find(&transfers).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find active transfers").WithInternal(err)
	}

	results := make([]*TransferQueryResult, len(transfers))
	for i, t := range transfers {
		results[i] = &TransferQueryResult{Transfer: t, TotalCount: totalCount}
	}
	return results, nil
}

func (r *transferQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllTransfers) ([]*TransferQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Transfer{}).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where(
			"transfer_from LIKE ? OR transfer_to LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count trashed transfers").WithInternal(err)
	}

	var transfers []*models.Transfer
	query2 := r.db.WithContext(ctx).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where(
			"transfer_from LIKE ? OR transfer_to LIKE ? OR status LIKE ?",
			like, like, like,
		)
	}
	if err := query2.Order("transfer_id ASC").Offset(offset).Limit(req.PageSize).Find(&transfers).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find trashed transfers").WithInternal(err)
	}

	results := make([]*TransferQueryResult, len(transfers))
	for i, t := range transfers {
		results[i] = &TransferQueryResult{Transfer: t, TotalCount: totalCount}
	}
	return results, nil
}

func (r *transferQueryRepository) FindById(ctx context.Context, id int) (*models.Transfer, error) {
	var transfer models.Transfer
	err := r.db.WithContext(ctx).Where("transfer_id = ?", id).First(&transfer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.ErrNotFoundResponse("transfer").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &transfer, nil
}

func (r *transferQueryRepository) FindTransferByTransferFrom(ctx context.Context, transfer_from string) ([]*models.Transfer, error) {
	var transfers []*models.Transfer
	err := r.db.WithContext(ctx).Where("transfer_from = ?", transfer_from).Find(&transfers).Error
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return transfers, nil
}

func (r *transferQueryRepository) FindTransferByTransferTo(ctx context.Context, transfer_to string) ([]*models.Transfer, error) {
	var transfers []*models.Transfer
	err := r.db.WithContext(ctx).Where("transfer_to = ?", transfer_to).Find(&transfers).Error
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return transfers, nil
}
