package handler

import (
	"context"
	"math"
	"time"

	pbhelper "github.com/MamangRust/microservice-payment-gateway-grpc/pb/common"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	transfer_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/transfer_errors/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type transferQueryHandleGrpc struct {
	pb.UnimplementedTransferQueryServiceServer

	transferQueryService service.TransferQueryService
}

func NewTransferQueryHandler(service service.TransferQueryService) TransferQueryHandleGrpc {
	return &transferQueryHandleGrpc{
		transferQueryService: service,
	}
}

func (s *transferQueryHandleGrpc) FindAllTransfer(ctx context.Context, request *pb.FindAllTransferRequest) (*pb.ApiResponsePaginationTransfer, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransfers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transfers, totalRecords, err := s.transferQueryService.FindAll(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbhelper.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	transferResponses := make([]*pb.TransferResponse, len(transfers))
	for i, t := range transfers {
		transfer := t.Transfer
		createdAt := ""
		updatedAt := ""
		if transfer.CreatedAt != nil {
			createdAt = transfer.CreatedAt.Format(time.RFC3339)
		}
		if transfer.UpdatedAt != nil {
			updatedAt = transfer.UpdatedAt.Format(time.RFC3339)
		}
		transferResponses[i] = &pb.TransferResponse{
			Id:             int32(transfer.TransferID),
			TransferNo:     transfer.TransferNo,
			TransferFrom:   transfer.TransferFrom,
			TransferTo:     transfer.TransferTo,
			TransferAmount: int64(transfer.TransferAmount),
			TransferTime:   transfer.TransferTime.Format(time.RFC3339),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		}
	}

	return &pb.ApiResponsePaginationTransfer{
		Status:         "success",
		Message:        "Successfully fetch transfer records",
		Data:           transferResponses,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *transferQueryHandleGrpc) FindByIdTransfer(ctx context.Context, request *pb.FindByIdTransferRequest) (*pb.ApiResponseTransfer, error) {
	id := int(request.GetTransferId())

	if id == 0 {
		return nil, transfer_errors.ErrGrpcTransferInvalidID
	}

	transfer, err := s.transferQueryService.FindById(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	createdAt := ""
	updatedAt := ""
	if transfer.CreatedAt != nil {
		createdAt = transfer.CreatedAt.Format(time.RFC3339)
	}
	if transfer.UpdatedAt != nil {
		updatedAt = transfer.UpdatedAt.Format(time.RFC3339)
	}

	return &pb.ApiResponseTransfer{
		Status:  "success",
		Message: "Successfully fetch transfer record",
		Data: &pb.TransferResponse{
			Id:             int32(transfer.TransferID),
			TransferNo:     transfer.TransferNo,
			TransferFrom:   transfer.TransferFrom,
			TransferTo:     transfer.TransferTo,
			TransferAmount: int64(transfer.TransferAmount),
			TransferTime:   transfer.TransferTime.Format(time.RFC3339),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		},
	}, nil
}

func (s *transferQueryHandleGrpc) FindTransferByTransferFrom(ctx context.Context, request *pb.FindTransferByTransferFromRequest) (*pb.ApiResponseTransfers, error) {
	transfer_from := request.GetTransferFrom()

	if transfer_from == "" {
		return nil, transfer_errors.ErrGrpcInvalidCardNumber
	}

	transfers, err := s.transferQueryService.FindTransferByTransferFrom(ctx, transfer_from)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	transferResponses := make([]*pb.TransferResponse, len(transfers))
	for i, transfer := range transfers {
		createdAt := ""
		updatedAt := ""
		if transfer.CreatedAt != nil {
			createdAt = transfer.CreatedAt.Format(time.RFC3339)
		}
		if transfer.UpdatedAt != nil {
			updatedAt = transfer.UpdatedAt.Format(time.RFC3339)
		}
		transferResponses[i] = &pb.TransferResponse{
			Id:             int32(transfer.TransferID),
			TransferNo:     transfer.TransferNo,
			TransferFrom:   transfer.TransferFrom,
			TransferTo:     transfer.TransferTo,
			TransferAmount: int64(transfer.TransferAmount),
			TransferTime:   transfer.TransferTime.Format(time.RFC3339),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		}
	}

	return &pb.ApiResponseTransfers{
		Status:  "success",
		Message: "Successfully fetch transfer records",
		Data:    transferResponses,
	}, nil
}

func (s *transferQueryHandleGrpc) FindTransferByTransferTo(ctx context.Context, request *pb.FindTransferByTransferToRequest) (*pb.ApiResponseTransfers, error) {
	transfer_to := request.GetTransferTo()

	if transfer_to == "" {
		return nil, transfer_errors.ErrGrpcInvalidCardNumber
	}

	transfers, err := s.transferQueryService.FindTransferByTransferTo(ctx, transfer_to)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	transferResponses := make([]*pb.TransferResponse, len(transfers))
	for i, transfer := range transfers {
		createdAt := ""
		updatedAt := ""
		if transfer.CreatedAt != nil {
			createdAt = transfer.CreatedAt.Format(time.RFC3339)
		}
		if transfer.UpdatedAt != nil {
			updatedAt = transfer.UpdatedAt.Format(time.RFC3339)
		}
		transferResponses[i] = &pb.TransferResponse{
			Id:             int32(transfer.TransferID),
			TransferNo:     transfer.TransferNo,
			TransferFrom:   transfer.TransferFrom,
			TransferTo:     transfer.TransferTo,
			TransferAmount: int64(transfer.TransferAmount),
			TransferTime:   transfer.TransferTime.Format(time.RFC3339),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		}
	}

	return &pb.ApiResponseTransfers{
		Status:  "success",
		Message: "Successfully fetch transfer records",
		Data:    transferResponses,
	}, nil
}

func (s *transferQueryHandleGrpc) FindByActiveTransfer(ctx context.Context, req *pb.FindAllTransferRequest) (*pb.ApiResponsePaginationTransferDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransfers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transfers, totalRecords, err := s.transferQueryService.FindByActive(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbhelper.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	transferResponses := make([]*pb.TransferResponseDeleteAt, len(transfers))
	for i, t := range transfers {
		transfer := t.Transfer
		createdAt := ""
		updatedAt := ""
		deletedAt := &wrapperspb.StringValue{}
		if transfer.CreatedAt != nil {
			createdAt = transfer.CreatedAt.Format(time.RFC3339)
		}
		if transfer.UpdatedAt != nil {
			updatedAt = transfer.UpdatedAt.Format(time.RFC3339)
		}
		if transfer.DeletedAt != nil {
			deletedAt = &wrapperspb.StringValue{Value: transfer.DeletedAt.Format(time.RFC3339)}
		}
		transferResponses[i] = &pb.TransferResponseDeleteAt{
			Id:             int32(transfer.TransferID),
			TransferNo:     transfer.TransferNo,
			TransferFrom:   transfer.TransferFrom,
			TransferTo:     transfer.TransferTo,
			TransferAmount: int64(transfer.TransferAmount),
			TransferTime:   transfer.TransferTime.Format(time.RFC3339),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			DeletedAt:      deletedAt,
		}
	}

	return &pb.ApiResponsePaginationTransferDeleteAt{
		Status:         "success",
		Message:        "Successfully fetch transfer records",
		Data:           transferResponses,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *transferQueryHandleGrpc) FindByTrashedTransfer(ctx context.Context, req *pb.FindAllTransferRequest) (*pb.ApiResponsePaginationTransferDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransfers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transfers, totalRecords, err := s.transferQueryService.FindByTrashed(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbhelper.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	transferResponses := make([]*pb.TransferResponseDeleteAt, len(transfers))
	for i, t := range transfers {
		transfer := t.Transfer
		createdAt := ""
		updatedAt := ""
		deletedAt := &wrapperspb.StringValue{}
		if transfer.CreatedAt != nil {
			createdAt = transfer.CreatedAt.Format(time.RFC3339)
		}
		if transfer.UpdatedAt != nil {
			updatedAt = transfer.UpdatedAt.Format(time.RFC3339)
		}
		if transfer.DeletedAt != nil {
			deletedAt = &wrapperspb.StringValue{Value: transfer.DeletedAt.Format(time.RFC3339)}
		}
		transferResponses[i] = &pb.TransferResponseDeleteAt{
			Id:             int32(transfer.TransferID),
			TransferNo:     transfer.TransferNo,
			TransferFrom:   transfer.TransferFrom,
			TransferTo:     transfer.TransferTo,
			TransferAmount: int64(transfer.TransferAmount),
			TransferTime:   transfer.TransferTime.Format(time.RFC3339),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			DeletedAt:      deletedAt,
		}
	}

	return &pb.ApiResponsePaginationTransferDeleteAt{
		Status:         "success",
		Message:        "Successfully fetch transfer records",
		Data:           transferResponses,
		PaginationMeta: paginationMeta,
	}, nil
}
