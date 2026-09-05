package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseReaderRepository struct {
	conn clickhouse.Conn
}

func NewClickHouseReaderRepository(conn clickhouse.Conn) *ClickHouseReaderRepository {
	return &ClickHouseReaderRepository{conn: conn}
}

// Common Structures
type MonthlyAmount struct {
	Year        string
	Month       string
	TotalAmount int64
}

type YearlyAmount struct {
	Year        string
	TotalAmount int64
}

type MonthlyMethodStats struct {
	Month             string
	PaymentMethod     string
	TotalTransactions uint64
	TotalAmount       int64
}

type YearlyMethodStats struct {
	Year              string
	PaymentMethod     string
	TotalTransactions uint64
	TotalAmount       int64
}

type MonthlyStatusStats struct {
	Year              string
	Month             string
	Status            string
	TotalTransactions uint64
	TotalAmount       int64
}

type YearlyStatusStats struct {
	Year              string
	Status            string
	TotalTransactions uint64
	TotalAmount       int64
}

// --- Dynamic Table-Driven Methods for Parity ---

func (r *ClickHouseReaderRepository) GetMonthlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]MonthlyAmount, error) {
	where := ""
	if table == "saldo_events" {
		where = fmt.Sprintf("toYear(created_at) = %d", year)
	} else {
		where = fmt.Sprintf("toYear(created_at) = %d AND status = 'success'", year)
	}

	if filterField != "" {
		where += fmt.Sprintf(" AND %s = ?", filterField)
	}

	amountCol := "amount"
	if table == "saldo_events" {
		amountCol = "total_balance"
	}

	query := fmt.Sprintf(`
		SELECT toString(toYear(created_at)) as year, formatDateTime(created_at, '%%b') as month, sum(%s) as total_amount
		FROM %s WHERE %s
		GROUP BY year, month, toMonth(created_at) ORDER BY year, toMonth(created_at)
	`, amountCol, table, where)

	args := []interface{}{}
	if filterField != "" {
		args = append(args, filterValue)
	}

	return r.queryMonthlyAmounts(ctx, query, args...)
}

func (r *ClickHouseReaderRepository) GetYearlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]YearlyAmount, error) {
	where := ""
	if table == "saldo_events" {
		where = fmt.Sprintf("toYear(created_at) >= %d AND toYear(created_at) <= %d", startYear, endYear)
	} else {
		where = fmt.Sprintf("toYear(created_at) >= %d AND toYear(created_at) <= %d AND status = 'success'", startYear, endYear)
	}

	if filterField != "" {
		where += fmt.Sprintf(" AND %s = ?", filterField)
	}

	amountCol := "amount"
	if table == "saldo_events" {
		amountCol = "total_balance"
	}

	query := fmt.Sprintf(`
		SELECT toString(toYear(created_at)) as year, sum(%s) as total_amount
		FROM %s WHERE %s
		GROUP BY year ORDER BY year
	`, amountCol, table, where)

	args := []interface{}{}
	if filterField != "" {
		args = append(args, filterValue)
	}

	return r.queryYearlyAmounts(ctx, query, args...)
}

func (r *ClickHouseReaderRepository) GetMonthlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int, targetStatus string) ([]MonthlyStatusStats, error) {
	where := fmt.Sprintf("toYear(created_at) = %d", year)
	if targetStatus != "" {
		where += fmt.Sprintf(" AND status = '%s'", targetStatus)
	}
	if filterField != "" {
		where += fmt.Sprintf(" AND %s = ?", filterField)
	}

	query := fmt.Sprintf(`
		SELECT toString(toYear(created_at)) as year, formatDateTime(created_at, '%%b') as month, status, count() as total_transactions, sum(amount) as total_amount
		FROM %s WHERE %s
		GROUP BY year, month, status, toMonth(created_at) ORDER BY year, toMonth(created_at), status
	`, table, where)

	args := []interface{}{}
	if filterField != "" {
		args = append(args, filterValue)
	}

	return r.queryMonthlyStatusStats(ctx, query, args...)
}

