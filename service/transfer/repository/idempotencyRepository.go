package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/idempotency"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type transferIdempotencyRecord struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Scope           string    `gorm:"column:scope"`
	IdempotencyKey  string    `gorm:"column:idempotency_key"`
	RequestHash     string    `gorm:"column:request_hash"`
	OperationID     *string   `gorm:"column:operation_id"`
	Status          string    `gorm:"column:status"`
	ResponsePayload []byte    `gorm:"column:response_payload"`
	ResourceID      *int32    `gorm:"column:resource_id"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
	ExpiresAt       time.Time `gorm:"column:expires_at"`
}

func (transferIdempotencyRecord) TableName() string { return "idempotency_records" }

type TransferIdempotencyRepository struct {
	db *gorm.DB
}

func NewTransferIdempotencyRepository(db *gorm.DB) *TransferIdempotencyRepository {
	return &TransferIdempotencyRepository{db: db}
}

func (r *TransferIdempotencyRepository) Claim(ctx context.Context, scope, key, requestHash string) (*idempotency.Record, error) {
	operationID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	rec := &transferIdempotencyRecord{
		Scope:          scope,
		IdempotencyKey: key,
		RequestHash:    requestHash,
		OperationID:    &operationID,
		Status:         "processing",
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      expiresAt,
	}

	err := r.db.WithContext(ctx).Create(rec).Error
	if err != nil {
		return nil, idempotency.ErrKeyInUse
	}

	return mapTransferIdempotencyRecord(rec), nil
}

func (r *TransferIdempotencyRepository) Get(ctx context.Context, scope, key string) (*idempotency.Record, error) {
	var rec transferIdempotencyRecord
	err := r.db.WithContext(ctx).Where("scope = ? AND idempotency_key = ?", scope, key).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return mapTransferIdempotencyRecord(&rec), nil
}

func (r *TransferIdempotencyRepository) Complete(ctx context.Context, scope, key, requestHash string, outcome idempotency.Outcome) error {
	return r.db.WithContext(ctx).Model(&transferIdempotencyRecord{}).
		Where("scope = ? AND idempotency_key = ? AND request_hash = ?", scope, key, requestHash).
		Updates(map[string]interface{}{
			"status":           outcome.Status,
			"response_payload": outcome.ResponseJSON,
			"resource_id":      outcome.ResourceID,
			"updated_at":       time.Now(),
		}).Error
}

func (r *TransferIdempotencyRepository) Release(ctx context.Context, scope, key, requestHash string) error {
	return r.db.WithContext(ctx).Unscoped().
		Where("scope = ? AND idempotency_key = ? AND request_hash = ?", scope, key, requestHash).
		Delete(&transferIdempotencyRecord{}).Error
}

func mapTransferIdempotencyRecord(rec *transferIdempotencyRecord) *idempotency.Record {
	if rec == nil {
		return nil
	}
	var opID string
	if rec.OperationID != nil {
		opID = *rec.OperationID
	}
	return &idempotency.Record{
		Key:          rec.IdempotencyKey,
		RequestHash:  rec.RequestHash,
		OperationID:  opID,
		Status:       rec.Status,
		ResponseJSON: rec.ResponsePayload,
		ResourceID:   rec.ResourceID,
		CreatedAt:    rec.CreatedAt,
		UpdatedAt:    rec.UpdatedAt,
		ExpiresAt:    rec.ExpiresAt,
	}
}
