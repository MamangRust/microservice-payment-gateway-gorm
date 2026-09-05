package repository

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type Repository interface {
	GetMonthlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]MonthlyAmount, error)
	GetYearlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]YearlyAmount, error)
	GetMonthlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int, targetStatus string) ([]MonthlyStatusStats, error)
	GetYearlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, currentYear int, targetStatus string) ([]YearlyStatusStats, error)
	GetMonthlyMethodStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]MonthlyMethodStats, error)
	GetMonthlyTotalSaldo(ctx context.Context, year int) ([]MonthlyAmount, error)
	GetYearlyTotalSaldo(ctx context.Context, startYear, endYear int) ([]YearlyAmount, error)
	GetYearlyMethodStats(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]YearlyMethodStats, error)

	FindMerchantTransactions(ctx context.Context, merchantID int32) ([]map[string]interface{}, error)
	FindMerchantTransactionsByApikey(ctx context.Context, apiKey string) ([]map[string]interface{}, error)
}

func NewRepository(conn clickhouse.Conn) Repository {
	return NewClickHouseReaderRepository(conn)
}
