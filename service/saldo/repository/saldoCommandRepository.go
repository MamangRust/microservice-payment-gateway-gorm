package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type saldoCommandRepository struct {
	db *gorm.DB
}

func NewSaldoCommandRepository(db *gorm.DB) SaldoCommandRepository {
	return &saldoCommandRepository{db: db}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r *saldoCommandRepository) CreateSaldo(ctx context.Context, request *requests.CreateSaldoRequest) (*SaldoMutationResult, error) {
	balance := amountToInt64(request.TotalBalance)

	// Try to create or find existing saldo
	var saldo models.Saldo
	err := r.db.WithContext(ctx).
		Where("card_number = ?", request.CardNumber).
		First(&saldo).Error

	if err == nil {
		// Already exists
		return &SaldoMutationResult{
			SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
		}, nil
	}

	// Create new
	saldo = models.Saldo{
		CardNumber:   request.CardNumber,
		TotalBalance: balance,
	}

	if err := r.db.WithContext(ctx).Create(&saldo).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return nil, sharedErrors.NewConflictError("saldo record already exists").WithInternal(err)
		}
		return nil, sharedErrors.ErrFailed("create saldo record").WithInternal(err)
	}

	return &SaldoMutationResult{
		SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
	}, nil
}

func (r *saldoCommandRepository) CreateSaldoIfNotExists(ctx context.Context, request *requests.CreateSaldoRequest) error {
	balance := amountToInt64(request.TotalBalance)
	result := r.db.WithContext(ctx).
		Where("card_number = ?", request.CardNumber).
		FirstOrCreate(&models.Saldo{
			CardNumber:   request.CardNumber,
			TotalBalance: balance,
		})
	return result.Error
}

func (r *saldoCommandRepository) UpdateSaldo(ctx context.Context, request *requests.UpdateSaldoRequest) (*SaldoMutationResult, error) {
	if request.SaldoID == nil {
		return nil, sharedErrors.NewBadRequestError("saldo ID is required")
	}

	var saldo models.Saldo
	err := r.db.WithContext(ctx).Where("saldo_id = ?", *request.SaldoID).First(&saldo).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "update saldo record")
	}

	if saldo.CardNumber != request.CardNumber {
		return nil, sharedErrors.NewConflictError("cannot change card number of an existing saldo")
	}

	target := amountToInt64(request.TotalBalance)
	delta := target - saldo.TotalBalance
	if delta == 0 {
		return &SaldoMutationResult{
			SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
		}, nil
	}

	// Apply balance update directly
	saldo.TotalBalance = target
	if err := r.db.WithContext(ctx).Save(&saldo).Error; err != nil {
		return nil, sharedErrors.ErrFailed("update saldo record").WithInternal(err)
	}

	return &SaldoMutationResult{
		SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
	}, nil
}

func (r *saldoCommandRepository) UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*SaldoMutationResult, error) {
	balance := amountToInt64(request.TotalBalance)
	var saldo models.Saldo
	err := r.db.WithContext(ctx).Where("card_number = ?", request.CardNumber).First(&saldo).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "update saldo balance")
	}
	saldo.TotalBalance = balance
	if err := r.db.WithContext(ctx).Save(&saldo).Error; err != nil {
		return nil, sharedErrors.ErrFailed("update saldo balance").WithInternal(err)
	}
	return &SaldoMutationResult{
		SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
	}, nil
}

