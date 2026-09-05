package handler

import (
	"context"
	"math"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/service"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/common"
	pbdocument "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant_document"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	merchantdocument_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/merchant_document_errors/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type merchantDocumentQueryHandleGrpc struct {
	pbdocument.UnimplementedMerchantDocumentQueryServiceServer
	merchantDocumentQuery service.MerchantDocumentQueryService
}

func NewMerchantDocumentQueryHandleGrpc(merchantQuery service.MerchantDocumentQueryService) MerchantDocumentQueryHandleGrpc {
	return &merchantDocumentQueryHandleGrpc{merchantDocumentQuery: merchantQuery}
}

func formatDocTimePtr(tp *time.Time) string {
	if tp == nil {
		return ""
	}
	return tp.Format(time.RFC3339)
}

func (s *merchantDocumentQueryHandleGrpc) FindAll(ctx context.Context, req *pbdocument.FindAllMerchantDocumentsRequest) (*pbdocument.ApiResponsePaginationMerchantDocument, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllMerchantDocuments{Page: page, PageSize: pageSize, Search: search}

	documents, totalRecords, err := s.merchantDocumentQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var protoDocuments []*pbdocument.MerchantDocument
	for _, doc := range documents {
		protoDocuments = append(protoDocuments, &pbdocument.MerchantDocument{
			DocumentId:   int32(doc.DocumentID),
			MerchantId:   int32(doc.MerchantID),
			DocumentType: doc.DocumentType,
			DocumentUrl:  doc.DocumentUrl,
			Status:       doc.Status,
			Note:         StringValue(doc.Note),
			UploadedAt:   formatDocTimePtr(doc.UploadedAt),
			UpdatedAt:    formatDocTimePtr(doc.UpdatedAt),
		})
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pb.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pbdocument.ApiResponsePaginationMerchantDocument{
		Status:         "success",
		Message:        "Successfully fetched merchant documents",
		Data:           protoDocuments,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *merchantDocumentQueryHandleGrpc) FindById(ctx context.Context, req *pbdocument.FindMerchantDocumentByIdRequest) (*pbdocument.ApiResponseMerchantDocument, error) {
	id := int(req.GetDocumentId())

	if id == 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}

	doc, err := s.merchantDocumentQuery.FindById(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoDocument := &pbdocument.MerchantDocument{
		DocumentId:   int32(doc.DocumentID),
		MerchantId:   int32(doc.MerchantID),
		DocumentType: doc.DocumentType,
		DocumentUrl:  doc.DocumentUrl,
		Status:       doc.Status,
		Note:         StringValue(doc.Note),
		UploadedAt:   formatDocTimePtr(doc.UploadedAt),
		UpdatedAt:    formatDocTimePtr(doc.UpdatedAt),
	}

	return &pbdocument.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully fetched merchant document",
		Data:    protoDocument,
	}, nil
}

func (s *merchantDocumentQueryHandleGrpc) FindAllActive(ctx context.Context, req *pbdocument.FindAllMerchantDocumentsRequest) (*pbdocument.ApiResponsePaginationMerchantDocumentAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllMerchantDocuments{Page: page, PageSize: pageSize, Search: search}

	documents, totalRecords, err := s.merchantDocumentQuery.FindByActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var protoDocuments []*pbdocument.MerchantDocumentDeleteAt
	for _, doc := range documents {
		protoDocuments = append(protoDocuments, &pbdocument.MerchantDocumentDeleteAt{
			DocumentId:   int32(doc.DocumentID),
			MerchantId:   int32(doc.MerchantID),
			DocumentType: doc.DocumentType,
			DocumentUrl:  doc.DocumentUrl,
			Status:       doc.Status,
			Note:         StringValue(doc.Note),
			UploadedAt:   formatDocTimePtr(doc.UploadedAt),
			UpdatedAt:    formatDocTimePtr(doc.UpdatedAt),
		})
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pb.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pbdocument.ApiResponsePaginationMerchantDocumentAt{
		Status:         "success",
		Message:        "Successfully fetched active merchant documents",
		Data:           protoDocuments,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *merchantDocumentQueryHandleGrpc) FindAllTrashed(ctx context.Context, req *pbdocument.FindAllMerchantDocumentsRequest) (*pbdocument.ApiResponsePaginationMerchantDocumentAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllMerchantDocuments{Page: page, PageSize: pageSize, Search: search}

	documents, totalRecords, err := s.merchantDocumentQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var protoDocuments []*pbdocument.MerchantDocumentDeleteAt
	for _, doc := range documents {
		var deletedAt *wrapperspb.StringValue
		if doc.DeletedAt != nil {
			deletedAt = wrapperspb.String(doc.DeletedAt.Format(time.RFC3339))
		}

		protoDocuments = append(protoDocuments, &pbdocument.MerchantDocumentDeleteAt{
			DocumentId:   int32(doc.DocumentID),
			MerchantId:   int32(doc.MerchantID),
			DocumentType: doc.DocumentType,
			DocumentUrl:  doc.DocumentUrl,
			Status:       doc.Status,
			Note:         StringValue(doc.Note),
			UploadedAt:   formatDocTimePtr(doc.UploadedAt),
			UpdatedAt:    formatDocTimePtr(doc.UpdatedAt),
			DeletedAt:    deletedAt,
		})
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pb.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pbdocument.ApiResponsePaginationMerchantDocumentAt{
		Status:         "success",
		Message:        "Successfully fetched trashed merchant documents",
		Data:           protoDocuments,
		PaginationMeta: paginationMeta,
	}, nil
}
