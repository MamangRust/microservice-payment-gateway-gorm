package adapter

import (
	"context"
	"time"

	pbtransaction "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
)

// TransactionAdapter is the contract consumers depend on for reading
// transactions owned by the transaction service. It wraps the generated
// transaction gRPC query client so cross-owner reads go through the network
// boundary instead of a direct SQL connection.
type TransactionAdapter interface {
	FindAllTransactions(ctx context.Context, page, pageSize int, search string) ([]*models.Transaction, *int, error)
	FindAllTransactionsByMerchantId(ctx context.Context, merchantID, page, pageSize int, search string) ([]*models.Transaction, *int, error)
}

type transactionGRPCAdapter struct {
	QueryClient pbtransaction.TransactionQueryServiceClient
	guard       *resilience.DependencyGuard
}

func (a *transactionGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

// NewTransactionAdapter wraps the generated transaction gRPC query client.
func NewTransactionAdapter(queryClient pbtransaction.TransactionQueryServiceClient, opts ...GuardOption) TransactionAdapter {
	a := &transactionGRPCAdapter{
		QueryClient: queryClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *transactionGRPCAdapter) FindAllTransactions(ctx context.Context, page, pageSize int, search string) ([]*models.Transaction, *int, error) {
	var resp *pbtransaction.ApiResponsePaginationTransaction
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindAllTransaction(callCtx, &pbtransaction.FindAllTransactionRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
			Search:   search,
		})
		return callErr
	})
	if err != nil {
		return nil, nil, err
	}
	if resp == nil {
		return nil, nil, nil
	}
	return mapTransactionResponses(resp.Data), intPtr(int(resp.PaginationMeta.GetTotalRecords())), nil
}

func (a *transactionGRPCAdapter) FindAllTransactionsByMerchantId(ctx context.Context, merchantID, page, pageSize int, search string) ([]*models.Transaction, *int, error) {
	var resp *pbtransaction.ApiResponsePaginationTransaction
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindAllTransactionByMerchantId(callCtx, &pbtransaction.FindAllTransactionByMerchantIdRequest{
			MerchantId: int32(merchantID),
			Page:       int32(page),
			PageSize:   int32(pageSize),
			Search:     search,
		})
		return callErr
	})
	if err != nil {
		return nil, nil, err
	}
	if resp == nil {
		return nil, nil, nil
	}
	return mapTransactionResponses(resp.Data), intPtr(int(resp.PaginationMeta.GetTotalRecords())), nil
}

func mapTransactionResponses(rows []*pbtransaction.TransactionResponse) []*models.Transaction {
	out := make([]*models.Transaction, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		tx := &models.Transaction{
			TransactionID:   r.Id,
			TransactionNo:   r.TransactionNo,
			CardNumber:      r.CardNumber,
			Amount:          r.Amount,
			PaymentMethod:   r.PaymentMethod,
			MerchantID:      r.MerchantId,
			TransactionTime: parseTimeSafe(r.TransactionTime),
			CreatedAt:       parseTime(r.CreatedAt),
			UpdatedAt:       parseTime(r.UpdatedAt),
		}
		if r.DeletedAt != "" {
			if t, perr := time.Parse(time.RFC3339, r.DeletedAt); perr == nil {
				tx.DeletedAt = &t
			}
		}
		out = append(out, tx)
	}
	return out
}

func parseTimeSafe(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func intPtr(v int) *int {
	return &v
}