func (r *saldoCommandRepository) DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*SaldoMutationResult, error) {
	amount := amountToInt64(request.Amount)

	operationID := request.OperationID
	if operationID == "" {
		operationID = uuid.NewString()
	}
	sourceType := request.SourceType
	if sourceType == "" {
		sourceType = "unknown"
	}

	var result *SaldoMutationResult
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Atomic debit with row-level lock inside explicit transaction
		var saldo models.Saldo
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("card_number = ?", request.CardNumber).
			First(&saldo).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return sharedErrors.NewConflictError("saldo not found")
			}
			return sharedErrors.ErrFailed("debit saldo").WithInternal(err)
		}

		if saldo.TotalBalance < amount {
			return sharedErrors.NewConflictError("insufficient balance")
		}

		balanceBefore := saldo.TotalBalance
		saldo.TotalBalance -= amount
		if err := tx.Save(&saldo).Error; err != nil {
			return sharedErrors.ErrFailed("debit saldo").WithInternal(err)
		}

		// Record ledger entry
		ledger := models.BalanceLedger{
			OperationID:   operationID,
			CardNumber:    request.CardNumber,
			Direction:     "debit",
			Amount:        amount,
			Delta:         -amount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  saldo.TotalBalance,
			SourceType:    sourceType,
			SourceID:      strPtr(request.SourceID),
		}
		if err := tx.Create(&ledger).Error; err != nil {
			return sharedErrors.ErrFailed("debit saldo ledger").WithInternal(err)
		}

		result = &SaldoMutationResult{
			SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
		}
		return nil
	})
	return result, err
}

func (r *saldoCommandRepository) CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*SaldoMutationResult, error) {
	amount := amountToInt64(request.Amount)

	operationID := request.OperationID
	if operationID == "" {
		operationID = uuid.NewString()
	}
	sourceType := request.SourceType
	if sourceType == "" {
		sourceType = "unknown"
	}

	var result *SaldoMutationResult
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var saldo models.Saldo
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("card_number = ?", request.CardNumber).
			First(&saldo).Error
		if err != nil {
			return sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "credit saldo")
		}

		balanceBefore := saldo.TotalBalance
		saldo.TotalBalance += amount
		if err := tx.Save(&saldo).Error; err != nil {
			return sharedErrors.ErrFailed("credit saldo").WithInternal(err)
		}

		ledger := models.BalanceLedger{
			OperationID:   operationID,
			CardNumber:    request.CardNumber,
			Direction:     "credit",
			Amount:        amount,
			Delta:         amount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  saldo.TotalBalance,
			SourceType:    sourceType,
			SourceID:      strPtr(request.SourceID),
		}
		if err := tx.Create(&ledger).Error; err != nil {
			return sharedErrors.ErrFailed("credit saldo ledger").WithInternal(err)
		}

		result = &SaldoMutationResult{
			SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
		}
		return nil
	})
	return result, err
}

func (r *saldoCommandRepository) ApplySaldoAdjustment(ctx context.Context, request *requests.ApplySaldoAdjustmentRequest) (*SaldoAdjustmentResult, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	var saldo models.Saldo
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("card_number = ?", request.CardNumber).First(&saldo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sharedErrors.NewConflictError("saldo adjustment rejected")
		}
		return nil, sharedErrors.ErrFailed("apply saldo adjustment").WithInternal(err)
	}

	balanceBefore := saldo.TotalBalance
	saldo.TotalBalance += request.Delta
	if err := r.db.WithContext(ctx).Save(&saldo).Error; err != nil {
		return nil, sharedErrors.ErrFailed("apply saldo adjustment").WithInternal(err)
	}

	direction := "credit"
	delta := request.Delta
	if delta < 0 {
		direction = "debit"
		delta = -delta
	}

	ledger := models.BalanceLedger{
		OperationID:   request.OperationID,
		CardNumber:    request.CardNumber,
		Direction:     direction,
		Amount:        delta,
		Delta:         request.Delta,
		BalanceBefore: balanceBefore,
		BalanceAfter:  saldo.TotalBalance,
		SourceType:    request.SourceType,
		SourceID:      strPtr(request.SourceID),
	}
	if err := r.db.WithContext(ctx).Create(&ledger).Error; err != nil {
		return nil, sharedErrors.ErrFailed("apply saldo adjustment ledger").WithInternal(err)
	}

	return &SaldoAdjustmentResult{
		SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
	}, nil
}

