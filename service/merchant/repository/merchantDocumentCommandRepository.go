package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type merchantDocumentCommandRepository struct {
	db *gorm.DB
}

func NewMerchantDocumentCommandRepository(db *gorm.DB) MerchantDocumentCommandRepository {
	return &merchantDocumentCommandRepository{db: db}
}

func (r *merchantDocumentCommandRepository) CreateMerchantDocument(ctx context.Context, request *requests.CreateMerchantDocumentRequest) (*models.MerchantDocument, error) {
	note := ""
	doc := &models.MerchantDocument{
		MerchantID:   int32(request.MerchantID),
		DocumentType: request.DocumentType,
		DocumentUrl:  request.DocumentUrl,
		Status:       "pending",
		Note:         &note,
	}

	if err := r.db.WithContext(ctx).Create(doc).Error; err != nil {
		return nil, sharedErrors.ErrFailed("create merchant document").WithInternal(err)
	}
	return doc, nil
}

func (r *merchantDocumentCommandRepository) UpdateMerchantDocument(ctx context.Context, request *requests.UpdateMerchantDocumentRequest) (*models.MerchantDocument, error) {
	if request.DocumentID == nil || *request.DocumentID <= 0 {
		return nil, sharedErrors.NewBadRequestError("merchant document ID is required")
	}

	var doc models.MerchantDocument
	err := r.db.WithContext(ctx).Where("document_id = ? AND deleted_at IS NULL", *request.DocumentID).First(&doc).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant document", "update merchant document")
	}

	doc.DocumentType = request.DocumentType
	doc.DocumentUrl = request.DocumentUrl
	doc.Status = request.Status
	note := ""
	doc.Note = &note

	if err := r.db.WithContext(ctx).Save(&doc).Error; err != nil {
		return nil, sharedErrors.ErrFailed("update merchant document").WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) UpdateMerchantDocumentStatus(ctx context.Context, request *requests.UpdateMerchantDocumentStatusRequest) (*models.MerchantDocument, error) {
	if request.DocumentID == nil || *request.DocumentID <= 0 {
		return nil, sharedErrors.NewBadRequestError("merchant document ID is required")
	}

	var doc models.MerchantDocument
	err := r.db.WithContext(ctx).Where("document_id = ? AND deleted_at IS NULL", *request.DocumentID).First(&doc).Error
	if err != nil {
		return nil, sharedErrors.ErrNoRowsOrFailed(err, "merchant document", "update merchant document status")
	}

	doc.Status = request.Status
	note := ""
	doc.Note = &note

	if err := r.db.WithContext(ctx).Save(&doc).Error; err != nil {
		return nil, sharedErrors.ErrFailed("update merchant document status").WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) TrashedMerchantDocument(ctx context.Context, documentID int) (*models.MerchantDocument, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.MerchantDocument{}).
		Where("document_id = ? AND deleted_at IS NULL", documentID).
		Update("deleted_at", now)
	if result.Error != nil {
		return nil, sharedErrors.ErrFailed("trash merchant document").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, sharedErrors.ErrNotFoundResponse("merchant document")
	}

	var doc models.MerchantDocument
	if err := r.db.WithContext(ctx).Where("document_id = ?", documentID).First(&doc).Error; err != nil {
		return nil, sharedErrors.ErrFailed("get trashed merchant document").WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) RestoreMerchantDocument(ctx context.Context, documentID int) (*models.MerchantDocument, error) {
	result := r.db.WithContext(ctx).Unscoped().Model(&models.MerchantDocument{}).
		Where("document_id = ? AND deleted_at IS NOT NULL", documentID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return nil, sharedErrors.ErrFailed("restore merchant document").WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, sharedErrors.ErrNotFoundResponse("merchant document")
	}

	var doc models.MerchantDocument
	if err := r.db.WithContext(ctx).Where("document_id = ?", documentID).First(&doc).Error; err != nil {
		return nil, sharedErrors.ErrFailed("get restored merchant document").WithInternal(err)
	}
	return &doc, nil
}

func (r *merchantDocumentCommandRepository) DeleteMerchantDocumentPermanent(ctx context.Context, documentID int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("document_id = ?", documentID).Delete(&models.MerchantDocument{})
	if result.Error != nil {
		return false, sharedErrors.ErrNoRowsOrFailed(result.Error, "merchant document", "delete merchant document permanently")
	}
	if result.RowsAffected == 0 {
		return false, sharedErrors.ErrNotFoundResponse("merchant document")
	}
	return true, nil
}

func (r *merchantDocumentCommandRepository) RestoreAllMerchantDocument(ctx context.Context) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Model(&models.MerchantDocument{}).Where("deleted_at IS NOT NULL").Update("deleted_at", nil).Error; err != nil {
		return false, sharedErrors.ErrFailed("restore all merchant documents").WithInternal(err)
	}
	return true, nil
}

func (r *merchantDocumentCommandRepository) DeleteAllMerchantDocumentPermanent(ctx context.Context) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.MerchantDocument{}).Error; err != nil {
		return false, sharedErrors.ErrFailed("delete all merchant documents permanently").WithInternal(err)
	}
	return true, nil
}
