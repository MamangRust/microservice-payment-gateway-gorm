package handler

import (
	"context"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	saldo_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/saldo_errors/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type saldoCommandHandleGrpc struct {
	pb.UnimplementedSaldoCommandServiceServer
	service service.SaldoCommandService
}

func NewSaldoCommandHandleGrpc(query service.SaldoCommandService) SaldoCommandHandleGrpc {
	return &saldoCommandHandleGrpc{service: query}
}

func (s *saldoCommandHandleGrpc) CreateSaldo(ctx context.Context, req *pb.CreateSaldoRequest) (*pb.ApiResponseSaldo, error) {
	request := requests.CreateSaldoRequest{
		CardNumber:   req.GetCardNumber(),
		TotalBalance: int(req.GetTotalBalance()),
	}
	if err := request.Validate(); err != nil {
		return nil, saldo_errors.ErrGrpcValidateCreateSaldo
	}

	saldo, err := s.service.CreateSaldo(ctx, &request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseSaldo{
		Status:  "success",
		Message: "Successfully created saldo record",
		Data: &pb.SaldoResponse{
			SaldoId:      saldo.SaldoID,
			CardNumber:   saldo.CardNumber,
			TotalBalance: saldo.TotalBalance,
		},
	}, nil
}

func (s *saldoCommandHandleGrpc) UpdateSaldo(ctx context.Context, req *pb.UpdateSaldoRequest) (*pb.ApiResponseSaldo, error) {
	id := int(req.GetSaldoId())
	if id == 0 {
		return nil, saldo_errors.ErrGrpcSaldoInvalidID
	}

	request := requests.UpdateSaldoRequest{
		SaldoID:      &id,
		CardNumber:   req.GetCardNumber(),
		TotalBalance: int(req.GetTotalBalance()),
	}
	if err := request.Validate(); err != nil {
		return nil, saldo_errors.ErrGrpcValidateUpdateSaldo
	}

	saldo, err := s.service.UpdateSaldo(ctx, &request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseSaldo{
		Status:  "success",
		Message: "Successfully updated saldo record",
		Data: &pb.SaldoResponse{
			SaldoId:      saldo.SaldoID,
			CardNumber:   saldo.CardNumber,
			TotalBalance: saldo.TotalBalance,
		},
	}, nil
}

func (s *saldoCommandHandleGrpc) DebitSaldo(ctx context.Context, req *pb.DebitSaldoRequest) (*pb.ApiResponseSaldo, error) {
	request := &requests.DebitSaldoRequest{
		CardNumber:  req.GetCardNumber(),
		Amount:      int(req.GetAmount()),
		OperationID: req.GetOperationId(),
		SourceType:  req.GetSourceType(),
		SourceID:    req.GetSourceId(),
	}
	if err := request.Validate(); err != nil {
		return nil, saldo_errors.ErrGrpcValidateUpdateSaldo
	}

	saldo, err := s.service.DebitSaldo(ctx, request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return saldoMutationResponse("Successfully debited saldo record", saldo.SaldoID, saldo.CardNumber, saldo.TotalBalance), nil
}

func (s *saldoCommandHandleGrpc) CreditSaldo(ctx context.Context, req *pb.CreditSaldoRequest) (*pb.ApiResponseSaldo, error) {
	request := &requests.CreditSaldoRequest{
		CardNumber:  req.GetCardNumber(),
		Amount:      int(req.GetAmount()),
		OperationID: req.GetOperationId(),
		SourceType:  req.GetSourceType(),
		SourceID:    req.GetSourceId(),
	}
	if err := request.Validate(); err != nil {
		return nil, saldo_errors.ErrGrpcValidateUpdateSaldo
	}

	saldo, err := s.service.CreditSaldo(ctx, request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return saldoMutationResponse("Successfully credited saldo record", saldo.SaldoID, saldo.CardNumber, saldo.TotalBalance), nil
}

func (s *saldoCommandHandleGrpc) ApplySaldoAdjustment(ctx context.Context, req *pb.ApplySaldoAdjustmentRequest) (*pb.ApiResponseAdjustment, error) {
	request := &requests.ApplySaldoAdjustmentRequest{
		CardNumber:  req.GetCardNumber(),
		Delta:       req.GetDelta(),
		OperationID: req.GetOperationId(),
		SourceType:  req.GetSourceType(),
		SourceID:    req.GetSourceId(),
		Note:        req.GetNote(),
	}
	if err := request.Validate(); err != nil {
		return nil, saldo_errors.ErrGrpcValidateUpdateSaldo
	}
	res, err := s.service.ApplySaldoAdjustment(ctx, request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseAdjustment{
		Status:  "success",
		Message: "Successfully applied saldo adjustment",
		Data: &pb.SaldoResponse{
			SaldoId:      res.SaldoID,
			CardNumber:   res.CardNumber,
			TotalBalance: res.TotalBalance,
		},
	}, nil
}

func (s *saldoCommandHandleGrpc) ResolveReconciliation(ctx context.Context, req *pb.ResolveReconciliationRequest) (*pb.ApiResponseSaldoAll, error) {
	if req.GetQueueId() <= 0 || req.GetOperationId() == "" {
		return nil, saldo_errors.ErrGrpcSaldoInvalidID
	}
	if err := s.service.ResolveReconciliation(ctx, req.GetQueueId(), req.GetOperationId(), req.GetNote()); err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseSaldoAll{Status: "success", Message: "Successfully resolved reconciliation item"}, nil
}

func saldoMutationResponse(message string, saldoID int32, cardNumber string, totalBalance int64) *pb.ApiResponseSaldo {
	return &pb.ApiResponseSaldo{
		Status:  "success",
		Message: message,
		Data: &pb.SaldoResponse{
			SaldoId:      saldoID,
			CardNumber:   cardNumber,
			TotalBalance: totalBalance,
		},
	}
}

func (s *saldoCommandHandleGrpc) UpdateSaldoWithdraw(ctx context.Context, req *pb.UpdateSaldoWithdrawRequest) (*pb.ApiResponseSaldo, error) {
	withdrawTime := time.Now()
	withdrawAmount := int(req.GetWithdrawAmount())

	request := requests.UpdateSaldoWithdraw{
		CardNumber:     req.GetCardNumber(),
		TotalBalance:   int(req.GetTotalBalance()),
		WithdrawTime:   &withdrawTime,
		WithdrawAmount: &withdrawAmount,
	}
	if err := request.Validate(); err != nil {
		return nil, saldo_errors.ErrGrpcValidateUpdateSaldoWithdraw
	}

	saldo, err := s.service.UpdateSaldoWithdraw(ctx, &request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseSaldo{
		Status:  "success",
		Message: "Successfully updated saldo withdraw record",
		Data: &pb.SaldoResponse{
			SaldoId:      saldo.SaldoID,
			CardNumber:   saldo.CardNumber,
			TotalBalance: saldo.TotalBalance,
		},
	}, nil
}

func (s *saldoCommandHandleGrpc) TrashedSaldo(ctx context.Context, req *pb.FindByIdSaldoRequest) (*pb.ApiResponseSaldoDeleteAt, error) {
	id := int(req.GetSaldoId())
	if id == 0 {
		return nil, saldo_errors.ErrGrpcSaldoInvalidID
	}

	saldo, err := s.service.TrashSaldo(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var deletedAt *wrapperspb.StringValue
	if saldo.DeletedAt != nil {
		deletedAt = wrapperspb.String(saldo.DeletedAt.Format(time.RFC3339))
	}

	return &pb.ApiResponseSaldoDeleteAt{
		Status:  "success",
		Message: "Successfully trashed saldo record",
		Data: &pb.SaldoResponseDeleteAt{
			SaldoId:        saldo.SaldoID,
			CardNumber:     saldo.CardNumber,
			TotalBalance:   saldo.TotalBalance,
			WithdrawTime:   formatSaldoTime(saldo.WithdrawTime),
			WithdrawAmount: saldoTimeValue(saldo.WithdrawAmount),
			CreatedAt:      formatSaldoTime(saldo.CreatedAt),
			UpdatedAt:      formatSaldoTime(saldo.UpdatedAt),
			DeletedAt:      deletedAt,
		},
	}, nil
}

func (s *saldoCommandHandleGrpc) RestoreSaldo(ctx context.Context, req *pb.FindByIdSaldoRequest) (*pb.ApiResponseSaldoDeleteAt, error) {
	id := int(req.GetSaldoId())
	if id == 0 {
		return nil, saldo_errors.ErrGrpcSaldoInvalidID
	}

	saldo, err := s.service.RestoreSaldo(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseSaldoDeleteAt{
		Status:  "success",
		Message: "Successfully restored saldo record",
		Data: &pb.SaldoResponseDeleteAt{
			SaldoId:      saldo.SaldoID,
			CardNumber:   saldo.CardNumber,
			TotalBalance: saldo.TotalBalance,
		},
	}, nil
}

func (s *saldoCommandHandleGrpc) DeleteSaldoPermanent(ctx context.Context, req *pb.FindByIdSaldoRequest) (*pb.ApiResponseSaldoDelete, error) {
	id := int(req.GetSaldoId())
	if id == 0 {
		return nil, saldo_errors.ErrGrpcSaldoInvalidID
	}

	_, err := s.service.DeleteSaldoPermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseSaldoDelete{
		Status:  "success",
		Message: "Successfully deleted saldo record",
	}, nil
}

func (s *saldoCommandHandleGrpc) RestoreAllSaldo(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseSaldoAll, error) {
	_, err := s.service.RestoreAllSaldo(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseSaldoAll{Status: "success", Message: "Successfully restore all saldo"}, nil
}

func (s *saldoCommandHandleGrpc) DeleteAllSaldoPermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseSaldoAll, error) {
	_, err := s.service.DeleteAllSaldoPermanent(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	return &pb.ApiResponseSaldoAll{Status: "success", Message: "delete saldo permanent"}, nil
}
