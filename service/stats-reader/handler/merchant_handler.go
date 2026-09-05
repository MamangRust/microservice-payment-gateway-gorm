package handler

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	pbMerchantStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
)

type MerchantRepository interface {
	GetMonthlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyAmount, error)
	GetYearlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]repository.YearlyAmount, error)
	GetMonthlyMethodStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyMethodStats, error)
	FindMerchantTransactions(ctx context.Context, merchantID int32) ([]map[string]interface{}, error)
	FindMerchantTransactionsByApikey(ctx context.Context, apiKey string) ([]map[string]interface{}, error)
}

type MerchantStatsHandler struct {
	pbMerchantStats.UnimplementedMerchantStatsAmountServiceServer
	pbMerchantStats.UnimplementedMerchantStatsMethodServiceServer
	pbMerchantStats.UnimplementedMerchantStatsTotalAmountServiceServer
	merchant.UnimplementedMerchantTransactionServiceServer
	repo MerchantRepository
	log  logger.LoggerInterface
}

func NewMerchantStatsHandler(repo MerchantRepository, log logger.LoggerInterface) *MerchantStatsHandler {
	return &MerchantStatsHandler{
		repo: repo,
		log:  log,
	}
}

// --- Merchant Transaction Service ---

func (h *MerchantStatsHandler) FindAllTransactionMerchant(ctx context.Context, req *merchant.FindAllMerchantRequest) (*merchant.ApiResponsePaginationMerchantTransaction, error) {
	data, err := h.repo.FindMerchantTransactions(ctx, 0)
	if err != nil {
		return nil, err
	}
	return &merchant.ApiResponsePaginationMerchantTransaction{
		Status:  "success",
		Message: "Retrieved all merchant transactions",
		Data:    h.mapToMerchantTransactions(data),
	}, nil
}

func (h *MerchantStatsHandler) FindAllTransactionByMerchant(ctx context.Context, req *merchant.FindAllMerchantTransaction) (*merchant.ApiResponsePaginationMerchantTransaction, error) {
	data, err := h.repo.FindMerchantTransactions(ctx, req.MerchantId)
	if err != nil {
		return nil, err
	}
	return &merchant.ApiResponsePaginationMerchantTransaction{
		Status:  "success",
		Message: "Retrieved transactions for merchant",
		Data:    h.mapToMerchantTransactions(data),
	}, nil
}

func (h *MerchantStatsHandler) FindAllTransactionByApikey(ctx context.Context, req *merchant.FindAllMerchantApikey) (*merchant.ApiResponsePaginationMerchantTransaction, error) {
	data, err := h.repo.FindMerchantTransactionsByApikey(ctx, req.ApiKey)
	if err != nil {
		return nil, err
	}
	return &merchant.ApiResponsePaginationMerchantTransaction{
		Status:  "success",
		Message: "Retrieved transactions by apikey",
		Data:    h.mapToMerchantTransactions(data),
	}, nil
}

// --- Merchant Stats Amount Service ---

func (h *MerchantStatsHandler) FindMonthlyAmountMerchant(ctx context.Context, req *merchant.FindYearMerchant) (*pbMerchantStats.ApiResponseMerchantMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transaction_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantMonthlyAmount{
		Status: "success",
		Data:   h.mapToMerchantMonthlyAmount(data),
	}, nil
}

func (h *MerchantStatsHandler) FindYearlyAmountMerchant(ctx context.Context, req *merchant.FindYearMerchant) (*pbMerchantStats.ApiResponseMerchantYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transaction_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantYearlyAmount{
		Status: "success",
		Data:   h.mapToMerchantYearlyAmount(data),
	}, nil
}

func (h *MerchantStatsHandler) FindMonthlyAmountByMerchants(ctx context.Context, req *merchant.FindYearMerchantById) (*pbMerchantStats.ApiResponseMerchantMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transaction_events", "merchant_id", req.MerchantId, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantMonthlyAmount{
		Status: "success",
		Data:   h.mapToMerchantMonthlyAmount(data),
	}, nil
}

func (h *MerchantStatsHandler) FindYearlyAmountByMerchants(ctx context.Context, req *merchant.FindYearMerchantById) (*pbMerchantStats.ApiResponseMerchantYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transaction_events", "merchant_id", req.MerchantId, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantYearlyAmount{
		Status: "success",
		Data:   h.mapToMerchantYearlyAmount(data),
	}, nil
}

