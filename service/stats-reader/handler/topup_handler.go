package handler

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup"
	pbTopupStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
)

type TopupRepository interface {
	GetMonthlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyAmount, error)
	GetYearlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]repository.YearlyAmount, error)
	GetMonthlyMethodStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyMethodStats, error)
	GetYearlyMethodStats(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]repository.YearlyMethodStats, error)
	GetMonthlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int, targetStatus string) ([]repository.MonthlyStatusStats, error)
	GetYearlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, currentYear int, targetStatus string) ([]repository.YearlyStatusStats, error)
}

type TopupStatsHandler struct {
	pbTopupStats.UnimplementedTopupStatsAmountServiceServer
	pbTopupStats.UnimplementedTopupStatsMethodServiceServer
	pbTopupStats.UnimplementedTopupStatsStatusServiceServer
	repo TopupRepository
	log  logger.LoggerInterface
}

func NewTopupStatsHandler(repo TopupRepository, log logger.LoggerInterface) *TopupStatsHandler {
	return &TopupStatsHandler{
		repo: repo,
		log:  log,
	}
}

// --- Topup Stats Amount Service ---

func (h *TopupStatsHandler) FindMonthlyTopupAmounts(ctx context.Context, req *topup.FindYearTopupStatus) (*pbTopupStats.ApiResponseTopupMonthAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "topup_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupMonthAmount{
		Status:  "success",
		Message: "Retrieved monthly topup amounts",
		Data:    h.mapToTopupMonthAmountData(data),
	}, nil
}

func (h *TopupStatsHandler) FindYearlyTopupAmounts(ctx context.Context, req *topup.FindYearTopupStatus) (*pbTopupStats.ApiResponseTopupYearAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "topup_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupYearAmount{
		Status:  "success",
		Message: "Retrieved yearly topup amounts",
		Data:    h.mapToTopupYearAmountData(data),
	}, nil
}

func (h *TopupStatsHandler) FindMonthlyTopupAmountsByCardNumber(ctx context.Context, req *topup.FindYearTopupCardNumber) (*pbTopupStats.ApiResponseTopupMonthAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "topup_events", "card_number", req.CardNumber, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupMonthAmount{
		Status:  "success",
		Message: "Retrieved monthly topup amounts by card number",
		Data:    h.mapToTopupMonthAmountData(data),
	}, nil
}

func (h *TopupStatsHandler) FindYearlyTopupAmountsByCardNumber(ctx context.Context, req *topup.FindYearTopupCardNumber) (*pbTopupStats.ApiResponseTopupYearAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "topup_events", "card_number", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupYearAmount{
		Status:  "success",
		Message: "Retrieved yearly topup amounts by card number",
		Data:    h.mapToTopupYearAmountData(data),
	}, nil
}

// --- Topup Stats Method Service ---

func (h *TopupStatsHandler) FindMonthlyTopupMethods(ctx context.Context, req *topup.FindYearTopupStatus) (*pbTopupStats.ApiResponseTopupMonthMethod, error) {
	data, err := h.repo.GetMonthlyMethodStats(ctx, "topup_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupMonthMethod{
		Status:  "success",
		Message: "Retrieved monthly topup methods",
		Data:    h.mapToTopupMonthMethodData(data),
	}, nil
}

func (h *TopupStatsHandler) FindYearlyTopupMethods(ctx context.Context, req *topup.FindYearTopupStatus) (*pbTopupStats.ApiResponseTopupYearMethod, error) {
	data, err := h.repo.GetYearlyMethodStats(ctx, "topup_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupYearMethod{
		Status:  "success",
		Message: "Retrieved yearly topup methods",
		Data:    h.mapToTopupYearMethodData(data),
	}, nil
}

