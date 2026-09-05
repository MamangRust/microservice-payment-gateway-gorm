package handler

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction"
	pbTransactionStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
)

type TransactionRepository interface {
	GetMonthlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyAmount, error)
	GetYearlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]repository.YearlyAmount, error)
	GetMonthlyMethodStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyMethodStats, error)
	GetYearlyMethodStats(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]repository.YearlyMethodStats, error)
	GetMonthlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int, targetStatus string) ([]repository.MonthlyStatusStats, error)
	GetYearlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, currentYear int, targetStatus string) ([]repository.YearlyStatusStats, error)
}

type TransactionStatsHandler struct {
	pbTransactionStats.UnimplementedTransactionStatsAmountServiceServer
	pbTransactionStats.UnimplementedTransactionStatsMethodServiceServer
	pbTransactionStats.UnimplementedTransactionStatsStatusServiceServer
	repo TransactionRepository
	log  logger.LoggerInterface
}

func NewTransactionStatsHandler(repo TransactionRepository, log logger.LoggerInterface) *TransactionStatsHandler {
	return &TransactionStatsHandler{
		repo: repo,
		log:  log,
	}
}

// --- Transaction Stats Amount Service ---

func (h *TransactionStatsHandler) FindMonthlyAmounts(ctx context.Context, req *transaction.FindYearTransactionStatus) (*pbTransactionStats.ApiResponseTransactionMonthAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transaction_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionMonthAmount{
		Status:  "success",
		Message: "Retrieved monthly transaction amounts",
		Data:    h.mapToTransactionMonthAmountData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindYearlyAmounts(ctx context.Context, req *transaction.FindYearTransactionStatus) (*pbTransactionStats.ApiResponseTransactionYearAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transaction_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionYearAmount{
		Status:  "success",
		Message: "Retrieved yearly transaction amounts",
		Data:    h.mapToTransactionYearAmountData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindMonthlyAmountsByCardNumber(ctx context.Context, req *transaction.FindByYearCardNumberTransactionRequest) (*pbTransactionStats.ApiResponseTransactionMonthAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transaction_events", "card_number", req.CardNumber, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionMonthAmount{
		Status:  "success",
		Message: "Retrieved monthly transaction amounts by card number",
		Data:    h.mapToTransactionMonthAmountData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindYearlyAmountsByCardNumber(ctx context.Context, req *transaction.FindByYearCardNumberTransactionRequest) (*pbTransactionStats.ApiResponseTransactionYearAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transaction_events", "card_number", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionYearAmount{
		Status:  "success",
		Message: "Retrieved yearly transaction amounts by card number",
		Data:    h.mapToTransactionYearAmountData(data),
	}, nil
}

// --- Transaction Stats Method Service ---

func (h *TransactionStatsHandler) FindMonthlyPaymentMethods(ctx context.Context, req *transaction.FindYearTransactionStatus) (*pbTransactionStats.ApiResponseTransactionMonthMethod, error) {
	data, err := h.repo.GetMonthlyMethodStats(ctx, "transaction_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionMonthMethod{
		Status:  "success",
		Message: "Retrieved monthly transaction methods",
		Data:    h.mapToTransactionMonthMethodData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindYearlyPaymentMethods(ctx context.Context, req *transaction.FindYearTransactionStatus) (*pbTransactionStats.ApiResponseTransactionYearMethod, error) {
	data, err := h.repo.GetYearlyMethodStats(ctx, "transaction_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionYearMethod{
		Status:  "success",
		Message: "Retrieved yearly transaction methods",
		Data:    h.mapToTransactionYearMethodData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindMonthlyPaymentMethodsByCardNumber(ctx context.Context, req *transaction.FindByYearCardNumberTransactionRequest) (*pbTransactionStats.ApiResponseTransactionMonthMethod, error) {
	data, err := h.repo.GetMonthlyMethodStats(ctx, "transaction_events", "card_number", req.CardNumber, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionMonthMethod{
		Status:  "success",
		Message: "Retrieved monthly transaction methods by card number",
		Data:    h.mapToTransactionMonthMethodData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindYearlyPaymentMethodsByCardNumber(ctx context.Context, req *transaction.FindByYearCardNumberTransactionRequest) (*pbTransactionStats.ApiResponseTransactionYearMethod, error) {
	data, err := h.repo.GetYearlyMethodStats(ctx, "transaction_events", "card_number", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionYearMethod{
		Status:  "success",
		Message: "Retrieved yearly transaction methods by card number",
		Data:    h.mapToTransactionYearMethodData(data),
	}, nil
}

// --- Transaction Stats Status Service ---

func (h *TransactionStatsHandler) FindMonthlyTransactionStatusSuccess(ctx context.Context, req *transaction.FindMonthlyTransactionStatus) (*pbTransactionStats.ApiResponseTransactionMonthStatusSuccess, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "transaction_events", "", nil, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionMonthStatusSuccess{
		Status:  "success",
		Message: "Retrieved monthly transaction status success",
		Data:    h.mapToTransactionMonthStatusSuccessData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindYearlyTransactionStatusSuccess(ctx context.Context, req *transaction.FindYearTransactionStatus) (*pbTransactionStats.ApiResponseTransactionYearStatusSuccess, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "transaction_events", "", nil, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionYearStatusSuccess{
		Status:  "success",
		Message: "Retrieved yearly transaction status success",
		Data:    h.mapToTransactionYearStatusSuccessData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindMonthlyTransactionStatusFailed(ctx context.Context, req *transaction.FindMonthlyTransactionStatus) (*pbTransactionStats.ApiResponseTransactionMonthStatusFailed, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "transaction_events", "", nil, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionMonthStatusFailed{
		Status:  "success",
		Message: "Retrieved monthly transaction status failed",
		Data:    h.mapToTransactionMonthStatusFailedData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindYearlyTransactionStatusFailed(ctx context.Context, req *transaction.FindYearTransactionStatus) (*pbTransactionStats.ApiResponseTransactionYearStatusFailed, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "transaction_events", "", nil, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionYearStatusFailed{
		Status:  "success",
		Message: "Retrieved yearly transaction status failed",
		Data:    h.mapToTransactionYearStatusFailedData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindMonthlyTransactionStatusSuccessByCardNumber(ctx context.Context, req *transaction.FindMonthlyTransactionStatusCardNumber) (*pbTransactionStats.ApiResponseTransactionMonthStatusSuccess, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "transaction_events", "card_number", req.CardNumber, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionMonthStatusSuccess{
		Status:  "success",
		Message: "Retrieved monthly transaction status success by card number",
		Data:    h.mapToTransactionMonthStatusSuccessData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindYearlyTransactionStatusSuccessByCardNumber(ctx context.Context, req *transaction.FindYearTransactionStatusCardNumber) (*pbTransactionStats.ApiResponseTransactionYearStatusSuccess, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "transaction_events", "card_number", req.CardNumber, int(req.Year), "success")
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionYearStatusSuccess{
		Status:  "success",
		Message: "Retrieved yearly transaction status success by card number",
		Data:    h.mapToTransactionYearStatusSuccessData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindMonthlyTransactionStatusFailedByCardNumber(ctx context.Context, req *transaction.FindMonthlyTransactionStatusCardNumber) (*pbTransactionStats.ApiResponseTransactionMonthStatusFailed, error) {
	data, err := h.repo.GetMonthlyStatusStats(ctx, "transaction_events", "card_number", req.CardNumber, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionMonthStatusFailed{
		Status:  "success",
		Message: "Retrieved monthly transaction status failed by card number",
		Data:    h.mapToTransactionMonthStatusFailedData(data),
	}, nil
}

func (h *TransactionStatsHandler) FindYearlyTransactionStatusFailedByCardNumber(ctx context.Context, req *transaction.FindYearTransactionStatusCardNumber) (*pbTransactionStats.ApiResponseTransactionYearStatusFailed, error) {
	data, err := h.repo.GetYearlyStatusStats(ctx, "transaction_events", "card_number", req.CardNumber, int(req.Year), "failed")
	if err != nil {
		return nil, err
	}
	return &pbTransactionStats.ApiResponseTransactionYearStatusFailed{
		Status:  "success",
		Message: "Retrieved yearly transaction status failed by card number",
		Data:    h.mapToTransactionYearStatusFailedData(data),
	}, nil
}

// --- Mappers ---

func (h *TransactionStatsHandler) mapToTransactionMonthAmountData(data []repository.MonthlyAmount) []*pbTransactionStats.TransactionMonthAmountResponse {
	var results []*pbTransactionStats.TransactionMonthAmountResponse
	for _, d := range data {
		results = append(results, &pbTransactionStats.TransactionMonthAmountResponse{
			Month:       d.Month,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *TransactionStatsHandler) mapToTransactionYearAmountData(data []repository.YearlyAmount) []*pbTransactionStats.TransactionYearlyAmountResponse {
	var results []*pbTransactionStats.TransactionYearlyAmountResponse
	for _, d := range data {
		results = append(results, &pbTransactionStats.TransactionYearlyAmountResponse{
			Year:        d.Year,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *TransactionStatsHandler) mapToTransactionMonthMethodData(data []repository.MonthlyMethodStats) []*pbTransactionStats.TransactionMonthMethodResponse {
	var results []*pbTransactionStats.TransactionMonthMethodResponse
	for _, d := range data {
		results = append(results, &pbTransactionStats.TransactionMonthMethodResponse{
			Month:             d.Month,
			PaymentMethod:     d.PaymentMethod,
			TotalTransactions: int32(d.TotalTransactions),
			TotalAmount:       d.TotalAmount,
		})
	}
	return results
}

func (h *TransactionStatsHandler) mapToTransactionMonthStatusSuccessData(data []repository.MonthlyStatusStats) []*pbTransactionStats.TransactionMonthStatusSuccessResponse {
	var results []*pbTransactionStats.TransactionMonthStatusSuccessResponse
	for _, d := range data {
		results = append(results, &pbTransactionStats.TransactionMonthStatusSuccessResponse{
			Year:         d.Year,
			Month:        d.Month,
			TotalSuccess: int32(d.TotalTransactions),
			TotalAmount:  d.TotalAmount,
		})
	}
	return results
}

func (h *TransactionStatsHandler) mapToTransactionYearStatusSuccessData(data []repository.YearlyStatusStats) []*pbTransactionStats.TransactionYearStatusSuccessResponse {
	var results []*pbTransactionStats.TransactionYearStatusSuccessResponse
	for _, d := range data {
		results = append(results, &pbTransactionStats.TransactionYearStatusSuccessResponse{
			Year:         d.Year,
			TotalSuccess: int32(d.TotalTransactions),
			TotalAmount:  d.TotalAmount,
		})
	}
	return results
}

func (h *TransactionStatsHandler) mapToTransactionMonthStatusFailedData(data []repository.MonthlyStatusStats) []*pbTransactionStats.TransactionMonthStatusFailedResponse {
	var results []*pbTransactionStats.TransactionMonthStatusFailedResponse
	for _, d := range data {
		results = append(results, &pbTransactionStats.TransactionMonthStatusFailedResponse{
			Year:        d.Year,
			Month:       d.Month,
			TotalFailed: int32(d.TotalTransactions),
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *TransactionStatsHandler) mapToTransactionYearStatusFailedData(data []repository.YearlyStatusStats) []*pbTransactionStats.TransactionYearStatusFailedResponse {
	var results []*pbTransactionStats.TransactionYearStatusFailedResponse
	for _, d := range data {
		results = append(results, &pbTransactionStats.TransactionYearStatusFailedResponse{
			Year:        d.Year,
			TotalFailed: int32(d.TotalTransactions),
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *TransactionStatsHandler) mapToTransactionYearMethodData(data []repository.YearlyMethodStats) []*pbTransactionStats.TransactionYearMethodResponse {
	var results []*pbTransactionStats.TransactionYearMethodResponse
	for _, d := range data {
		results = append(results, &pbTransactionStats.TransactionYearMethodResponse{
			Year:              d.Year,
			PaymentMethod:     d.PaymentMethod,
			TotalTransactions: int32(d.TotalTransactions),
			TotalAmount:       d.TotalAmount,
		})
	}
	return results
}