func (h *MerchantStatsHandler) FindMonthlyAmountByApikey(ctx context.Context, req *merchant.FindYearMerchantByApikey) (*pbMerchantStats.ApiResponseMerchantMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transaction_events", "apikey", req.ApiKey, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantMonthlyAmount{
		Status: "success",
		Data:   h.mapToMerchantMonthlyAmount(data),
	}, nil
}

func (h *MerchantStatsHandler) FindYearlyAmountByApikey(ctx context.Context, req *merchant.FindYearMerchantByApikey) (*pbMerchantStats.ApiResponseMerchantYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transaction_events", "apikey", req.ApiKey, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantYearlyAmount{
		Status: "success",
		Data:   h.mapToMerchantYearlyAmount(data),
	}, nil
}

// --- Merchant Stats Method Service ---

func (h *MerchantStatsHandler) FindMonthlyPaymentMethodsMerchant(ctx context.Context, req *merchant.FindYearMerchant) (*pbMerchantStats.ApiResponseMerchantMonthlyPaymentMethod, error) {
	data, err := h.repo.GetMonthlyMethodStats(ctx, "transaction_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantMonthlyPaymentMethod{
		Status: "success",
		Data:   h.mapToMerchantMonthlyMethod(data),
	}, nil
}

func (h *MerchantStatsHandler) FindYearlyPaymentMethodMerchant(ctx context.Context, req *merchant.FindYearMerchant) (*pbMerchantStats.ApiResponseMerchantYearlyPaymentMethod, error) {
	return &pbMerchantStats.ApiResponseMerchantYearlyPaymentMethod{Status: "success"}, nil
}

func (h *MerchantStatsHandler) FindMonthlyPaymentMethodByMerchants(ctx context.Context, req *merchant.FindYearMerchantById) (*pbMerchantStats.ApiResponseMerchantMonthlyPaymentMethod, error) {
	data, err := h.repo.GetMonthlyMethodStats(ctx, "transaction_events", "merchant_id", req.MerchantId, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantMonthlyPaymentMethod{
		Status: "success",
		Data:   h.mapToMerchantMonthlyMethod(data),
	}, nil
}

func (h *MerchantStatsHandler) FindYearlyPaymentMethodByMerchants(ctx context.Context, req *merchant.FindYearMerchantById) (*pbMerchantStats.ApiResponseMerchantYearlyPaymentMethod, error) {
	return &pbMerchantStats.ApiResponseMerchantYearlyPaymentMethod{Status: "success"}, nil
}

func (h *MerchantStatsHandler) FindMonthlyPaymentMethodByApikey(ctx context.Context, req *merchant.FindYearMerchantByApikey) (*pbMerchantStats.ApiResponseMerchantMonthlyPaymentMethod, error) {
	data, err := h.repo.GetMonthlyMethodStats(ctx, "transaction_events", "apikey", req.ApiKey, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantMonthlyPaymentMethod{
		Status: "success",
		Data:   h.mapToMerchantMonthlyMethod(data),
	}, nil
}

func (h *MerchantStatsHandler) FindYearlyPaymentMethodByApikey(ctx context.Context, req *merchant.FindYearMerchantByApikey) (*pbMerchantStats.ApiResponseMerchantYearlyPaymentMethod, error) {
	return &pbMerchantStats.ApiResponseMerchantYearlyPaymentMethod{Status: "success"}, nil
}

// --- Merchant Stats Total Amount Service ---

func (h *MerchantStatsHandler) FindMonthlyTotalAmountMerchant(ctx context.Context, req *merchant.FindYearMerchant) (*pbMerchantStats.ApiResponseMerchantMonthlyTotalAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transaction_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantMonthlyTotalAmount{
		Status: "success",
		Data:   h.mapToMerchantMonthlyTotalAmount(data),
	}, nil
}

func (h *MerchantStatsHandler) FindYearlyTotalAmountMerchant(ctx context.Context, req *merchant.FindYearMerchant) (*pbMerchantStats.ApiResponseMerchantYearlyTotalAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transaction_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantYearlyTotalAmount{
		Status: "success",
		Data:   h.mapToMerchantYearlyTotalAmount(data),
	}, nil
}

func (h *MerchantStatsHandler) FindMonthlyTotalAmountByMerchants(ctx context.Context, req *merchant.FindYearMerchantById) (*pbMerchantStats.ApiResponseMerchantMonthlyTotalAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transaction_events", "merchant_id", req.MerchantId, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantMonthlyTotalAmount{
		Status: "success",
		Data:   h.mapToMerchantMonthlyTotalAmount(data),
	}, nil
}

func (h *MerchantStatsHandler) FindYearlyTotalAmountByMerchants(ctx context.Context, req *merchant.FindYearMerchantById) (*pbMerchantStats.ApiResponseMerchantYearlyTotalAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transaction_events", "merchant_id", req.MerchantId, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantYearlyTotalAmount{
		Status: "success",
		Data:   h.mapToMerchantYearlyTotalAmount(data),
	}, nil
}