func (r *ClickHouseReaderRepository) GetYearlyStatusStats(ctx context.Context, table string, filterField string, filterValue interface{}, currentYear int, targetStatus string) ([]YearlyStatusStats, error) {
	where := fmt.Sprintf("(toYear(created_at) = %d OR toYear(created_at) = %d)", currentYear, currentYear-1)
	if targetStatus != "" {
		where += fmt.Sprintf(" AND status = '%s'", targetStatus)
	}
	if filterField != "" {
		where += fmt.Sprintf(" AND %s = ?", filterField)
	}

	query := fmt.Sprintf(`
		SELECT toString(toYear(created_at)) as year, status, count() as total_transactions, sum(amount) as total_amount
		FROM %s WHERE %s
		GROUP BY year, status ORDER BY year DESC, status
	`, table, where)

	args := []interface{}{}
	if filterField != "" {
		args = append(args, filterValue)
	}

	return r.queryYearlyStatusStats(ctx, query, args...)
}

func (r *ClickHouseReaderRepository) GetMonthlyMethodStats(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]MonthlyMethodStats, error) {
	where := fmt.Sprintf("toYear(created_at) = %d AND status = 'success'", year)
	if filterField != "" {
		where += fmt.Sprintf(" AND %s = ?", filterField)
	}

	methodCol := "payment_method"
	if strings.Contains(table, "topup") {
		methodCol = "payment_method"
	}

	query := fmt.Sprintf(`
		SELECT formatDateTime(created_at, '%%b') as month, %s as method, count() as total_transactions, sum(amount) as total_amount
		FROM %s WHERE %s
		GROUP BY month, method, toMonth(created_at) ORDER BY toMonth(created_at), method
	`, methodCol, table, where)

	args := []interface{}{}
	if filterField != "" {
		args = append(args, filterValue)
	}

	return r.queryMonthlyMethodStats(ctx, query, args...)
}

func (r *ClickHouseReaderRepository) GetYearlyMethodStats(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]YearlyMethodStats, error) {
	where := fmt.Sprintf("toYear(created_at) >= %d AND toYear(created_at) <= %d AND status = 'success'", startYear, endYear)
	if filterField != "" {
		where += fmt.Sprintf(" AND %s = ?", filterField)
	}

	methodCol := "payment_method"
	query := fmt.Sprintf(`
		SELECT toString(toYear(created_at)) as year, %s as method, count() as total_transactions, sum(amount) as total_amount
		FROM %s WHERE %s
		GROUP BY year, method ORDER BY year, method
	`, methodCol, table, where)

	args := []interface{}{}
	if filterField != "" {
		args = append(args, filterValue)
	}

	return r.queryYearlyMethodStats(ctx, query, args...)
}

// --- Specialized Saldo Stats ---

func (r *ClickHouseReaderRepository) GetMonthlyTotalSaldo(ctx context.Context, year int) ([]MonthlyAmount, error) {
	query := `
		SELECT 
			toString(toYear(created_at)) as year,
			formatDateTime(created_at, '%%b') as month,
			sum(total_balance) as total_amount
		FROM (
			SELECT card_number, total_balance, created_at,
			ROW_NUMBER() OVER (PARTITION BY card_number, formatDateTime(created_at, '%%Y-%%m') ORDER BY created_at DESC) as rn
			FROM saldo_events
			WHERE toYear(created_at) = ?
		)
		WHERE rn = 1
		GROUP BY year, month, toMonth(created_at)
		ORDER BY year, toMonth(created_at)
	`
	return r.queryMonthlyAmounts(ctx, query, year)
}

func (r *ClickHouseReaderRepository) GetYearlyTotalSaldo(ctx context.Context, startYear, endYear int) ([]YearlyAmount, error) {
	query := `
		SELECT 
			toString(toYear(created_at)) as year,
			sum(total_balance) as total_amount
		FROM (
			SELECT card_number, total_balance, created_at,
			ROW_NUMBER() OVER (PARTITION BY card_number, toString(toYear(created_at)) ORDER BY created_at DESC) as rn
			FROM saldo_events
			WHERE toYear(created_at) >= ? AND toYear(created_at) <= ?
		)
		WHERE rn = 1
		GROUP BY year
		ORDER BY year
	`
	return r.queryYearlyAmounts(ctx, query, startYear, endYear)
}

