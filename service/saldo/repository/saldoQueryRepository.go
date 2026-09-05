package repository

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type saldoQueryRepository struct {
	db *gorm.DB
}

func NewSaldoQueryRepository(db *gorm.DB) SaldoQueryRepository {
	return &saldoQueryRepository{db: db}
}

func (r *saldoQueryRepository) ListReconciliationQueue(ctx context.Context, status string, limit int32) ([]*ReconciliationQueueResult, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	var results []*ReconciliationQueueResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			rq.queue_id,
			s.saldo_id,
			s.card_number,
			s.total_balance AS current_balance,
			COALESCE((
				SELECT SUM(bl.amount * CASE WHEN bl.direction = 'debit' THEN -1 ELSE 1 END) FROM balance_ledger bl WHERE bl.card_number = s.card_number
			), 0) AS ledger_balance,
			ABS(s.total_balance - COALESCE((
				SELECT SUM(bl.amount * CASE WHEN bl.direction = 'debit' THEN -1 ELSE 1 END) FROM balance_ledger bl WHERE bl.card_number = s.card_number
			), 0)) AS difference,
			COALESCE((
				SELECT COUNT(*)::int32 FROM balance_ledger bl WHERE bl.card_number = s.card_number
			), 0) AS ledger_entries,
			rq.status,
			COALESCE((
				SELECT COUNT(*)::int32 FROM reconciliation_queue rq2 
				WHERE rq2.card_number = s.card_number AND rq2.status != 'resolved'
			), 0) AS mismatch_count,
			rq.first_seen_at,
			rq.last_seen_at,
			rq.resolved_at,
			rq.resolution_operation_id,
			rq.resolution_note
		FROM reconciliation_queue rq
		JOIN saldos s ON s.card_number = rq.card_number
		WHERE rq.status = ?
		ORDER BY rq.created_at DESC
		LIMIT ?
	`, status, limit).Scan(&results).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("list reconciliation queue").WithInternal(err)
	}
	return results, nil
}

func (r *saldoQueryRepository) ListLedgerEntries(ctx context.Context, cardNumber string, limit int32) ([]*LedgerEntryResult, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	var results []*LedgerEntryResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			bl.entry_id,
			bl.operation_id,
			bl.card_number,
			bl.direction,
			bl.amount,
			bl.delta,
			bl.balance_before,
			bl.balance_after,
			bl.source_type,
			bl.source_id,
			bl.created_at
		FROM balance_ledger bl
		WHERE bl.card_number = ?
		ORDER BY bl.entry_id DESC
		LIMIT ?
	`, cardNumber, limit).Scan(&results).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("list ledger entries").WithInternal(err)
	}
	return results, nil
}

func (r *saldoQueryRepository) FindAllSaldos(ctx context.Context, req *requests.FindAllSaldos) ([]*SaldoResult, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*SaldoResult
	query := r.db.WithContext(ctx).Table("saldos s").
		Select("s.saldo_id, s.card_number, s.total_balance, s.withdraw_amount, s.withdraw_time, s.created_at, s.updated_at, s.deleted_at")
	if req.Search != "" {
		query = query.Where("s.card_number ILIKE ?", "%"+req.Search+"%")
	}
	if err := query.Offset(offset).Limit(req.PageSize).Order("s.saldo_id ASC").Scan(&results).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find all saldo records").WithInternal(err)
	}
	return results, nil
}

func (r *saldoQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllSaldos) ([]*SaldoResult, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*SaldoResult
	query := r.db.WithContext(ctx).Table("saldos s").
		Select("s.saldo_id, s.card_number, s.total_balance, s.withdraw_amount, s.withdraw_time, s.created_at, s.updated_at, s.deleted_at").
		Where("s.deleted_at IS NULL")
	if req.Search != "" {
		query = query.Where("s.card_number ILIKE ?", "%"+req.Search+"%")
	}
	if err := query.Offset(offset).Limit(req.PageSize).Order("s.saldo_id ASC").Scan(&results).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find active saldo records").WithInternal(err)
	}
	return results, nil
}

func (r *saldoQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllSaldos) ([]*SaldoResult, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*SaldoResult
	query := r.db.WithContext(ctx).Table("saldos s").
		Select("s.saldo_id, s.card_number, s.total_balance, s.withdraw_amount, s.withdraw_time, s.created_at, s.updated_at, s.deleted_at").
		Where("s.deleted_at IS NOT NULL")
	if req.Search != "" {
		query = query.Where("s.card_number ILIKE ?", "%"+req.Search+"%")
	}
	if err := query.Offset(offset).Limit(req.PageSize).Order("s.saldo_id ASC").Scan(&results).Error; err != nil {
		return nil, sharedErrors.ErrFailed("find trashed saldo records").WithInternal(err)
	}
	return results, nil
}

func (r *saldoQueryRepository) FindById(ctx context.Context, saldo_id int) (*SaldoResult, error) {
	var result SaldoResult
	err := r.db.WithContext(ctx).Table("saldos s").
		Select("s.saldo_id, s.card_number, s.total_balance, s.withdraw_amount, s.withdraw_time, s.created_at, s.updated_at, s.deleted_at").
		Where("s.saldo_id = ?", saldo_id).Scan(&result).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sharedErrors.ErrNotFoundResponse("saldo record").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &result, nil
}

func (r *saldoQueryRepository) FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error) {
	var saldo models.Saldo
	err := r.db.WithContext(ctx).Where("card_number = ?", card_number).First(&saldo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sharedErrors.ErrNotFoundResponse("saldo record").WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &saldo, nil
}

func (r *saldoQueryRepository) CountAllSaldos(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Saldo{})
	if search != "" {
		query = query.Where("card_number ILIKE ?", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *saldoQueryRepository) CountActiveSaldos(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Saldo{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("card_number ILIKE ?", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *saldoQueryRepository) CountTrashedSaldos(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Saldo{}).Where("deleted_at IS NOT NULL")
	if search != "" {
		query = query.Where("card_number ILIKE ?", "%"+search+"%")
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
