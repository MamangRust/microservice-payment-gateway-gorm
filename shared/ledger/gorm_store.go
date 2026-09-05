package ledger

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ReconciliationQueueRow maps to the reconciliation_queue table.
type ReconciliationQueueRow struct {
	QueueID               int64      `gorm:"column:queue_id;primaryKey;autoIncrement" json:"queue_id"`
	SaldoID               int32      `gorm:"column:saldo_id" json:"saldo_id"`
	CardNumber            string     `gorm:"column:card_number" json:"card_number"`
	CurrentBalance        int64      `gorm:"column:current_balance" json:"current_balance"`
	LedgerBalance         int64      `gorm:"column:ledger_balance" json:"ledger_balance"`
	Difference            int64      `gorm:"column:difference" json:"difference"`
	LedgerEntries         int64      `gorm:"column:ledger_entries" json:"ledger_entries"`
	Status                string     `gorm:"column:status;default:pending" json:"status"`
	MismatchCount         int64      `gorm:"column:mismatch_count;default:1" json:"mismatch_count"`
	FirstSeenAt           *time.Time `gorm:"column:first_seen_at" json:"first_seen_at"`
	LastSeenAt            *time.Time `gorm:"column:last_seen_at" json:"last_seen_at"`
	ResolvedAt            *time.Time `gorm:"column:resolved_at" json:"resolved_at"`
	ResolutionOperationID *string    `gorm:"column:resolution_operation_id" json:"resolution_operation_id"`
	ResolutionNote        *string    `gorm:"column:resolution_note" json:"resolution_note"`
}

func (ReconciliationQueueRow) TableName() string { return "reconciliation_queue" }

// GormReconciliationStore implements ReconciliationStore and reconciliationQueueStore
// using GORM.
type GormReconciliationStore struct {
	db *gorm.DB
}

// NewGormReconciliationStore creates a new GORM-backed reconciliation store.
func NewGormReconciliationStore(db *gorm.DB) *GormReconciliationStore {
	return &GormReconciliationStore{db: db}
}

// ListReconciliationMismatches returns accounts whose current balance
// differs from the immutable ledger total.
func (s *GormReconciliationStore) ListReconciliationMismatches(ctx context.Context, limit int32) ([]*ReconciliationRow, error) {
	var rows []*ReconciliationRow

	result := s.db.WithContext(ctx).Raw(`
		WITH ledger_totals AS (
			SELECT card_number, COALESCE(SUM(delta), 0) AS ledger_balance,
			       COUNT(*) AS ledger_entries
			FROM balance_ledger
			GROUP BY card_number
		)
		SELECT s.saldo_id, s.card_number, s.total_balance,
		       COALESCE(l.ledger_balance, 0),
		       s.total_balance - COALESCE(l.ledger_balance, 0),
		       COALESCE(l.ledger_entries, 0)
		FROM saldos s
		LEFT JOIN ledger_totals l ON l.card_number = s.card_number
		WHERE s.deleted_at IS NULL
		  AND s.total_balance <> COALESCE(l.ledger_balance, 0)
		ORDER BY ABS(s.total_balance - COALESCE(l.ledger_balance, 0)) DESC, s.saldo_id
		LIMIT ?
	`, limit).Scan(&rows)

	return rows, result.Error
}

// EnqueueReconciliationMismatch upserts a mismatch record into the reconciliation_queue table.
// Uses ON CONFLICT to update existing active rows for the same saldo_id.
// Uses raw SQL because GORM's OnConflict doesn't support partial unique index conditions.
func (s *GormReconciliationStore) EnqueueReconciliationMismatch(ctx context.Context, item *ReconciliationRow) error {
	return s.db.WithContext(ctx).Exec(`
		INSERT INTO reconciliation_queue (
			saldo_id, card_number, current_balance, ledger_balance,
			difference, ledger_entries
		) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (saldo_id) WHERE status IN ('pending', 'investigating')
		DO UPDATE SET current_balance = EXCLUDED.current_balance,
		              ledger_balance = EXCLUDED.ledger_balance,
		              difference = EXCLUDED.difference,
		              ledger_entries = EXCLUDED.ledger_entries,
		              mismatch_count = reconciliation_queue.mismatch_count + 1,
		              last_seen_at = current_timestamp
	`, item.SaldoID, item.CardNumber, item.CurrentBalance, item.LedgerBalance, item.Difference, item.LedgerEntries).Error
}