func (h *TopupStatsHandler) FindMonthlyTopupMethodsByCardNumber(ctx context.Context, req *topup.FindYearTopupCardNumber) (*pbTopupStats.ApiResponseTopupMonthMethod, error) {
	data, err := h.repo.GetMonthlyMethodStats(ctx, "topup_events", "card_number", req.CardNumber, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupMonthMethod{
		Status:  "success",
		Message: "Retrieved monthly topup methods by card number",
		Data:    h.mapToTopupMonthMethodData(data),
	}, nil
}

func (h *TopupStatsHandler) FindYearlyTopupMethodsByCardNumber(ctx context.Context, req *topup.FindYearTopupCardNumber) (*pbTopupStats.ApiResponseTopupYearMethod, error) {
	data, err := h.repo.GetYearlyMethodStats(ctx, "topup_events", "card_number", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupYearMethod{
		Status:  "success",
		Message: "Retrieved yearly topup methods by card number",
		Data:    h.mapToTopupYearMethodData(data),
	}, nil
}

// --- Topup Stats Status Service ---

func (h *TopupStatsHandler) FindMonthlyTopupStatusSuccess(ctx context.Context, req *topup.FindMonthlyTopupStatus) (*pbTopupStats.ApiResponseTopupMonthStatusSuccess, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "topup_events", "", nil, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupMonthStatusSuccess{
		Status:  "success",
		Message: "Retrieved monthly topup status success",
		Data:    h.mapToTopupMonthStatusSuccessData(data),
	}, nil
}

func (h *TopupStatsHandler) FindYearlyTopupStatusSuccess(ctx context.Context, req *topup.FindYearTopupStatus) (*pbTopupStats.ApiResponseTopupYearStatusSuccess, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "topup_events", "", nil, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupYearStatusSuccess{
		Status:  "success",
		Message: "Retrieved yearly topup status success",
		Data:    h.mapToTopupYearStatusSuccessData(data),
	}, nil
}

func (h *TopupStatsHandler) FindMonthlyTopupStatusFailed(ctx context.Context, req *topup.FindMonthlyTopupStatus) (*pbTopupStats.ApiResponseTopupMonthStatusFailed, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "topup_events", "", nil, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupMonthStatusFailed{
		Status:  "success",
		Message: "Retrieved monthly topup status failed",
		Data:    h.mapToTopupMonthStatusFailedData(data),
	}, nil
}

func (h *TopupStatsHandler) FindYearlyTopupStatusFailed(ctx context.Context, req *topup.FindYearTopupStatus) (*pbTopupStats.ApiResponseTopupYearStatusFailed, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "topup_events", "", nil, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupYearStatusFailed{
		Status:  "success",
		Message: "Retrieved yearly topup status failed",
		Data:    h.mapToTopupYearStatusFailedData(data),
	}, nil
}

func (h *TopupStatsHandler) FindMonthlyTopupStatusSuccessByCardNumber(ctx context.Context, req *topup.FindMonthlyTopupStatusCardNumber) (*pbTopupStats.ApiResponseTopupMonthStatusSuccess, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "topup_events", "card_number", req.CardNumber, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupMonthStatusSuccess{
		Status:  "success",
		Message: "Retrieved monthly topup status success by card number",
		Data:    h.mapToTopupMonthStatusSuccessData(data),
	}, nil
}

func (h *TopupStatsHandler) FindYearlyTopupStatusSuccessByCardNumber(ctx context.Context, req *topup.FindYearTopupStatusCardNumber) (*pbTopupStats.ApiResponseTopupYearStatusSuccess, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "topup_events", "card_number", req.CardNumber, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupYearStatusSuccess{
		Status:  "success",
		Message: "Retrieved yearly topup status success by card number",
		Data:    h.mapToTopupYearStatusSuccessData(data),
	}, nil
}

func (h *TopupStatsHandler) FindMonthlyTopupStatusFailedByCardNumber(ctx context.Context, req *topup.FindMonthlyTopupStatusCardNumber) (*pbTopupStats.ApiResponseTopupMonthStatusFailed, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "topup_events", "card_number", req.CardNumber, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupMonthStatusFailed{
		Status:  "success",
		Message: "Retrieved monthly topup status failed by card number",
		Data:    h.mapToTopupMonthStatusFailedData(data),
	}, nil
}

func (h *TopupStatsHandler) FindYearlyTopupStatusFailedByCardNumber(ctx context.Context, req *topup.FindYearTopupStatusCardNumber) (*pbTopupStats.ApiResponseTopupYearStatusFailed, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "topup_events", "card_number", req.CardNumber, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbTopupStats.ApiResponseTopupYearStatusFailed{
		Status:  "success",
		Message: "Retrieved yearly topup status failed by card number",
		Data:    h.mapToTopupYearStatusFailedData(data),
	}, nil
}

// --- Mappers ---

func (h *TopupStatsHandler) mapToTopupMonthAmountData(data []repository.MonthlyAmount) []*pbTopupStats.TopupMonthAmountResponse {
	var results []*pbTopupStats.TopupMonthAmountResponse
	for _, d := range data {
		results = append(results, &pbTopupStats.TopupMonthAmountResponse{
			Month:       d.Month,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *TopupStatsHandler) mapToTopupYearAmountData(data []repository.YearlyAmount) []*pbTopupStats.TopupYearlyAmountResponse {
	var results []*pbTopupStats.TopupYearlyAmountResponse
	for _, d := range data {
		results = append(results, &pbTopupStats.TopupYearlyAmountResponse{
			Year:        d.Year,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *TopupStatsHandler) mapToTopupMonthMethodData(data []repository.MonthlyMethodStats) []*pbTopupStats.TopupMonthMethodResponse {
	var results []*pbTopupStats.TopupMonthMethodResponse
	for _, d := range data {
		results = append(results, &pbTopupStats.TopupMonthMethodResponse{
			Month:       d.Month,
			TopupMethod: d.PaymentMethod,
			TotalTopups: int32(d.TotalTransactions),
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *TopupStatsHandler) mapToTopupMonthStatusSuccessData(data []repository.MonthlyStatusStats) []*pbTopupStats.TopupMonthStatusSuccessResponse {
	var results []*pbTopupStats.TopupMonthStatusSuccessResponse
	for _, d := range data {
		results = append(results, &pbTopupStats.TopupMonthStatusSuccessResponse{
			Year:         d.Year,
			Month:        d.Month,
			TotalSuccess: int32(d.TotalTransactions),
			TotalAmount:  d.TotalAmount,
		})
	}
	return results
}

func (h *TopupStatsHandler) mapToTopupYearStatusSuccessData(data []repository.YearlyStatusStats) []*pbTopupStats.TopupYearStatusSuccessResponse {
	var results []*pbTopupStats.TopupYearStatusSuccessResponse
	for _, d := range data {
		results = append(results, &pbTopupStats.TopupYearStatusSuccessResponse{
			Year:         d.Year,
			TotalSuccess: int32(d.TotalTransactions),
			TotalAmount:  d.TotalAmount,
		})
	}
	return results
}

func (h *TopupStatsHandler) mapToTopupMonthStatusFailedData(data []repository.MonthlyStatusStats) []*pbTopupStats.TopupMonthStatusFailedResponse {
	var results []*pbTopupStats.TopupMonthStatusFailedResponse
	for _, d := range data {
		results = append(results, &pbTopupStats.TopupMonthStatusFailedResponse{
			Year:        d.Year,
			Month:       d.Month,
			TotalFailed: int32(d.TotalTransactions),
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *TopupStatsHandler) mapToTopupYearStatusFailedData(data []repository.YearlyStatusStats) []*pbTopupStats.TopupYearStatusFailedResponse {
	var results []*pbTopupStats.TopupYearStatusFailedResponse
	for _, d := range data {
		results = append(results, &pbTopupStats.TopupYearStatusFailedResponse{
			Year:        d.Year,
			TotalFailed: int32(d.TotalTransactions),
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *TopupStatsHandler) mapToTopupYearMethodData(data []repository.YearlyMethodStats) []*pbTopupStats.TopupYearlyMethodResponse {
	var results []*pbTopupStats.TopupYearlyMethodResponse
	for _, d := range data {
		results = append(results, &pbTopupStats.TopupYearlyMethodResponse{
			Year:        d.Year,
			TopupMethod: d.PaymentMethod,
			TotalTopups: int32(d.TotalTransactions),
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}
