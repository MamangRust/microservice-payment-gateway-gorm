package tests

import (
	"context"
	"sync"

	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
)

type MockRepository struct {
	mu sync.RWMutex

	MonthlyAmounts       []repository.MonthlyAmount
	YearlyAmounts        []repository.YearlyAmount
	MonthlyStatusStats   []repository.MonthlyStatusStats
	YearlyStatusStats    []repository.YearlyStatusStats
	MonthlyMethodStats   []repository.MonthlyMethodStats
	YearlyMethodStats    []repository.YearlyMethodStats
	MerchantTransactions []map[string]interface{}

	Err error
}

func (m *MockRepository) GetMonthlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyAmount, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MonthlyAmounts, m.Err
}

func (m *MockRepository) GetYearlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]repository.YearlyAmount, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.YearlyAmounts, m.Err
}

func (m *MockRepository) GetMonthlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int, targetStatus string) ([]repository.MonthlyStatusStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MonthlyStatusStats, m.Err
}

func (m *MockRepository) GetYearlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, currentYear int, targetStatus string) ([]repository.YearlyStatusStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.YearlyStatusStats, m.Err
}

func (m *MockRepository) GetMonthlyMethodStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyMethodStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MonthlyMethodStats, m.Err
}

func (m *MockRepository) GetMonthlyTotalSaldo(ctx context.Context, year int) ([]repository.MonthlyAmount, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MonthlyAmounts, m.Err
}

func (m *MockRepository) GetYearlyTotalSaldo(ctx context.Context, startYear, endYear int) ([]repository.YearlyAmount, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.YearlyAmounts, m.Err
}

func (m *MockRepository) GetYearlyMethodStats(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]repository.YearlyMethodStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.YearlyMethodStats, m.Err
}

func (m *MockRepository) FindMerchantTransactions(ctx context.Context, merchantID int32) ([]map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MerchantTransactions, m.Err
}

func (m *MockRepository) FindMerchantTransactionsByApikey(ctx context.Context, apiKey string) ([]map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MerchantTransactions, m.Err
}
