package handler

import (
	"context"
	"math"
	"time"

	pbhelpers "github.com/MamangRust/microservice-payment-gateway-grpc/pb/common"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	transaction_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/transaction_errors/grpc"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/service"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type transactionQueryHandleGrpc struct {
	pb.UnimplementedTransactionQueryServiceServer

	service service.TransactionQueryService
}

func NewTransactionQueryHandleGrpc(service service.TransactionQueryService) TransactionQueryHandleGrpc {
	return &transactionQueryHandleGrpc{
		service: service,
	}
}

func formatTxTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func (t *transactionQueryHandleGrpc) FindAllTransaction(ctx context.Context, request *pb.FindAllTransactionRequest) (*pb.ApiResponsePaginationTransaction, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransactions{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transactions, totalRecords, err := t.service.FindAll(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbhelpers.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	transactionResponses := make([]*pb.TransactionResponse, len(transactions))
	for i, result := range transactions {
		tx := result.Transaction
		transactionResponses[i] = &pb.TransactionResponse{
			Id:              tx.TransactionID,
			CardNumber:      tx.CardNumber,
			TransactionNo:   tx.TransactionNo,
			Amount:          tx.Amount,
			PaymentMethod:   tx.PaymentMethod,
			MerchantId:      tx.MerchantID,
			TransactionTime: tx.TransactionTime.Format(time.RFC3339),
			CreatedAt:       formatTxTime(tx.CreatedAt),
			UpdatedAt:       formatTxTime(tx.UpdatedAt),
		}
	}

	return &pb.ApiResponsePaginationTransaction{
		Status:         "success",
		Message:        "Successfully fetched transaction records",
		Data:           transactionResponses,
		PaginationMeta: paginationMeta,
	}, nil
}

func (t *transactionQueryHandleGrpc) FindAllTransactionByCardNumber(ctx context.Context, request *pb.FindAllTransactionCardNumberRequest) (*pb.ApiResponsePaginationTransaction, error) {
	card_number := request.GetCardNumber()
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	reqService := requests.FindAllTransactionCardNumber{
		CardNumber: card_number,
		Page:       page,
		PageSize:   pageSize,
		Search:     search,
	}

	transactions, totalRecords, err := t.service.FindAllByCardNumber(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbhelpers.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	transactionResponses := make([]*pb.TransactionResponse, len(transactions))
	for i, result := range transactions {
		tx := result.Transaction
		transactionResponses[i] = &pb.TransactionResponse{
			Id:              tx.TransactionID,
			CardNumber:      tx.CardNumber,
			TransactionNo:   tx.TransactionNo,
			Amount:          tx.Amount,
			PaymentMethod:   tx.PaymentMethod,
			MerchantId:      tx.MerchantID,
			TransactionTime: tx.TransactionTime.Format(time.RFC3339),
			CreatedAt:       formatTxTime(tx.CreatedAt),
			UpdatedAt:       formatTxTime(tx.UpdatedAt),
		}
	}

	return &pb.ApiResponsePaginationTransaction{
		Status:         "success",
		Message:        "Successfully fetched transaction records",
		Data:           transactionResponses,
		PaginationMeta: paginationMeta,
	}, nil
}

func (t *transactionQueryHandleGrpc) FindByIdTransaction(ctx context.Context, req *pb.FindByIdTransactionRequest) (*pb.ApiResponseTransaction, error) {
	id := int(req.GetTransactionId())

	if id == 0 {
		return nil, transaction_errors.ErrGrpcTransactionInvalidID
	}

	transaction, err := t.service.FindById(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseTransaction{
		Status:  "success",
		Message: "Transaction fetched successfully",
		Data: &pb.TransactionResponse{
			Id:              transaction.TransactionID,
			CardNumber:      transaction.CardNumber,
			TransactionNo:   transaction.TransactionNo,
			Amount:          transaction.Amount,
			PaymentMethod:   transaction.PaymentMethod,
			MerchantId:      transaction.MerchantID,
			TransactionTime: transaction.TransactionTime.Format(time.RFC3339),
			CreatedAt:       formatTxTime(transaction.CreatedAt),
			UpdatedAt:       formatTxTime(transaction.UpdatedAt),
		},
	}, nil
}

func (t *transactionQueryHandleGrpc) FindTransactionByMerchantId(ctx context.Context, request *pb.FindTransactionByMerchantIdRequest) (*pb.ApiResponseTransactions, error) {
	merchantID := int(request.GetMerchantId())

	if merchantID == 0 {
		return nil, transaction_errors.ErrGrpcTransactionInvalidID
	}

	transactions, err := t.service.FindTransactionByMerchantId(ctx, merchantID)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	transactionResponses := make([]*pb.TransactionResponse, len(transactions))
	for i, transaction := range transactions {
		transactionResponses[i] = &pb.TransactionResponse{
			Id:              transaction.TransactionID,
			CardNumber:      transaction.CardNumber,
			TransactionNo:   transaction.TransactionNo,
			Amount:          transaction.Amount,
			PaymentMethod:   transaction.PaymentMethod,
			MerchantId:      transaction.MerchantID,
			TransactionTime: transaction.TransactionTime.Format(time.RFC3339),
			CreatedAt:       formatTxTime(transaction.CreatedAt),
			UpdatedAt:       formatTxTime(transaction.UpdatedAt),
		}
	}

	return &pb.ApiResponseTransactions{
		Status:  "success",
		Message: "Successfully fetched transaction records",
		Data:    transactionResponses,
	}, nil
}

func (t *transactionQueryHandleGrpc) FindByActiveTransaction(ctx context.Context, req *pb.FindAllTransactionRequest) (*pb.ApiResponsePaginationTransactionDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransactions{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transactions, totalRecords, err := t.service.FindByActive(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbhelpers.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	transactionResponses := make([]*pb.TransactionResponseDeleteAt, len(transactions))
	for i, result := range transactions {
		tx := result.Transaction
		transactionResponses[i] = &pb.TransactionResponseDeleteAt{
			Id:              tx.TransactionID,
			CardNumber:      tx.CardNumber,
			TransactionNo:   tx.TransactionNo,
			Amount:          tx.Amount,
			PaymentMethod:   tx.PaymentMethod,
			MerchantId:      tx.MerchantID,
			TransactionTime: tx.TransactionTime.Format(time.RFC3339),
			CreatedAt:       formatTxTime(tx.CreatedAt),
			UpdatedAt:       formatTxTime(tx.UpdatedAt),
			DeletedAt:       &wrapperspb.StringValue{Value: formatTxTime(tx.DeletedAt)},
		}
	}

	return &pb.ApiResponsePaginationTransactionDeleteAt{
		Status:         "success",
		Message:        "Successfully fetch transactions",
		Data:           transactionResponses,
		PaginationMeta: paginationMeta,
	}, nil
}

func (t *transactionQueryHandleGrpc) FindByTrashedTransaction(ctx context.Context, req *pb.FindAllTransactionRequest) (*pb.ApiResponsePaginationTransactionDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransactions{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transactions, totalRecords, err := t.service.FindByTrashed(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbhelpers.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	transactionResponses := make([]*pb.TransactionResponseDeleteAt, len(transactions))
	for i, result := range transactions {
		tx := result.Transaction
		transactionResponses[i] = &pb.TransactionResponseDeleteAt{
			Id:              tx.TransactionID,
			CardNumber:      tx.CardNumber,
			TransactionNo:   tx.TransactionNo,
			Amount:          tx.Amount,
			PaymentMethod:   tx.PaymentMethod,
			MerchantId:      tx.MerchantID,
			TransactionTime: tx.TransactionTime.Format(time.RFC3339),
			CreatedAt:       formatTxTime(tx.CreatedAt),
			UpdatedAt:       formatTxTime(tx.UpdatedAt),
			DeletedAt:       &wrapperspb.StringValue{Value: formatTxTime(tx.DeletedAt)},
		}
	}

	return &pb.ApiResponsePaginationTransactionDeleteAt{
		Status:         "success",
		Message:        "Successfully fetch transactions",
		Data:           transactionResponses,
		PaginationMeta: paginationMeta,
	}, nil
}
