package handler

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"
	pbTransferStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
)

type TransferRepository interface {
	GetMonthlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyAmount, error)
	GetYearlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]repository.YearlyAmount, error)
	GetMonthlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int, targetStatus string) ([]repository.MonthlyStatusStats, error)
	GetYearlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, currentYear int, targetStatus string) ([]repository.YearlyStatusStats, error)
}

type TransferStatsHandler struct {
	pbTransferStats.UnimplementedTransferStatsAmountServiceServer
	pbTransferStats.UnimplementedTransferStatsStatusServiceServer
	repo TransferRepository
	log  logger.LoggerInterface
}

func NewTransferStatsHandler(repo TransferRepository, log logger.LoggerInterface) *TransferStatsHandler {
	return &TransferStatsHandler{
		repo: repo,
		log:  log,
	}
}

// --- Transfer Stats Service ---

func (h *TransferStatsHandler) FindMonthlyTransferAmounts(ctx context.Context, req *transfer.FindYearTransferStatus) (*pbTransferStats.ApiResponseTransferMonthAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transfer_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferMonthAmount{
		Status:  "success",
		Message: "Retrieved monthly transfer amounts",
		Data:    h.mapToTransferMonthAmountData(data),
	}, nil
}

func (h *TransferStatsHandler) FindYearlyTransferAmounts(ctx context.Context, req *transfer.FindYearTransferStatus) (*pbTransferStats.ApiResponseTransferYearAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transfer_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferYearAmount{
		Status:  "success",
		Message: "Retrieved yearly transfer amounts",
		Data:    h.mapToTransferYearAmountData(data),
	}, nil
}

func (h *TransferStatsHandler) FindMonthlyTransferAmountsBySenderCardNumber(ctx context.Context, req *transfer.FindByCardNumberTransferRequest) (*pbTransferStats.ApiResponseTransferMonthAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transfer_events", "source_card", req.CardNumber, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferMonthAmount{
		Status:  "success",
		Message: "Retrieved monthly transfer amounts by sender card number",
		Data:    h.mapToTransferMonthAmountData(data),
	}, nil
}

func (h *TransferStatsHandler) FindMonthlyTransferAmountsByReceiverCardNumber(ctx context.Context, req *transfer.FindByCardNumberTransferRequest) (*pbTransferStats.ApiResponseTransferMonthAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transfer_events", "destination_card", req.CardNumber, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferMonthAmount{
		Status:  "success",
		Message: "Retrieved monthly transfer amounts by receiver card number",
		Data:    h.mapToTransferMonthAmountData(data),
	}, nil
}

func (h *TransferStatsHandler) FindYearlyTransferAmountsBySenderCardNumber(ctx context.Context, req *transfer.FindByCardNumberTransferRequest) (*pbTransferStats.ApiResponseTransferYearAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transfer_events", "source_card", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferYearAmount{
		Status:  "success",
		Message: "Retrieved yearly transfer amounts by sender card number",
		Data:    h.mapToTransferYearAmountData(data),
	}, nil
}

func (h *TransferStatsHandler) FindYearlyTransferAmountsByReceiverCardNumber(ctx context.Context, req *transfer.FindByCardNumberTransferRequest) (*pbTransferStats.ApiResponseTransferYearAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transfer_events", "destination_card", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferYearAmount{
		Status:  "success",
		Message: "Retrieved yearly transfer amounts by receiver card number",
		Data:    h.mapToTransferYearAmountData(data),
	}, nil
}

func (h *TransferStatsHandler) FindMonthlyTransferStatusSuccess(ctx context.Context, req *transfer.FindMonthlyTransferStatus) (*pbTransferStats.ApiResponseTransferMonthStatusSuccess, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "transfer_events", "", nil, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferMonthStatusSuccess{
		Status:  "success",
		Message: "Retrieved monthly transfer status success",
		Data:    h.mapToTransferMonthStatusSuccessData(data),
	}, nil
}

func (h *TransferStatsHandler) FindYearlyTransferStatusSuccess(ctx context.Context, req *transfer.FindYearTransferStatus) (*pbTransferStats.ApiResponseTransferYearStatusSuccess, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "transfer_events", "", nil, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferYearStatusSuccess{
		Status:  "success",
		Message: "Retrieved yearly transfer status success",
		Data:    h.mapToTransferYearStatusSuccessData(data),
	}, nil
}

func (h *TransferStatsHandler) FindMonthlyTransferStatusFailed(ctx context.Context, req *transfer.FindMonthlyTransferStatus) (*pbTransferStats.ApiResponseTransferMonthStatusFailed, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "transfer_events", "", nil, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferMonthStatusFailed{
		Status:  "success",
		Message: "Retrieved monthly transfer status failed",
		Data:    h.mapToTransferMonthStatusFailedData(data),
	}, nil
}

