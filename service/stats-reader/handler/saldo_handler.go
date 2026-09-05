package handler

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	pbSaldoStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
)

type SaldoRepository interface {
	GetMonthlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyAmount, error)
	GetYearlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]repository.YearlyAmount, error)
	GetMonthlyTotalSaldo(ctx context.Context, year int) ([]repository.MonthlyAmount, error)
	GetYearlyTotalSaldo(ctx context.Context, startYear, endYear int) ([]repository.YearlyAmount, error)
}

type SaldoStatsHandler struct {
	pbSaldoStats.UnimplementedSaldoStatsBalanceServiceServer
	pbSaldoStats.UnimplementedSaldoStatsTotalBalanceServer
	repo SaldoRepository
	log  logger.LoggerInterface
}

func NewSaldoStatsHandler(repo SaldoRepository, log logger.LoggerInterface) *SaldoStatsHandler {
	return &SaldoStatsHandler{
		repo: repo,
		log:  log,
	}
}

// --- Saldo Stats Balance Service ---

func (h *SaldoStatsHandler) FindMonthlySaldoBalances(ctx context.Context, req *saldo.FindYearlySaldo) (*pbSaldoStats.ApiResponseMonthSaldoBalances, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "saldo_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbSaldoStats.ApiResponseMonthSaldoBalances{
		Status:  "success",
		Message: "Retrieved monthly saldo balances",
		Data:    h.mapToSaldoMonthBalanceData(data),
	}, nil
}

func (h *SaldoStatsHandler) FindYearlySaldoBalances(ctx context.Context, req *saldo.FindYearlySaldo) (*pbSaldoStats.ApiResponseYearSaldoBalances, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "saldo_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbSaldoStats.ApiResponseYearSaldoBalances{
		Status:  "success",
		Message: "Retrieved yearly saldo balances",
		Data:    h.mapToSaldoYearBalanceData(data),
	}, nil
}

// --- Saldo Stats Total Balance Service ---

func (h *SaldoStatsHandler) FindMonthlyTotalSaldoBalance(ctx context.Context, req *saldo.FindMonthlySaldoTotalBalance) (*pbSaldoStats.ApiResponseMonthTotalSaldo, error) {
	data, err := h.repo.GetMonthlyTotalSaldo(ctx, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbSaldoStats.ApiResponseMonthTotalSaldo{
		Status:  "success",
		Message: "Retrieved monthly total saldo",
		Data:    h.mapToSaldoMonthTotalData(data),
	}, nil
}

func (h *SaldoStatsHandler) FindYearTotalSaldoBalance(ctx context.Context, req *saldo.FindYearlySaldo) (*pbSaldoStats.ApiResponseYearTotalSaldo, error) {
	data, err := h.repo.GetYearlyTotalSaldo(ctx, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbSaldoStats.ApiResponseYearTotalSaldo{
		Status:  "success",
		Message: "Retrieved yearly total saldo",
		Data:    h.mapToSaldoYearTotalData(data),
	}, nil
}


// --- Mappers ---

func (h *SaldoStatsHandler) mapToSaldoMonthBalanceData(data []repository.MonthlyAmount) []*pbSaldoStats.SaldoMonthBalanceResponse {
	var results []*pbSaldoStats.SaldoMonthBalanceResponse
	for _, d := range data {
		results = append(results, &pbSaldoStats.SaldoMonthBalanceResponse{
			Month:        d.Month,
			TotalBalance: d.TotalAmount,
		})
	}
	return results
}

func (h *SaldoStatsHandler) mapToSaldoYearBalanceData(data []repository.YearlyAmount) []*pbSaldoStats.SaldoYearBalanceResponse {
	var results []*pbSaldoStats.SaldoYearBalanceResponse
	for _, d := range data {
		results = append(results, &pbSaldoStats.SaldoYearBalanceResponse{
			Year:         d.Year,
			TotalBalance: d.TotalAmount,
		})
	}
	return results
}

func (h *SaldoStatsHandler) mapToSaldoMonthTotalData(data []repository.MonthlyAmount) []*pbSaldoStats.SaldoMonthTotalBalanceResponse {
	var results []*pbSaldoStats.SaldoMonthTotalBalanceResponse
	for _, d := range data {
		results = append(results, &pbSaldoStats.SaldoMonthTotalBalanceResponse{
			Month:        d.Month,
			Year:         d.Year,
			TotalBalance: d.TotalAmount,
		})
	}
	return results
}

func (h *SaldoStatsHandler) mapToSaldoYearTotalData(data []repository.YearlyAmount) []*pbSaldoStats.SaldoYearTotalBalanceResponse {
	var results []*pbSaldoStats.SaldoYearTotalBalanceResponse
	for _, d := range data {
		results = append(results, &pbSaldoStats.SaldoYearTotalBalanceResponse{
			Year:         d.Year,
			TotalBalance: d.TotalAmount,
		})
	}
	return results
}
