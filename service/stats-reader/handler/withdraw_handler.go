package handler

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw"
	pbWithdrawStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
)

type WithdrawRepository interface {
	GetMonthlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyAmount, error)
	GetYearlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]repository.YearlyAmount, error)
	GetMonthlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int, targetStatus string) ([]repository.MonthlyStatusStats, error)
	GetYearlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, currentYear int, targetStatus string) ([]repository.YearlyStatusStats, error)
}

type WithdrawStatsHandler struct {
	pbWithdrawStats.UnimplementedWithdrawStatsAmountServiceServer
	pbWithdrawStats.UnimplementedWithdrawStatsStatusServiceServer
	repo WithdrawRepository
	log  logger.LoggerInterface
}

func NewWithdrawStatsHandler(repo WithdrawRepository, log logger.LoggerInterface) *WithdrawStatsHandler {
	return &WithdrawStatsHandler{
		repo: repo,
		log:  log,
	}
}

// --- Withdraw Stats Amount Service ---

func (h *WithdrawStatsHandler) FindMonthlyWithdraws(ctx context.Context, req *withdraw.FindYearWithdrawStatus) (*pbWithdrawStats.ApiResponseWithdrawMonthAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "withdraw_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbWithdrawStats.ApiResponseWithdrawMonthAmount{
		Status:  "success",
		Message: "Retrieved monthly withdraw amounts",
		Data:    h.mapToWithdrawMonthAmountData(data),
	}, nil
}

func (h *WithdrawStatsHandler) FindYearlyWithdraws(ctx context.Context, req *withdraw.FindYearWithdrawStatus) (*pbWithdrawStats.ApiResponseWithdrawYearAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "withdraw_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbWithdrawStats.ApiResponseWithdrawYearAmount{
		Status:  "success",
		Message: "Retrieved yearly withdraw amounts",
		Data:    h.mapToWithdrawYearAmountData(data),
	}, nil
}

func (h *WithdrawStatsHandler) FindMonthlyWithdrawsByCardNumber(ctx context.Context, req *withdraw.FindYearWithdrawCardNumber) (*pbWithdrawStats.ApiResponseWithdrawMonthAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "withdraw_events", "card_number", req.CardNumber, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbWithdrawStats.ApiResponseWithdrawMonthAmount{
		Status:  "success",
		Message: "Retrieved monthly withdraw amounts by card number",
		Data:    h.mapToWithdrawMonthAmountData(data),
	}, nil
}

func (h *WithdrawStatsHandler) FindYearlyWithdrawsByCardNumber(ctx context.Context, req *withdraw.FindYearWithdrawCardNumber) (*pbWithdrawStats.ApiResponseWithdrawYearAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "withdraw_events", "card_number", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbWithdrawStats.ApiResponseWithdrawYearAmount{
		Status:  "success",
		Message: "Retrieved yearly withdraw amounts by card number",
		Data:    h.mapToWithdrawYearAmountData(data),
	}, nil
}

// --- Withdraw Stats Status Service ---

func (h *WithdrawStatsHandler) FindMonthlyWithdrawStatusSuccess(ctx context.Context, req *withdraw.FindMonthlyWithdrawStatus) (*pbWithdrawStats.ApiResponseWithdrawMonthStatusSuccess, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "withdraw_events", "", nil, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbWithdrawStats.ApiResponseWithdrawMonthStatusSuccess{
		Status:  "success",
		Message: "Retrieved monthly withdraw status success",
		Data:    h.mapToWithdrawMonthStatusSuccessData(data),
	}, nil
}

func (h *WithdrawStatsHandler) FindYearlyWithdrawStatusSuccess(ctx context.Context, req *withdraw.FindYearWithdrawStatus) (*pbWithdrawStats.ApiResponseWithdrawYearStatusSuccess, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "withdraw_events", "", nil, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbWithdrawStats.ApiResponseWithdrawYearStatusSuccess{
		Status:  "success",
		Message: "Retrieved yearly withdraw status success",
		Data:    h.mapToWithdrawYearStatusSuccessData(data),
	}, nil
}

func (h *WithdrawStatsHandler) FindMonthlyWithdrawStatusFailed(ctx context.Context, req *withdraw.FindMonthlyWithdrawStatus) (*pbWithdrawStats.ApiResponseWithdrawMonthStatusFailed, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "withdraw_events", "", nil, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbWithdrawStats.ApiResponseWithdrawMonthStatusFailed{
		Status:  "success",
		Message: "Retrieved monthly withdraw status failed",
		Data:    h.mapToWithdrawMonthStatusFailedData(data),
	}, nil
}

func (h *WithdrawStatsHandler) FindYearlyWithdrawStatusFailed(ctx context.Context, req *withdraw.FindYearWithdrawStatus) (*pbWithdrawStats.ApiResponseWithdrawYearStatusFailed, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "withdraw_events", "", nil, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbWithdrawStats.ApiResponseWithdrawYearStatusFailed{
		Status:  "success",
		Message: "Retrieved yearly withdraw status failed",
		Data:    h.mapToWithdrawYearStatusFailedData(data),
	}, nil
}

