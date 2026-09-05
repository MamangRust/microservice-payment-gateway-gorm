package handler

import (
	"context"
	"strconv"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	withdraw_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/withdraw_errors/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type withdrawCommandHandleGrpc struct {
	pb.UnimplementedWithdrawCommandServiceServer

	withdrawCommand service.WithdrawCommandService
}

func NewWithdrawCommandHandleGrpc(
	withdrawCommand service.WithdrawCommandService,
) WithdrawCommandHandlerGrpc {
	return &withdrawCommandHandleGrpc{
		withdrawCommand: withdrawCommand,
	}
}

func (w *withdrawCommandHandleGrpc) CreateWithdraw(ctx context.Context, req *pb.CreateWithdrawRequest) (*pb.ApiResponseWithdraw, error) {
	request := &requests.CreateWithdrawRequest{
		CardNumber:     req.CardNumber,
		WithdrawAmount: int(req.WithdrawAmount),
		WithdrawTime:   req.WithdrawTime.AsTime(),
		IdempotencyKey: req.GetIdempotencyKey(),
	}
	if values := metadata.ValueFromIncomingContext(ctx, "x-user-id"); len(values) > 0 {
		request.AuthenticatedUserID, _ = strconv.Atoi(values[0])
	}

	if err := request.Validate(); err != nil {
		return nil, withdraw_errors.ErrGrpcValidateCreateWithdrawRequest
	}

	withdraw, err := w.withdrawCommand.Create(ctx, request)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	createdAt := ""
	updatedAt := ""
	if withdraw.CreatedAt != nil {
		createdAt = withdraw.CreatedAt.Format(time.RFC3339)
	}
	if withdraw.UpdatedAt != nil {
		updatedAt = withdraw.UpdatedAt.Format(time.RFC3339)
	}

	return &pb.ApiResponseWithdraw{
		Status:  "success",
		Message: "Successfully created withdraw",
		Data: &pb.WithdrawResponse{
			WithdrawId:     int32(withdraw.WithdrawID),
			WithdrawNo:     withdraw.WithdrawNo,
			CardNumber:     withdraw.CardNumber,
			WithdrawAmount: int64(withdraw.WithdrawAmount),
			WithdrawTime:   withdraw.WithdrawTime.Format(time.RFC3339),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		},
	}, nil
}

func (w *withdrawCommandHandleGrpc) UpdateWithdraw(ctx context.Context, req *pb.UpdateWithdrawRequest) (*pb.ApiResponseWithdraw, error) {
	id := int(req.GetWithdrawId())

	if id == 0 {
		return nil, withdraw_errors.ErrGrpcWithdrawInvalidID
	}

	request := &requests.UpdateWithdrawRequest{
		WithdrawID:     &id,
		CardNumber:     req.CardNumber,
		WithdrawAmount: int(req.WithdrawAmount),
		WithdrawTime:   req.WithdrawTime.AsTime(),
	}

	if err := request.Validate(); err != nil {
		return nil, withdraw_errors.ErrGrpcValidateUpdateWithdrawRequest
	}

	withdraw, err := w.withdrawCommand.Update(ctx, request)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	createdAt := ""
	updatedAt := ""
	if withdraw.CreatedAt != nil {
		createdAt = withdraw.CreatedAt.Format(time.RFC3339)
	}
	if withdraw.UpdatedAt != nil {
		updatedAt = withdraw.UpdatedAt.Format(time.RFC3339)
	}

	return &pb.ApiResponseWithdraw{
		Status:  "success",
		Message: "Successfully updated withdraw",
		Data: &pb.WithdrawResponse{
			WithdrawId:     int32(withdraw.WithdrawID),
			WithdrawNo:     withdraw.WithdrawNo,
			CardNumber:     withdraw.CardNumber,
			WithdrawAmount: int64(withdraw.WithdrawAmount),
			WithdrawTime:   withdraw.WithdrawTime.Format(time.RFC3339),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		},
	}, nil
}

func (w *withdrawCommandHandleGrpc) TrashedWithdraw(ctx context.Context, req *pb.FindByIdWithdrawRequest) (*pb.ApiResponseWithdrawDeleteAt, error) {
	id := int(req.GetWithdrawId())

	if id == 0 {
		return nil, withdraw_errors.ErrGrpcWithdrawInvalidID
	}

	withdraw, err := w.withdrawCommand.TrashedWithdraw(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	createdAt := ""
	updatedAt := ""
	deletedAt := &wrapperspb.StringValue{}
	if withdraw.CreatedAt != nil {
		createdAt = withdraw.CreatedAt.Format(time.RFC3339)
	}
	if withdraw.UpdatedAt != nil {
		updatedAt = withdraw.UpdatedAt.Format(time.RFC3339)
	}
	if withdraw.DeletedAt != nil {
		deletedAt = &wrapperspb.StringValue{Value: withdraw.DeletedAt.Format(time.RFC3339)}
	}

	return &pb.ApiResponseWithdrawDeleteAt{
		Status:  "success",
		Message: "Successfully trashed withdraw",
		Data: &pb.WithdrawResponseDeleteAt{
			WithdrawId:     int32(withdraw.WithdrawID),
			WithdrawNo:     withdraw.WithdrawNo,
			CardNumber:     withdraw.CardNumber,
			WithdrawAmount: int64(withdraw.WithdrawAmount),
			WithdrawTime:   withdraw.WithdrawTime.Format(time.RFC3339),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			DeletedAt:      deletedAt,
		},
	}, nil
}

func (w *withdrawCommandHandleGrpc) RestoreWithdraw(ctx context.Context, req *pb.FindByIdWithdrawRequest) (*pb.ApiResponseWithdrawDeleteAt, error) {
	id := int(req.GetWithdrawId())

	if id == 0 {
		return nil, withdraw_errors.ErrGrpcWithdrawInvalidID
	}

	withdraw, err := w.withdrawCommand.RestoreWithdraw(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	createdAt := ""
	updatedAt := ""
	deletedAt := &wrapperspb.StringValue{}
	if withdraw.CreatedAt != nil {
		createdAt = withdraw.CreatedAt.Format(time.RFC3339)
	}
	if withdraw.UpdatedAt != nil {
		updatedAt = withdraw.UpdatedAt.Format(time.RFC3339)
	}
	if withdraw.DeletedAt != nil {
		deletedAt = &wrapperspb.StringValue{Value: withdraw.DeletedAt.Format(time.RFC3339)}
	}

	return &pb.ApiResponseWithdrawDeleteAt{
		Status:  "success",
		Message: "Successfully restored withdraw",
		Data: &pb.WithdrawResponseDeleteAt{
			WithdrawId:     int32(withdraw.WithdrawID),
			WithdrawNo:     withdraw.WithdrawNo,
			CardNumber:     withdraw.CardNumber,
			WithdrawAmount: int64(withdraw.WithdrawAmount),
			WithdrawTime:   withdraw.WithdrawTime.Format(time.RFC3339),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			DeletedAt:      deletedAt,
		},
	}, nil
}

func (w *withdrawCommandHandleGrpc) DeleteWithdrawPermanent(ctx context.Context, req *pb.FindByIdWithdrawRequest) (*pb.ApiResponseWithdrawDelete, error) {
	id := int(req.GetWithdrawId())

	if id == 0 {
		return nil, withdraw_errors.ErrGrpcWithdrawInvalidID
	}

	_, err := w.withdrawCommand.DeleteWithdrawPermanent(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseWithdrawDelete{
		Status:  "success",
		Message: "Successfully deleted withdraw permanently",
	}, nil
}

func (s *withdrawCommandHandleGrpc) RestoreAllWithdraw(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseWithdrawAll, error) {
	_, err := s.withdrawCommand.RestoreAllWithdraw(ctx)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseWithdrawAll{
		Status:  "success",
		Message: "Successfully restore all withdraw",
	}, nil
}

func (s *withdrawCommandHandleGrpc) DeleteAllWithdrawPermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseWithdrawAll, error) {
	_, err := s.withdrawCommand.DeleteAllWithdrawPermanent(ctx)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseWithdrawAll{
		Status:  "success",
		Message: "Successfully delete withdraw permanent",
	}, nil
}
