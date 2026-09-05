package handler

import (
	"context"
	"strconv"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	topup_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/topup_errors/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type topupCommandHandleGrpc struct {
	pb.UnimplementedTopupCommandServiceServer

	service service.TopupCommandService
}

func NewTopupCommandHandleGrpc(service service.TopupCommandService) TopupCommandHandleGrpc {
	return &topupCommandHandleGrpc{
		service: service,
	}
}

func (s *topupCommandHandleGrpc) CreateTopup(ctx context.Context, req *pb.CreateTopupRequest) (*pb.ApiResponseTopup, error) {
	request := requests.CreateTopupRequest{
		CardNumber:     req.GetCardNumber(),
		TopupAmount:    int(req.GetTopupAmount()),
		TopupMethod:    req.GetTopupMethod(),
		IdempotencyKey: req.GetIdempotencyKey(),
	}
	if values := metadata.ValueFromIncomingContext(ctx, "x-user-id"); len(values) > 0 {
		request.AuthenticatedUserID, _ = strconv.Atoi(values[0])
	}

	if err := request.Validate(); err != nil {
		return nil, topup_errors.ErrGrpcValidateCreateTopup
	}

	res, err := s.service.CreateTopup(ctx, &request)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoTopup := &pb.TopupResponse{
		Id:          res.TopupID,
		CardNumber:  res.CardNumber,
		TopupNo:     res.TopupNo,
		TopupAmount: res.TopupAmount,
		TopupMethod: res.TopupMethod,
		TopupTime:   res.TopupTime.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:   formatTopupTime(res.CreatedAt),
		UpdatedAt:   formatTopupTime(res.UpdatedAt),
	}

	return &pb.ApiResponseTopup{
		Status:  "success",
		Message: "Successfully created topup",
		Data:    protoTopup,
	}, nil
}

func (s *topupCommandHandleGrpc) UpdateTopup(ctx context.Context, req *pb.UpdateTopupRequest) (*pb.ApiResponseTopup, error) {
	id := int(req.GetTopupId())

	if id == 0 {
		return nil, topup_errors.ErrGrpcTopupInvalidID
	}

	request := requests.UpdateTopupRequest{
		TopupID:     &id,
		CardNumber:  req.GetCardNumber(),
		TopupAmount: int(req.GetTopupAmount()),
		TopupMethod: req.GetTopupMethod(),
	}

	if err := request.Validate(); err != nil {
		return nil, topup_errors.ErrGrpcValidateUpdateTopup
	}

	res, err := s.service.UpdateTopup(ctx, &request)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoTopup := &pb.TopupResponse{
		Id:          res.TopupID,
		CardNumber:  res.CardNumber,
		TopupNo:     res.TopupNo,
		TopupAmount: res.TopupAmount,
		TopupMethod: res.TopupMethod,
		TopupTime:   res.TopupTime.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:   formatTopupTime(res.CreatedAt),
		UpdatedAt:   formatTopupTime(res.UpdatedAt),
	}

	return &pb.ApiResponseTopup{
		Status:  "success",
		Message: "Successfully updated topup",
		Data:    protoTopup,
	}, nil
}

func (s *topupCommandHandleGrpc) TrashedTopup(ctx context.Context, req *pb.FindByIdTopupRequest) (*pb.ApiResponseTopupDeleteAt, error) {
	id := int(req.GetTopupId())

	if id == 0 {
		return nil, topup_errors.ErrGrpcTopupInvalidID
	}

	res, err := s.service.TrashedTopup(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoTopup := &pb.TopupResponseDeleteAt{
		Id:          res.TopupID,
		CardNumber:  res.CardNumber,
		TopupNo:     res.TopupNo,
		TopupAmount: res.TopupAmount,
		TopupMethod: res.TopupMethod,
		TopupTime:   res.TopupTime.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:   formatTopupTime(res.CreatedAt),
		UpdatedAt:   formatTopupTime(res.UpdatedAt),
		DeletedAt:   wrapperspb.String(formatTopupTime(res.DeletedAt)),
	}

	return &pb.ApiResponseTopupDeleteAt{
		Status:  "success",
		Message: "Successfully trashed topup",
		Data:    protoTopup,
	}, nil
}

func (s *topupCommandHandleGrpc) RestoreTopup(ctx context.Context, req *pb.FindByIdTopupRequest) (*pb.ApiResponseTopupDeleteAt, error) {
	id := int(req.GetTopupId())

	if id == 0 {
		return nil, topup_errors.ErrGrpcTopupInvalidID
	}

	res, err := s.service.RestoreTopup(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoTopup := &pb.TopupResponseDeleteAt{
		Id:          res.TopupID,
		CardNumber:  res.CardNumber,
		TopupNo:     res.TopupNo,
		TopupAmount: res.TopupAmount,
		TopupMethod: res.TopupMethod,
		TopupTime:   res.TopupTime.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:   formatTopupTime(res.CreatedAt),
		UpdatedAt:   formatTopupTime(res.UpdatedAt),
		DeletedAt:   wrapperspb.String(formatTopupTime(res.DeletedAt)),
	}

	return &pb.ApiResponseTopupDeleteAt{
		Status:  "success",
		Message: "Successfully restored topup",
		Data:    protoTopup,
	}, nil
}

func (s *topupCommandHandleGrpc) DeleteTopupPermanent(ctx context.Context, req *pb.FindByIdTopupRequest) (*pb.ApiResponseTopupDelete, error) {
	id := int(req.GetTopupId())

	if id == 0 {
		return nil, topup_errors.ErrGrpcTopupInvalidID
	}

	_, err := s.service.DeleteTopupPermanent(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTopupDelete{
		Status:  "success",
		Message: "Successfully deleted topup permanently",
	}, nil
}

func (s *topupCommandHandleGrpc) RestoreAllTopup(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseTopupAll, error) {
	_, err := s.service.RestoreAllTopup(ctx)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTopupAll{
		Status:  "success",
		Message: "Successfully restore all topup",
	}, nil
}

func (s *topupCommandHandleGrpc) DeleteAllTopupPermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseTopupAll, error) {
	_, err := s.service.DeleteAllTopupPermanent(ctx)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTopupAll{
		Status:  "success",
		Message: "Successfully delete topup permanent",
	}, nil
}
