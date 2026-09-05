package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OutboxRecord represents a row in the outbox_events table. It is defined here
// (schema-agnostic) so the shared store/publisher work against any service that
// owns the identical outbox_events table.
type OutboxRecord struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AggregateType string     `gorm:"column:aggregate_type" json:"aggregate_type"`
	AggregateID   string     `gorm:"column:aggregate_id" json:"aggregate_id"`
	EventID       string     `gorm:"column:event_id" json:"event_id"`
	EventType     string     `gorm:"column:event_type" json:"event_type"`
	EventVersion  int32      `gorm:"column:event_version" json:"event_version"`
	Payload       []byte     `gorm:"column:payload" json:"payload"`
	Status        string     `gorm:"column:status" json:"status"`
	Attempts      int32      `gorm:"column:attempts" json:"attempts"`
	LastError     *string    `gorm:"column:last_error" json:"last_error"`
	AvailableAt   time.Time  `gorm:"column:available_at" json:"available_at"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"created_at"`
	PublishedAt   *time.Time `gorm:"column:published_at" json:"published_at"`
}

func (OutboxRecord) TableName() string { return "outbox_events" }

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
)

// InsertOutbox inserts a single event into the outbox. The insert runs outside
// any transaction — it is the caller's responsibility to ensure the outbox row
// is inserted atomically within the same DB transaction as the business
// mutation.
func InsertOutbox(ctx context.Context, db *gorm.DB, record OutboxRecord) error {
	if record.EventID == "" {
		record.EventID = uuid.New().String()
	}
	if record.EventVersion == 0 {
		record.EventVersion = 1
	}
	if record.Status == "" {
		record.Status = OutboxStatusPending
	}
	if record.AvailableAt.IsZero() {
		record.AvailableAt = time.Now()
	}

	return db.WithContext(ctx).Create(&record).Error
}

// ClaimPendingOutbox locks up to maxRows pending/retriable outbox rows
// and returns them for publishing. Uses FOR UPDATE SKIP LOCKED so
// multiple publisher instances can safely poll concurrently.
// Stale 'processing' rows (crashed publisher) are also reclaimed after
// a grace period.
//
// Uses raw SQL because GORM cannot express FOR UPDATE SKIP LOCKED on a
// subquery + RETURNING in a portable way.
func ClaimPendingOutbox(ctx context.Context, db *gorm.DB, maxRows int32, maxAttempts int32) ([]*OutboxRecord, error) {
	var records []*OutboxRecord

	result := db.WithContext(ctx).Raw(`
		UPDATE outbox_events
		SET status = 'processing',
			attempts = attempts + 1,
			available_at = NOW() + INTERVAL '5 minutes'
		WHERE id IN (
			SELECT id FROM outbox_events
			WHERE (
			    status IN ('pending', 'failed')
			    OR (status = 'processing' AND available_at <= NOW() - INTERVAL '10 minutes')
			)
			  AND attempts < ?
			  AND available_at <= NOW()
			ORDER BY available_at
			LIMIT ?
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, aggregate_type, aggregate_id, event_id,
			event_type, event_version, payload, status, attempts,
			last_error, available_at, created_at, published_at
	`, maxAttempts, maxRows).Scan(&records)

	return records, result.Error
}

// CompleteOutbox marks an outbox event as successfully published.
func CompleteOutbox(ctx context.Context, db *gorm.DB, eventID string) error {
	return db.WithContext(ctx).
		Model(&OutboxRecord{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":       OutboxStatusPublished,
			"published_at": gorm.Expr("NOW()"),
		}).Error
}

// FailOutbox marks an outbox event as failed with error details.
func FailOutbox(ctx context.Context, db *gorm.DB, eventID string, errMsg string) error {
	return db.WithContext(ctx).
		Model(&OutboxRecord{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":       OutboxStatusFailed,
			"last_error":   errMsg,
			"available_at": gorm.Expr("NOW() + INTERVAL '30 seconds'"),
		}).Error
}

// CountPendingOutbox returns the number of events still awaiting delivery
// (pending or retriable failed). Used to expose outbox lag as a gauge.
func CountPendingOutbox(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64
	err := db.WithContext(ctx).
		Model(&OutboxRecord{}).
		Where("status IN ?", []string{"pending", "failed"}).
		Count(&count).Error
	return count, err
}
