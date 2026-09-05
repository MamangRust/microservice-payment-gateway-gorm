package handler

import (
	"context"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	transaction_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/transaction_errors/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type transactionCommandHandleGrpc struct {
	pb.UnimplementedTransactionCommandServiceServer

	service service.TransactionCommandService
}

func NewTransactionCommandHandleGrpc(service service.TransactionCommandService) TransactionCommandHandleGrpc {
	return &transactionCommandHandleGrpc{
		service: service,
	}
}

func (t *transactionCommandHandleGrpc) CreateTransaction(ctx context.Context, request *pb.CreateTransactionRequest) (*pb.ApiResponseTransaction, error) {
	transactionTime := request.GetTransactionTime().AsTime()
	merchantID := int(request.GetMerchantId())

	req := requests.CreateTransactionRequest{
		CardNumber:      request.GetCardNumber(),
		Amount:          int(request.GetAmount()),
		PaymentMethod:   request.GetPaymentMethod(),
		MerchantID:      &merchantID,
		TransactionTime: transactionTime,
		IdempotencyKey:  request.GetIdempotencyKey(),
	}

	if err := req.Validate(); err != nil {
		return nil, transaction_errors.ErrGrpcValidateCreateTransactionRequest
	}

	res, err := t.service.Create(ctx, request.ApiKey, &req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTransaction{
		Status:  "success",
		Message: "Successfully created transaction",
		Data: &pb.TransactionResponse{
			Id:              res.TransactionID,
			CardNumber:      res.CardNumber,
			TransactionNo:   res.TransactionNo,
			Amount:          res.Amount,
			PaymentMethod:   res.PaymentMethod,
			MerchantId:      res.MerchantID,
			TransactionTime: res.TransactionTime.Format(time.RFC3339),
			CreatedAt:       formatTxTime(res.CreatedAt),
			UpdatedAt:       formatTxTime(res.UpdatedAt),
		},
	}, nil
}

func (t *transactionCommandHandleGrpc) UpdateTransaction(ctx context.Context, request *pb.UpdateTransactionRequest) (*pb.ApiResponseTransaction, error) {
	id := int(request.GetTransactionId())

	if id == 0 {
		return nil, transaction_errors.ErrGrpcTransactionInvalidID
	}
	transactionTime := request.GetTransactionTime().AsTime()
	merchantID := int(request.GetMerchantId())

	req := requests.UpdateTransactionRequest{
		TransactionID:   &id,
		CardNumber:      request.GetCardNumber(),
		Amount:          int(request.GetAmount()),
		PaymentMethod:   request.GetPaymentMethod(),
		MerchantID:      &merchantID,
		TransactionTime: transactionTime,
	}

	if err := req.Validate(); err != nil {
		return nil, transaction_errors.ErrGrpcValidateUpdateTransactionRequest
	}

	res, err := t.service.Update(ctx, request.ApiKey, &req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTransaction{
		Status:  "success",
		Message: "Successfully updated transaction",
		Data: &pb.TransactionResponse{
			Id:              res.TransactionID,
			CardNumber:      res.CardNumber,
			TransactionNo:   res.TransactionNo,
			Amount:          res.Amount,
			PaymentMethod:   res.PaymentMethod,
			MerchantId:      res.MerchantID,
			TransactionTime: res.TransactionTime.Format(time.RFC3339),
			CreatedAt:       formatTxTime(res.CreatedAt),
			UpdatedAt:       formatTxTime(res.UpdatedAt),
		},
	}, nil
}

func (t *transactionCommandHandleGrpc) TrashedTransaction(ctx context.Context, request *pb.FindByIdTransactionRequest) (*pb.ApiResponseTransactionDeleteAt, error) {
	id := int(request.GetTransactionId())

	if id == 0 {
		return nil, transaction_errors.ErrGrpcTransactionInvalidID
	}

	res, err := t.service.TrashedTransaction(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTransactionDeleteAt{
		Status:  "success",
		Message: "Successfully trashed transaction",
		Data: &pb.TransactionResponseDeleteAt{
			Id:              res.TransactionID,
			CardNumber:      res.CardNumber,
			TransactionNo:   res.TransactionNo,
			Amount:          res.Amount,
			PaymentMethod:   res.PaymentMethod,
			MerchantId:      res.MerchantID,
			TransactionTime: res.TransactionTime.Format(time.RFC3339),
			CreatedAt:       formatTxTime(res.CreatedAt),
			UpdatedAt:       formatTxTime(res.UpdatedAt),
			DeletedAt:       &wrapperspb.StringValue{Value: formatTxTime(res.DeletedAt)},
		},
	}, nil
}

func (t *transactionCommandHandleGrpc) RestoreTransaction(ctx context.Context, request *pb.FindByIdTransactionRequest) (*pb.ApiResponseTransactionDeleteAt, error) {
	id := int(request.GetTransactionId())

	if id == 0 {
		return nil, transaction_errors.ErrGrpcTransactionInvalidID
	}

	res, err := t.service.RestoreTransaction(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTransactionDeleteAt{
		Status:  "success",
		Message: "Successfully restored transaction",
		Data: &pb.TransactionResponseDeleteAt{
			Id:              res.TransactionID,
			CardNumber:      res.CardNumber,
			TransactionNo:   res.TransactionNo,
			Amount:          res.Amount,
			PaymentMethod:   res.PaymentMethod,
			MerchantId:      res.MerchantID,
			TransactionTime: res.TransactionTime.Format(time.RFC3339),
			CreatedAt:       formatTxTime(res.CreatedAt),
			UpdatedAt:       formatTxTime(res.UpdatedAt),
			DeletedAt:       &wrapperspb.StringValue{Value: formatTxTime(res.DeletedAt)},
		},
	}, nil
}

func (t *transactionCommandHandleGrpc) DeleteTransactionPermanent(ctx context.Context, request *pb.FindByIdTransactionRequest) (*pb.ApiResponseTransactionDelete, error) {
	id := int(request.GetTransactionId())

	if id == 0 {
		return nil, transaction_errors.ErrGrpcTransactionInvalidID
	}

	_, err := t.service.DeleteTransactionPermanent(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTransactionDelete{
		Status:  "success",
		Message: "Successfully deleted transaction",
	}, nil
}

func (t *transactionCommandHandleGrpc) RestoreAllTransaction(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseTransactionAll, error) {
	_, err := t.service.RestoreAllTransaction(ctx)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTransactionAll{
		Status:  "success",
		Message: "Successfully restore all transaction",
	}, nil
}

func (t *transactionCommandHandleGrpc) DeleteAllTransactionPermanent(ctx context.Context, _ *emptypb.Empty) (*pb.ApiResponseTransactionAll, error) {
	_, err := t.service.DeleteAllTransactionPermanent(ctx)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTransactionAll{
		Status:  "success",
		Message: "Successfully delete transaction permanent",
	}, nil
}
