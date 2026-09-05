package handler

import (
	"context"
	"math"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/service"
	pbutils "github.com/MamangRust/microservice-payment-gateway-grpc/pb/common"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/convert"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type merchantTransactionHandleGrpc struct {
	pb.UnimplementedMerchantTransactionServiceServer
	merchantTransaction service.MerchantTransactionService
}

func NewMerchantTransactionHandleGrpc(merchantTransaction service.MerchantTransactionService) MerchantTransactionHandleGrpc {
	return &merchantTransactionHandleGrpc{merchantTransaction: merchantTransaction}
}

func (s *merchantTransactionHandleGrpc) FindAllTransactionMerchant(ctx context.Context, req *pb.FindAllMerchantRequest) (*pb.ApiResponsePaginationMerchantTransaction, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllMerchantTransactions{Page: page, PageSize: pageSize, Search: search}

	transactions, totalRecords, err := s.merchantTransaction.FindAllTransactions(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoTransactions := make([]*pb.MerchantTransactionResponse, len(transactions))
	for i, txn := range transactions {
		var deletedAt *wrapperspb.StringValue
		if txn.DeletedAt != nil {
			deletedAt = wrapperspb.String(txn.DeletedAt.Format(time.RFC3339))
		}

		protoTransactions[i] = &pb.MerchantTransactionResponse{
			Id:              int32(txn.TransactionID),
			CardNumber:      txn.CardNumber,
			Amount:          int64(txn.Amount),
			PaymentMethod:   txn.PaymentMethod,
			MerchantId:      int32(txn.MerchantID),
			MerchantName:    txn.MerchantName,
			TransactionTime: txn.TransactionTime.Format(time.RFC3339),
			CreatedAt:       convert.FormatTimeRFC3339(txn.CreatedAt),
			UpdatedAt:       convert.FormatTimeRFC3339(txn.UpdatedAt),
			DeletedAt:       deletedAt,
		}
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb.ApiResponsePaginationMerchantTransaction{
		Status:         "success",
		Message:        "Successfully fetched merchant record",
		Data:           protoTransactions,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *merchantTransactionHandleGrpc) FindAllTransactionByMerchant(ctx context.Context, req *pb.FindAllMerchantTransaction) (*pb.ApiResponsePaginationMerchantTransaction, error) {
	merchant_id := int(req.MerchantId)
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllMerchantTransactionsById{
		Page:       page,
		PageSize:   pageSize,
		Search:     search,
		MerchantID: merchant_id,
	}

	transactions, totalRecords, err := s.merchantTransaction.FindAllTransactionsByMerchant(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoTransactions := make([]*pb.MerchantTransactionResponse, len(transactions))
	for i, txn := range transactions {
		var deletedAt *wrapperspb.StringValue
		if txn.DeletedAt != nil {
			deletedAt = wrapperspb.String(txn.DeletedAt.Format(time.RFC3339))
		}

		protoTransactions[i] = &pb.MerchantTransactionResponse{
			Id:              int32(txn.TransactionID),
			CardNumber:      txn.CardNumber,
			Amount:          int64(txn.Amount),
			PaymentMethod:   txn.PaymentMethod,
			MerchantId:      int32(txn.MerchantID),
			MerchantName:    txn.MerchantName,
			TransactionTime: txn.TransactionTime.Format(time.RFC3339),
			CreatedAt:       convert.FormatTimeRFC3339(txn.CreatedAt),
			UpdatedAt:       convert.FormatTimeRFC3339(txn.UpdatedAt),
			DeletedAt:       deletedAt,
		}
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb.ApiResponsePaginationMerchantTransaction{
		Status:         "success",
		Message:        "Successfully fetched merchant record",
		Data:           protoTransactions,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *merchantTransactionHandleGrpc) FindAllTransactionByApikey(ctx context.Context, req *pb.FindAllMerchantApikey) (*pb.ApiResponsePaginationMerchantTransaction, error) {
	api_key := req.GetApiKey()
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllMerchantTransactionsByApiKey{
		ApiKey:   api_key,
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transactions, totalRecords, err := s.merchantTransaction.FindAllTransactionsByApikey(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoTransactions := make([]*pb.MerchantTransactionResponse, len(transactions))
	for i, txn := range transactions {
		var deletedAt *wrapperspb.StringValue
		if txn.DeletedAt != nil {
			deletedAt = wrapperspb.String(txn.DeletedAt.Format(time.RFC3339))
		}

		protoTransactions[i] = &pb.MerchantTransactionResponse{
			Id:              int32(txn.TransactionID),
			CardNumber:      txn.CardNumber,
			Amount:          int64(txn.Amount),
			PaymentMethod:   txn.PaymentMethod,
			MerchantId:      int32(txn.MerchantID),
			MerchantName:    txn.MerchantName,
			TransactionTime: txn.TransactionTime.Format(time.RFC3339),
			CreatedAt:       convert.FormatTimeRFC3339(txn.CreatedAt),
			UpdatedAt:       convert.FormatTimeRFC3339(txn.UpdatedAt),
			DeletedAt:       deletedAt,
		}
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb.ApiResponsePaginationMerchantTransaction{
		Status:         "success",
		Message:        "Successfully fetched merchant record",
		Data:           protoTransactions,
		PaginationMeta: paginationMeta,
	}, nil
}
