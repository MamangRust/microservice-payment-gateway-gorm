package handler

import (
	"context"
	"math"
	"time"

	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbhelpers "github.com/MamangRust/microservice-payment-gateway-grpc/pb/common"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	saldo_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/saldo_errors/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type saldoQueryHandleGrpc struct {
	pb.UnimplementedSaldoQueryServiceServer
	service service.SaldoQueryService
}

func NewSaldoQueryHandleGrpc(query service.SaldoQueryService) SaldoQueryHandleGrpc {
	return &saldoQueryHandleGrpc{service: query}
}

func formatSaldoTime(tp *time.Time) string {
	if tp == nil {
		return ""
	}
	return tp.Format(time.RFC3339)
}

func saldoTimeValue(v interface{}) int64 {
	switch value := v.(type) {
	case *int32:
		if value != nil {
			return int64(*value)
		}
	case *int64:
		if value != nil {
			return *value
		}
	}
	return 0
}

func (s *saldoQueryHandleGrpc) FindAllSaldo(ctx context.Context, req *pb.FindAllSaldoRequest) (*pb.ApiResponsePaginationSaldo, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllSaldos{Page: page, PageSize: pageSize, Search: search}

	res, totalRecords, err := s.service.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoSaldos := make([]*pb.SaldoResponse, len(res))
	for i, saldo := range res {
		protoSaldos[i] = &pb.SaldoResponse{
			SaldoId:        saldo.SaldoID,
			CardNumber:     saldo.CardNumber,
			TotalBalance:   saldo.TotalBalance,
			WithdrawTime:   formatSaldoTime(saldo.WithdrawTime),
			WithdrawAmount: saldoTimeValue(saldo.WithdrawAmount),
			CreatedAt:      formatSaldoTime(saldo.CreatedAt),
			UpdatedAt:      formatSaldoTime(saldo.UpdatedAt),
		}
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pbhelpers.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb.ApiResponsePaginationSaldo{
		Status:         "success",
		Message:        "Successfully fetched saldo record",
		Data:           protoSaldos,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *saldoQueryHandleGrpc) FindByIdSaldo(ctx context.Context, req *pb.FindByIdSaldoRequest) (*pb.ApiResponseSaldo, error) {
	id := int(req.GetSaldoId())
	if id == 0 {
		return nil, saldo_errors.ErrGrpcSaldoInvalidID
	}

	saldo, err := s.service.FindById(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseSaldo{
		Status:  "success",
		Message: "Successfully fetched saldo record",
		Data: &pb.SaldoResponse{
			SaldoId:        saldo.SaldoID,
			CardNumber:     saldo.CardNumber,
			TotalBalance:   saldo.TotalBalance,
			WithdrawTime:   formatSaldoTime(saldo.WithdrawTime),
			WithdrawAmount: saldoTimeValue(saldo.WithdrawAmount),
			CreatedAt:      formatSaldoTime(saldo.CreatedAt),
			UpdatedAt:      formatSaldoTime(saldo.UpdatedAt),
		},
	}, nil
}

func (s *saldoQueryHandleGrpc) FindByCardNumber(ctx context.Context, req *pbcard.FindByCardNumberRequest) (*pb.ApiResponseSaldo, error) {
	cardNumber := req.GetCardNumber()
	if cardNumber == "" {
		return nil, saldo_errors.ErrGrpcSaldoInvalidCardNumber
	}

	saldo, err := s.service.FindByCardNumber(ctx, cardNumber)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseSaldo{
		Status:  "success",
		Message: "Successfully fetched saldo record",
		Data: &pb.SaldoResponse{
			SaldoId:      saldo.SaldoID,
			CardNumber:   saldo.CardNumber,
			TotalBalance: saldo.TotalBalance,
			WithdrawTime: formatSaldoTime(saldo.WithdrawTime),
			CreatedAt:    formatSaldoTime(saldo.CreatedAt),
			UpdatedAt:    formatSaldoTime(saldo.UpdatedAt),
		},
	}, nil
}

func (s *saldoQueryHandleGrpc) FindByActive(ctx context.Context, req *pb.FindAllSaldoRequest) (*pb.ApiResponsePaginationSaldoDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllSaldos{Page: page, PageSize: pageSize, Search: search}

	res, totalRecords, err := s.service.FindByActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoSaldos := make([]*pb.SaldoResponseDeleteAt, len(res))
	for i, saldo := range res {
		var deletedAt *wrapperspb.StringValue
		if saldo.DeletedAt != nil {
			deletedAt = &wrapperspb.StringValue{Value: saldo.DeletedAt.Format(time.RFC3339)}
		}
		protoSaldos[i] = &pb.SaldoResponseDeleteAt{
			SaldoId:        saldo.SaldoID,
			CardNumber:     saldo.CardNumber,
			TotalBalance:   saldo.TotalBalance,
			WithdrawTime:   formatSaldoTime(saldo.WithdrawTime),
			WithdrawAmount: saldoTimeValue(saldo.WithdrawAmount),
			CreatedAt:      formatSaldoTime(saldo.CreatedAt),
			UpdatedAt:      formatSaldoTime(saldo.UpdatedAt),
			DeletedAt:      deletedAt,
		}
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pbhelpers.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb.ApiResponsePaginationSaldoDeleteAt{
		Status:         "success",
		Message:        "Successfully fetched saldo record",
		Data:           protoSaldos,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *saldoQueryHandleGrpc) FindByTrashed(ctx context.Context, req *pb.FindAllSaldoRequest) (*pb.ApiResponsePaginationSaldoDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllSaldos{Page: page, PageSize: pageSize, Search: search}

	res, totalRecords, err := s.service.FindByTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoSaldos := make([]*pb.SaldoResponseDeleteAt, len(res))
	for i, saldo := range res {
		var deletedAt *wrapperspb.StringValue
		if saldo.DeletedAt != nil {
			deletedAt = &wrapperspb.StringValue{Value: saldo.DeletedAt.Format(time.RFC3339)}
		}
		protoSaldos[i] = &pb.SaldoResponseDeleteAt{
			SaldoId:        saldo.SaldoID,
			CardNumber:     saldo.CardNumber,
			TotalBalance:   saldo.TotalBalance,
			WithdrawTime:   formatSaldoTime(saldo.WithdrawTime),
			WithdrawAmount: saldoTimeValue(saldo.WithdrawAmount),
			CreatedAt:      formatSaldoTime(saldo.CreatedAt),
			UpdatedAt:      formatSaldoTime(saldo.UpdatedAt),
			DeletedAt:      deletedAt,
		}
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pbhelpers.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb.ApiResponsePaginationSaldoDeleteAt{
		Status:         "success",
		Message:        "Successfully fetched saldo record",
		Data:           protoSaldos,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *saldoQueryHandleGrpc) ListReconciliationQueue(ctx context.Context, req *pb.ListReconciliationRequest) (*pb.ApiResponseReconciliation, error) {
	items, err := s.service.ListReconciliationQueue(ctx, req.GetStatus(), req.GetLimit())
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	data := make([]*pb.ReconciliationQueueItem, 0, len(items))
	for _, item := range items {
		row := &pb.ReconciliationQueueItem{
			QueueId: item.QueueID, SaldoId: item.SaldoID, CardNumber: item.CardNumber,
			CurrentBalance: item.CurrentBalance, LedgerBalance: item.LedgerBalance,
			Difference: item.Difference, LedgerEntries: item.LedgerEntries,
			Status: item.Status, MismatchCount: item.MismatchCount,
			FirstSeenAt: item.FirstSeenAt.Format(time.RFC3339), LastSeenAt: item.LastSeenAt.Format(time.RFC3339),
		}
		if item.ResolvedAt != nil {
			row.ResolvedAt = item.ResolvedAt.Format(time.RFC3339)
		}
		if item.ResolutionOperationID != nil {
			row.ResolutionOperationId = *item.ResolutionOperationID
		}
		if item.ResolutionNote != nil {
			row.ResolutionNote = *item.ResolutionNote
		}
		data = append(data, row)
	}
	return &pb.ApiResponseReconciliation{Status: "success", Message: "Successfully fetched reconciliation queue", Data: data}, nil
}

func (s *saldoQueryHandleGrpc) ListLedgerEntries(ctx context.Context, req *pb.ListLedgerEntriesRequest) (*pb.ApiResponseLedger, error) {
	items, err := s.service.ListLedgerEntries(ctx, req.GetCardNumber(), req.GetLimit())
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}
	data := make([]*pb.LedgerEntryResponse, 0, len(items))
	for _, item := range items {
		sourceID := ""
		if item.SourceID != nil {
			sourceID = *item.SourceID
		}
		data = append(data, &pb.LedgerEntryResponse{
			EntryId: item.EntryID, OperationId: item.OperationID, CardNumber: item.CardNumber,
			Direction: item.Direction, Amount: item.Amount, Delta: item.Delta,
			BalanceBefore: item.BalanceBefore, BalanceAfter: item.BalanceAfter,
			SourceType: item.SourceType, SourceId: sourceID, CreatedAt: item.CreatedAt.Format(time.RFC3339),
		})
	}
	return &pb.ApiResponseLedger{Status: "success", Message: "Successfully fetched ledger entries", Data: data}, nil
}