func (h *WithdrawStatsHandler) FindMonthlyWithdrawStatusSuccessCardNumber(ctx context.Context, req *withdraw.FindMonthlyWithdrawStatusCardNumber) (*pbWithdrawStats.ApiResponseWithdrawMonthStatusSuccess, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "withdraw_events", "card_number", req.CardNumber, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbWithdrawStats.ApiResponseWithdrawMonthStatusSuccess{
		Status:  "success",
		Message: "Retrieved monthly withdraw status success by card number",
		Data:    h.mapToWithdrawMonthStatusSuccessData(data),
	}, nil
}

func (h *WithdrawStatsHandler) FindYearlyWithdrawStatusSuccessCardNumber(ctx context.Context, req *withdraw.FindYearWithdrawStatusCardNumber) (*pbWithdrawStats.ApiResponseWithdrawYearStatusSuccess, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "withdraw_events", "card_number", req.CardNumber, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbWithdrawStats.ApiResponseWithdrawYearStatusSuccess{
		Status:  "success",
		Message: "Retrieved yearly withdraw status success by card number",
		Data:    h.mapToWithdrawYearStatusSuccessData(data),
	}, nil
}

func (h *WithdrawStatsHandler) FindMonthlyWithdrawStatusFailedCardNumber(ctx context.Context, req *withdraw.FindMonthlyWithdrawStatusCardNumber) (*pbWithdrawStats.ApiResponseWithdrawMonthStatusFailed, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "withdraw_events", "card_number", req.CardNumber, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbWithdrawStats.ApiResponseWithdrawMonthStatusFailed{
		Status:  "success",
		Message: "Retrieved monthly withdraw status failed by card number",
		Data:    h.mapToWithdrawMonthStatusFailedData(data),
	}, nil
}

func (h *WithdrawStatsHandler) FindYearlyWithdrawStatusFailedCardNumber(ctx context.Context, req *withdraw.FindYearWithdrawStatusCardNumber) (*pbWithdrawStats.ApiResponseWithdrawYearStatusFailed, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "withdraw_events", "card_number", req.CardNumber, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbWithdrawStats.ApiResponseWithdrawYearStatusFailed{
		Status:  "success",
		Message: "Retrieved yearly withdraw status failed by card number",
		Data:    h.mapToWithdrawYearStatusFailedData(data),
	}, nil
}

// --- Mappers ---

func (h *WithdrawStatsHandler) mapToWithdrawMonthAmountData(data []repository.MonthlyAmount) []*pbWithdrawStats.WithdrawMonthlyAmountResponse {
	var results []*pbWithdrawStats.WithdrawMonthlyAmountResponse
	for _, d := range data {
		results = append(results, &pbWithdrawStats.WithdrawMonthlyAmountResponse{
			Month:       d.Month,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *WithdrawStatsHandler) mapToWithdrawYearAmountData(data []repository.YearlyAmount) []*pbWithdrawStats.WithdrawYearlyAmountResponse {
	var results []*pbWithdrawStats.WithdrawYearlyAmountResponse
	for _, d := range data {
		results = append(results, &pbWithdrawStats.WithdrawYearlyAmountResponse{
			Year:        d.Year,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *WithdrawStatsHandler) mapToWithdrawMonthStatusSuccessData(data []repository.MonthlyStatusStats) []*pbWithdrawStats.WithdrawMonthStatusSuccessResponse {
	var results []*pbWithdrawStats.WithdrawMonthStatusSuccessResponse
	for _, d := range data {
		results = append(results, &pbWithdrawStats.WithdrawMonthStatusSuccessResponse{
			Year:         d.Year,
			Month:        d.Month,
			TotalSuccess: int32(d.TotalTransactions),
			TotalAmount:  d.TotalAmount,
		})
	}
	return results
}

func (h *WithdrawStatsHandler) mapToWithdrawYearStatusSuccessData(data []repository.YearlyStatusStats) []*pbWithdrawStats.WithdrawYearStatusSuccessResponse {
	var results []*pbWithdrawStats.WithdrawYearStatusSuccessResponse
	for _, d := range data {
		results = append(results, &pbWithdrawStats.WithdrawYearStatusSuccessResponse{
			Year:         d.Year,
			TotalSuccess: int32(d.TotalTransactions),
			TotalAmount:  d.TotalAmount,
		})
	}
	return results
}

func (h *WithdrawStatsHandler) mapToWithdrawMonthStatusFailedData(data []repository.MonthlyStatusStats) []*pbWithdrawStats.WithdrawMonthStatusFailedResponse {
	var results []*pbWithdrawStats.WithdrawMonthStatusFailedResponse
	for _, d := range data {
		results = append(results, &pbWithdrawStats.WithdrawMonthStatusFailedResponse{
			Year:        d.Year,
			Month:       d.Month,
			TotalFailed: int32(d.TotalTransactions),
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *WithdrawStatsHandler) mapToWithdrawYearStatusFailedData(data []repository.YearlyStatusStats) []*pbWithdrawStats.WithdrawYearStatusFailedResponse {
	var results []*pbWithdrawStats.WithdrawYearStatusFailedResponse
	for _, d := range data {
		results = append(results, &pbWithdrawStats.WithdrawYearStatusFailedResponse{
			Year:        d.Year,
			TotalFailed: int32(d.TotalTransactions),
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}