func (h *TransferStatsHandler) FindYearlyTransferStatusFailed(ctx context.Context, req *transfer.FindYearTransferStatus) (*pbTransferStats.ApiResponseTransferYearStatusFailed, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "transfer_events", "", nil, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferYearStatusFailed{
		Status:  "success",
		Message: "Retrieved yearly transfer status failed",
		Data:    h.mapToTransferYearStatusFailedData(data),
	}, nil
}

func (h *TransferStatsHandler) FindMonthlyTransferStatusSuccessByCardNumber(ctx context.Context, req *transfer.FindMonthlyTransferStatusCardNumber) (*pbTransferStats.ApiResponseTransferMonthStatusSuccess, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "transfer_events", "source_card", req.CardNumber, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferMonthStatusSuccess{
		Status:  "success",
		Message: "Retrieved monthly transfer status success by card number",
		Data:    h.mapToTransferMonthStatusSuccessData(data),
	}, nil
}

func (h *TransferStatsHandler) FindYearlyTransferStatusSuccessByCardNumber(ctx context.Context, req *transfer.FindYearTransferStatusCardNumber) (*pbTransferStats.ApiResponseTransferYearStatusSuccess, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "transfer_events", "source_card", req.CardNumber, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferYearStatusSuccess{
		Status:  "success",
		Message: "Retrieved yearly transfer status success by card number",
		Data:    h.mapToTransferYearStatusSuccessData(data),
	}, nil
}

func (h *TransferStatsHandler) FindMonthlyTransferStatusFailedByCardNumber(ctx context.Context, req *transfer.FindMonthlyTransferStatusCardNumber) (*pbTransferStats.ApiResponseTransferMonthStatusFailed, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "transfer_events", "source_card", req.CardNumber, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferMonthStatusFailed{
		Status:  "success",
		Message: "Retrieved monthly transfer status failed by card number",
		Data:    h.mapToTransferMonthStatusFailedData(data),
	}, nil
}

func (h *TransferStatsHandler) FindYearlyTransferStatusFailedByCardNumber(ctx context.Context, req *transfer.FindYearTransferStatusCardNumber) (*pbTransferStats.ApiResponseTransferYearStatusFailed, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "transfer_events", "source_card", req.CardNumber, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbTransferStats.ApiResponseTransferYearStatusFailed{
		Status:  "success",
		Message: "Retrieved yearly transfer status failed by card number",
		Data:    h.mapToTransferYearStatusFailedData(data),
	}, nil
}

// --- Mappers ---

func (h *TransferStatsHandler) mapToTransferMonthAmountData(data []repository.MonthlyAmount) []*pbTransferStats.TransferMonthAmountResponse {
	var results []*pbTransferStats.TransferMonthAmountResponse
	for _, d := range data {
		results = append(results, &pbTransferStats.TransferMonthAmountResponse{
			Month:       d.Month,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *TransferStatsHandler) mapToTransferYearAmountData(data []repository.YearlyAmount) []*pbTransferStats.TransferYearAmountResponse {
	var results []*pbTransferStats.TransferYearAmountResponse
	for _, d := range data {
		results = append(results, &pbTransferStats.TransferYearAmountResponse{
			Year:        d.Year,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *TransferStatsHandler) mapToTransferMonthStatusSuccessData(data []repository.MonthlyStatusStats) []*pbTransferStats.TransferMonthStatusSuccessResponse {
	var results []*pbTransferStats.TransferMonthStatusSuccessResponse
	for _, d := range data {
		results = append(results, &pbTransferStats.TransferMonthStatusSuccessResponse{
			Year:         d.Year,
			Month:        d.Month,
			TotalSuccess: int32(d.TotalTransactions),
			TotalAmount:  d.TotalAmount,
		})
	}
	return results
}

func (h *TransferStatsHandler) mapToTransferYearStatusSuccessData(data []repository.YearlyStatusStats) []*pbTransferStats.TransferYearStatusSuccessResponse {
	var results []*pbTransferStats.TransferYearStatusSuccessResponse
	for _, d := range data {
		results = append(results, &pbTransferStats.TransferYearStatusSuccessResponse{
			Year:         d.Year,
			TotalSuccess: int32(d.TotalTransactions),
			TotalAmount:  d.TotalAmount,
		})
	}
	return results
}

func (h *TransferStatsHandler) mapToTransferMonthStatusFailedData(data []repository.MonthlyStatusStats) []*pbTransferStats.TransferMonthStatusFailedResponse {
	var results []*pbTransferStats.TransferMonthStatusFailedResponse
	for _, d := range data {
		results = append(results, &pbTransferStats.TransferMonthStatusFailedResponse{
			Year:        d.Year,
			Month:       d.Month,
			TotalFailed: int32(d.TotalTransactions),
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *TransferStatsHandler) mapToTransferYearStatusFailedData(data []repository.YearlyStatusStats) []*pbTransferStats.TransferYearStatusFailedResponse {
	var results []*pbTransferStats.TransferYearStatusFailedResponse
	for _, d := range data {
		results = append(results, &pbTransferStats.TransferYearStatusFailedResponse{
			Year:        d.Year,
			TotalFailed: int32(d.TotalTransactions),
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}
