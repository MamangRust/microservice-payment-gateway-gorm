package handler

import (
	"context"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"

	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	transfer_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/transfer_errors/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type transferCommandHandleGrpc struct {
	pb.UnimplementedTransferCommandServiceServer

	transferCommandService service.TransferCommandService
}

func NewTransferCommandHandler(service service.TransferCommandService) TransferCommandHandleGrpc {
	return &transferCommandHandleGrpc{
		transferCommandService: service,
	}
}

func (s *transferCommandHandleGrpc) CreateTransfer(ctx context.Context, request *pb.CreateTransferRequest) (*pb.ApiResponseTransfer, error) {
	req := requests.CreateTransferRequest{
		TransferFrom:   request.GetTransferFrom(),
		TransferTo:     request.GetTransferTo(),
		TransferAmount: int(request.GetTransferAmount()),
		IdempotencyKey: request.GetIdempotencyKey(),
	}
	if err := req.Validate(); err != nil {
		return nil, transfer_errors.ErrGrpcValidateCreateTransferRequest
	}

	transfer, err := s.transferCommandService.CreateTransaction(ctx, &req)

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
		Message: "Successfully created transfer",
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

func (s *transferCommandHandleGrpc) UpdateTransfer(ctx context.Context, request *pb.UpdateTransferRequest) (*pb.ApiResponseTransfer, error) {
	id := int(request.GetTransferId())

	if id == 0 {
		return nil, transfer_errors.ErrGrpcTransferInvalidID
	}

	req := requests.UpdateTransferRequest{
		TransferID:     &id,
		TransferFrom:   request.GetTransferFrom(),
		TransferTo:     request.GetTransferTo(),
		TransferAmount: int(request.GetTransferAmount()),
	}

	if err := req.Validate(); err != nil {
		return nil, transfer_errors.ErrGrpcValidateUpdateTransferRequest
	}

	transfer, err := s.transferCommandService.UpdateTransaction(ctx, &req)

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
		Message: "Successfully updated transfer",
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

func (s *transferCommandHandleGrpc) TrashedTransfer(ctx context.Context, request *pb.FindByIdTransferRequest) (*pb.ApiResponseTransferDeleteAt, error) {
	id := int(request.GetTransferId())

	if id == 0 {
		return nil, transfer_errors.ErrGrpcTransferInvalidID
	}

	transfer, err := s.transferCommandService.TrashedTransfer(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

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

	return &pb.ApiResponseTransferDeleteAt{
		Status:  "success",
		Message: "Successfully trashed transfer",
		Data: &pb.TransferResponseDeleteAt{
			Id:             int32(transfer.TransferID),
			TransferNo:     transfer.TransferNo,
			TransferFrom:   transfer.TransferFrom,
			TransferTo:     transfer.TransferTo,
			TransferAmount: int64(transfer.TransferAmount),
			TransferTime:   transfer.TransferTime.Format(time.RFC3339),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			DeletedAt:      deletedAt,
		},
	}, nil
}

func (s *transferCommandHandleGrpc) RestoreTransfer(ctx context.Context, request *pb.FindByIdTransferRequest) (*pb.ApiResponseTransferDeleteAt, error) {
	id := int(request.GetTransferId())

	if id == 0 {
		return nil, transfer_errors.ErrGrpcTransferInvalidID
	}

	transfer, err := s.transferCommandService.RestoreTransfer(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

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

	return &pb.ApiResponseTransferDeleteAt{
		Status:  "success",
		Message: "Successfully restored transfer",
		Data: &pb.TransferResponseDeleteAt{
			Id:             int32(transfer.TransferID),
			TransferNo:     transfer.TransferNo,
			TransferFrom:   transfer.TransferFrom,
			TransferTo:     transfer.TransferTo,
			TransferAmount: int64(transfer.TransferAmount),
			TransferTime:   transfer.TransferTime.Format(time.RFC3339),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			DeletedAt:      deletedAt,
		},
	}, nil
}

func (s *transferCommandHandleGrpc) DeleteTransferPermanent(ctx context.Context, request *pb.FindByIdTransferRequest) (*pb.ApiResponseTransferDelete, error) {
	id := int(request.GetTransferId())

	if id == 0 {
		return nil, transfer_errors.ErrGrpcTransferInvalidID
	}

	_, err := s.transferCommandService.DeleteTransferPermanent(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTransferDelete{
		Status:  "success",
		Message: "Successfully deleted transfer permanently",
	}, nil
}

func (s *transferCommandHandleGrpc) RestoreAllTransfer(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseTransferAll, error) {
	_, err := s.transferCommandService.RestoreAllTransfer(ctx)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTransferAll{
		Status:  "success",
		Message: "Successfully restored all transfers",
	}, nil
}

func (s *transferCommandHandleGrpc) DeleteAllTransferPermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseTransferAll, error) {
	_, err := s.transferCommandService.DeleteAllTransferPermanent(ctx)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTransferAll{
		Status:  "success",
		Message: "Successfully deleted all transfers",
	}, nil
}

