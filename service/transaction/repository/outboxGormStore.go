package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

type OutboxGormStore interface {
	Insert(ctx context.Context, record OutboxRecord) error
}

type outboxGormStore struct {
	db *gorm.DB
}

func NewOutboxGormStore(db *gorm.DB) OutboxGormStore {
	return &outboxGormStore{db: db}
}

func (s *outboxGormStore) Insert(ctx context.Context, record OutboxRecord) error {
	if record.EventID == "" {
		record.EventID = uuid.New().String()
	}
	if record.EventVersion == 0 {
		record.EventVersion = 1
	}
	if record.Status == "" {
		record.Status = "pending"
	}
	if record.AvailableAt.IsZero() {
		record.AvailableAt = time.Now()
	}

	return s.db.WithContext(ctx).Create(&record).Error
}