func (r *ClickHouseReaderRepository) FindMerchantTransactions(ctx context.Context, merchantID int32) ([]map[string]interface{}, error) {
	where := "status = 'success'"
	args := []interface{}{}
	if merchantID != 0 {
		where += " AND merchant_id = ?"
		args = append(args, merchantID)
	}

	query := fmt.Sprintf(`
		SELECT transaction_id, amount, payment_method as method, toString(created_at) as created_at
		FROM transaction_events WHERE %s
		ORDER BY created_at DESC LIMIT 100
	`, where)

	return r.queryMerchantTransactions(ctx, query, args...)
}

func (r *ClickHouseReaderRepository) FindMerchantTransactionsByApikey(ctx context.Context, apiKey string) ([]map[string]interface{}, error) {
	where := "status = 'success' AND apikey = ?"
	query := fmt.Sprintf(`
		SELECT transaction_id, amount, payment_method as method, toString(created_at) as created_at
		FROM transaction_events WHERE %s
		ORDER BY created_at DESC LIMIT 100
	`, where)

	return r.queryMerchantTransactions(ctx, query, apiKey)
}

func (r *ClickHouseReaderRepository) queryMerchantTransactions(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var transactionID uint64
		var amount int64
		var method, createdAt string
		if err := rows.Scan(&transactionID, &amount, &method, &createdAt); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"id":         transactionID,
			"amount":     amount,
			"method":     method,
			"created_at": createdAt,
		})
	}
	return results, nil
}

// --- Internal Helpers ---

func (r *ClickHouseReaderRepository) queryMonthlyAmounts(ctx context.Context, query string, args ...interface{}) ([]MonthlyAmount, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []MonthlyAmount
	for rows.Next() {
		var m MonthlyAmount
		if err := rows.Scan(&m.Year, &m.Month, &m.TotalAmount); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, nil
}

func (r *ClickHouseReaderRepository) queryYearlyAmounts(ctx context.Context, query string, args ...interface{}) ([]YearlyAmount, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []YearlyAmount
	for rows.Next() {
		var y YearlyAmount
		if err := rows.Scan(&y.Year, &y.TotalAmount); err != nil {
			return nil, err
		}
		results = append(results, y)
	}
	return results, nil
}

func (r *ClickHouseReaderRepository) queryMonthlyMethodStats(ctx context.Context, query string, args ...interface{}) ([]MonthlyMethodStats, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []MonthlyMethodStats
	for rows.Next() {
		var m MonthlyMethodStats
		if err := rows.Scan(&m.Month, &m.PaymentMethod, &m.TotalTransactions, &m.TotalAmount); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, nil
}

func (r *ClickHouseReaderRepository) queryMonthlyStatusStats(ctx context.Context, query string, args ...interface{}) ([]MonthlyStatusStats, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []MonthlyStatusStats
	for rows.Next() {
		var m MonthlyStatusStats
		if err := rows.Scan(&m.Year, &m.Month, &m.Status, &m.TotalTransactions, &m.TotalAmount); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, nil
}

func (r *ClickHouseReaderRepository) queryYearlyStatusStats(ctx context.Context, query string, args ...interface{}) ([]YearlyStatusStats, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []YearlyStatusStats
	for rows.Next() {
		var y YearlyStatusStats
		if err := rows.Scan(&y.Year, &y.Status, &y.TotalTransactions, &y.TotalAmount); err != nil {
			return nil, err
		}
		results = append(results, y)
	}
	return results, nil
}

func (r *ClickHouseReaderRepository) queryYearlyMethodStats(ctx context.Context, query string, args ...interface{}) ([]YearlyMethodStats, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []YearlyMethodStats
	for rows.Next() {
		var m YearlyMethodStats
		if err := rows.Scan(&m.Year, &m.PaymentMethod, &m.TotalTransactions, &m.TotalAmount); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, nil
}