func (h *MerchantStatsHandler) FindMonthlyTotalAmountByApikey(ctx context.Context, req *merchant.FindYearMerchantByApikey) (*pbMerchantStats.ApiResponseMerchantMonthlyTotalAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transaction_events", "apikey", req.ApiKey, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantMonthlyTotalAmount{
		Status: "success",
		Data:   h.mapToMerchantMonthlyTotalAmount(data),
	}, nil
}

func (h *MerchantStatsHandler) FindYearlyTotalAmountByApikey(ctx context.Context, req *merchant.FindYearMerchantByApikey) (*pbMerchantStats.ApiResponseMerchantYearlyTotalAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transaction_events", "apikey", req.ApiKey, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbMerchantStats.ApiResponseMerchantYearlyTotalAmount{
		Status: "success",
		Data:   h.mapToMerchantYearlyTotalAmount(data),
	}, nil
}

// --- Mappers ---

func (h *MerchantStatsHandler) mapToMerchantTransactions(data []map[string]interface{}) []*merchant.MerchantTransactionResponse {
	var results []*merchant.MerchantTransactionResponse
	for _, d := range data {
		id, _ := d["id"].(uint64)
		amount, _ := d["amount"].(int64)

		method, _ := d["method"].(string)
		createdAt, _ := d["created_at"].(string)
		results = append(results, &merchant.MerchantTransactionResponse{
			Id:              int32(id),
			Amount:          int64(amount),
			PaymentMethod:   method,
			TransactionTime: createdAt,
		})
	}
	return results
}

func (h *MerchantStatsHandler) mapToMerchantMonthlyAmount(data []repository.MonthlyAmount) []*pbMerchantStats.MerchantResponseMonthlyAmount {
	var results []*pbMerchantStats.MerchantResponseMonthlyAmount
	for _, d := range data {
		results = append(results, &pbMerchantStats.MerchantResponseMonthlyAmount{
			Month:       d.Month,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *MerchantStatsHandler) mapToMerchantYearlyAmount(data []repository.YearlyAmount) []*pbMerchantStats.MerchantResponseYearlyAmount {
	var results []*pbMerchantStats.MerchantResponseYearlyAmount
	for _, d := range data {
		results = append(results, &pbMerchantStats.MerchantResponseYearlyAmount{
			Year:        d.Year,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *MerchantStatsHandler) mapToMerchantMonthlyMethod(data []repository.MonthlyMethodStats) []*pbMerchantStats.MerchantResponseMonthlyPaymentMethod {
	var results []*pbMerchantStats.MerchantResponseMonthlyPaymentMethod
	for _, d := range data {
		results = append(results, &pbMerchantStats.MerchantResponseMonthlyPaymentMethod{
			Month:         d.Month,
			PaymentMethod: d.PaymentMethod,
			TotalAmount:   d.TotalAmount,
		})
	}
	return results
}

func (h *MerchantStatsHandler) mapToMerchantMonthlyTotalAmount(data []repository.MonthlyAmount) []*pbMerchantStats.MerchantResponseMonthlyTotalAmount {
	var results []*pbMerchantStats.MerchantResponseMonthlyTotalAmount
	for _, d := range data {
		results = append(results, &pbMerchantStats.MerchantResponseMonthlyTotalAmount{
			Month:       d.Month,
			Year:        d.Year,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *MerchantStatsHandler) mapToMerchantYearlyTotalAmount(data []repository.YearlyAmount) []*pbMerchantStats.MerchantResponseYearlyTotalAmount {
	var results []*pbMerchantStats.MerchantResponseYearlyTotalAmount
	for _, d := range data {
		results = append(results, &pbMerchantStats.MerchantResponseYearlyTotalAmount{
			Year:        d.Year,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}
