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

type withdrawQueryRepository struct {
	db *gorm.DB
}

func NewWithdrawQueryRepository(db *gorm.DB) WithdrawQueryRepository {
	return &withdrawQueryRepository{db: db}
}

func (r *withdrawQueryRepository) FindAll(ctx context.Context, req *requests.FindAllWithdraws) ([]*WithdrawQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Withdraw{})
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where("card_number LIKE ? OR status LIKE ?", like, like)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count withdraws").WithInternal(err)
	}

	var withdraws []*models.Withdraw
	query2 := r.db.WithContext(ctx)
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where("card_number LIKE ? OR status LIKE ?", like, like)
	}
	if err := query2.Order("withdraw_id ASC").Offset(offset).Limit(req.PageSize).Find(&withdraws).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find all withdraws").WithInternal(err)
	}

	results := make([]*WithdrawQueryResult, len(withdraws))
	for i, w := range withdraws {
		results[i] = &WithdrawQueryResult{Withdraw: w, TotalCount: totalCount}
	}
	return results, nil
}

func (r *withdrawQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllWithdraws) ([]*WithdrawQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Withdraw{}).Where("deleted_at IS NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where("card_number LIKE ? OR status LIKE ?", like, like)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count active withdraws").WithInternal(err)
	}

	var withdraws []*models.Withdraw
	query2 := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where("card_number LIKE ? OR status LIKE ?", like, like)
	}
	if err := query2.Order("withdraw_id ASC").Offset(offset).Limit(req.PageSize).Find(&withdraws).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find active withdraws").WithInternal(err)
	}

	results := make([]*WithdrawQueryResult, len(withdraws))
	for i, w := range withdraws {
		results[i] = &WithdrawQueryResult{Withdraw: w, TotalCount: totalCount}
	}
	return results, nil
}

func (r *withdrawQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllWithdraws) ([]*WithdrawQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Withdraw{}).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where("card_number LIKE ? OR status LIKE ?", like, like)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count trashed withdraws").WithInternal(err)
	}

	var withdraws []*models.Withdraw
	query2 := r.db.WithContext(ctx).Where("deleted_at IS NOT NULL")
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where("card_number LIKE ? OR status LIKE ?", like, like)
	}
	if err := query2.Order("withdraw_id ASC").Offset(offset).Limit(req.PageSize).Find(&withdraws).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find trashed withdraws").WithInternal(err)
	}

	results := make([]*WithdrawQueryResult, len(withdraws))
	for i, w := range withdraws {
		results[i] = &WithdrawQueryResult{Withdraw: w, TotalCount: totalCount}
	}
	return results, nil
}

func (r *withdrawQueryRepository) FindAllByCardNumber(ctx context.Context, req *requests.FindAllWithdrawCardNumber) ([]*WithdrawQueryResult, error) {
	offset := (req.Page - 1) * req.PageSize

	var totalCount int64
	query := r.db.WithContext(ctx).Model(&models.Withdraw{}).Where("card_number = ?", req.CardNumber)
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query = query.Where("status LIKE ?", like)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, sharedErrors.ErrFailed("count withdraws by card number").WithInternal(err)
	}

	var withdraws []*models.Withdraw
	query2 := r.db.WithContext(ctx).Where("card_number = ?", req.CardNumber)
	if req.Search != "" {
		like := fmt.Sprintf("%%%s%%", req.Search)
		query2 = query2.Where("status LIKE ?", like)
	}
	if err := query2.Order("withdraw_id ASC").Offset(offset).Limit(req.PageSize).Find(&withdraws).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find withdraws by card number").WithInternal(err)
	}

	results := make([]*WithdrawQueryResult, len(withdraws))
	for i, w := range withdraws {
		results[i] = &WithdrawQueryResult{Withdraw: w, TotalCount: totalCount}
	}
	return results, nil
}

func (r *withdrawQueryRepository) FindById(ctx context.Context, id int) (*models.Withdraw, error) {
	var withdraw models.Withdraw
	err := r.db.WithContext(ctx).Where("withdraw_id = ?", id).First(&withdraw).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.ErrNotFoundResponse("withdraw").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &withdraw, nil
}

func (r *withdrawQueryRepository) GetTodayWithdrawSumByCardNumber(ctx context.Context, cardNumber string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.Withdraw{}).
		// withdraw_time is stored as the app's wall clock (TimeZone=Asia/Jakarta in
		// the DSN) without a timezone; compare against the current Jakarta date
		// explicitly so the daily window is consistent regardless of the DB
		// session timezone (e.g. UTC in test containers).
		Where("card_number = ? AND status = 'success' AND DATE(withdraw_time) = (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta')::date", cardNumber).
		Select("COALESCE(SUM(withdraw_amount), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, sharedErrors.ErrFailed("get today withdraw sum").WithInternal(err)
	}
	return total, nil
}