func (r *saldoCommandRepository) ResolveReconciliation(ctx context.Context, queueID int64, operationID, note string) error {
	if queueID <= 0 || operationID == "" {
		return sharedErrors.NewBadRequestError("queue ID and operation ID are required")
	}
	err := r.db.WithContext(ctx).Model(&models.ReconciliationQueue{}).
		Where("queue_id = ?", queueID).
		Updates(map[string]interface{}{
			"status":                  "resolved",
			"resolved_at":             gorm.Expr("NOW()"),
			"resolution_operation_id": operationID,
			"resolution_note":         note,
		}).Error
	if err != nil {
		return sharedErrors.ErrNoRowsOrFailed(err, "reconciliation queue item", "resolve reconciliation")
	}
	return nil
}

func (r *saldoCommandRepository) UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*SaldoMutationResult, error) {
	var saldo models.Saldo
	err := r.db.WithContext(ctx).Where("card_number = ?", request.CardNumber).First(&saldo).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "update saldo withdrawal metadata")
	}

	if request.WithdrawAmount != nil && request.WithdrawTime != nil {
		withdrawAmount := int64(*request.WithdrawAmount)
		saldo.WithdrawAmount = &withdrawAmount
		saldo.WithdrawTime = request.WithdrawTime
	}

	if err := r.db.WithContext(ctx).Save(&saldo).Error; err != nil {
		return nil, sharedErrors.ErrFailed("update saldo withdrawal metadata").WithInternal(err)
	}

	return &SaldoMutationResult{
		SaldoID: saldo.SaldoID, CardNumber: saldo.CardNumber, TotalBalance: saldo.TotalBalance,
	}, nil
}

func (r *saldoCommandRepository) TrashedSaldo(ctx context.Context, saldo_id int) (*models.Saldo, error) {
	var saldo models.Saldo
	err := r.db.WithContext(ctx).Where("saldo_id = ?", saldo_id).First(&saldo).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "trash saldo record")
	}
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&saldo).Update("deleted_at", now).Error; err != nil {
		return nil, sharedErrors.ErrFailed("trash saldo record").WithInternal(err)
	}
	saldo.DeletedAt = &now
	return &saldo, nil
}

func (r *saldoCommandRepository) RestoreSaldo(ctx context.Context, saldo_id int) (*models.Saldo, error) {
	var saldo models.Saldo
	err := r.db.WithContext(ctx).Unscoped().Where("saldo_id = ? AND deleted_at IS NOT NULL", saldo_id).First(&saldo).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "saldo record", "restore saldo record")
	}
	if err := r.db.WithContext(ctx).Unscoped().Model(&saldo).Update("deleted_at", nil).Error; err != nil {
		return nil, sharedErrors.ErrFailed("restore saldo record").WithInternal(err)
	}
	return &saldo, nil
}

func (r *saldoCommandRepository) DeleteSaldoPermanent(ctx context.Context, saldo_id int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("saldo_id = ?", saldo_id).Delete(&models.Saldo{})
	if result.Error != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(result.Error, "saldo record", "delete saldo record permanently")
	}
	if result.RowsAffected == 0 {
		return false, sharedErrors.ErrNotFoundResponse("saldo record")
	}
	return true, nil
}

func (r *saldoCommandRepository) RestoreAllSaldo(ctx context.Context) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Model(&models.Saldo{}).Where("deleted_at IS NOT NULL").Update("deleted_at", nil).Error; err != nil {
		return false, sharedErrors.ErrFailed("restore all saldo records").WithInternal(err)
	}
	return true, nil
}

func (r *saldoCommandRepository) DeleteAllSaldoPermanent(ctx context.Context) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Saldo{}).Error; err != nil {
		return false, sharedErrors.ErrFailed(fmt.Sprintf("delete all saldo records permanently")).WithInternal(err)
	}
	return true, nil
}
