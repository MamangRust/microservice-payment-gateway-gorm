package adapter

import (
	"context"
	"time"

	pbmerchant "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
)

type MerchantAdapter interface {
	FindByApiKey(ctx context.Context, api_key string) (*models.Merchant, error)
	FindByMerchantId(ctx context.Context, merchant_id int) (*models.Merchant, error)
}

type merchantGRPCAdapter struct {
	QueryClient pbmerchant.MerchantQueryServiceClient
	guard       *resilience.DependencyGuard
}

func (a *merchantGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func NewMerchantAdapter(queryClient pbmerchant.MerchantQueryServiceClient, opts ...GuardOption) MerchantAdapter {
	a := &merchantGRPCAdapter{
		QueryClient: queryClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *merchantGRPCAdapter) FindByApiKey(ctx context.Context, api_key string) (*models.Merchant, error) {
	var resp *pbmerchant.ApiResponseMerchant
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindByApiKey(callCtx, &pbmerchant.FindByApiKeyRequest{
			ApiKey: api_key,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}

	parseTime := func(ts string) *time.Time {
		if ts == "" {
			return nil
		}
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			return nil
		}
		return &t
	}

	return &models.Merchant{
		MerchantID: resp.Data.Id,
		Name:       resp.Data.Name,
		ApiKey:     resp.Data.ApiKey,
		UserID:     resp.Data.UserId,
		Status:     resp.Data.Status,
		CreatedAt:  parseTime(resp.Data.CreatedAt),
		UpdatedAt:  parseTime(resp.Data.UpdatedAt),
	}, nil
}

func (a *merchantGRPCAdapter) FindByMerchantId(ctx context.Context, merchant_id int) (*models.Merchant, error) {
	var resp *pbmerchant.ApiResponseMerchant
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindByIdMerchant(callCtx, &pbmerchant.FindByIdMerchantRequest{
			MerchantId: int32(merchant_id),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}

	parseTime := func(ts string) *time.Time {
		if ts == "" {
			return nil
		}
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			return nil
		}
		return &t
	}

	return &models.Merchant{
		MerchantID: resp.Data.Id,
		Name:       resp.Data.Name,
		ApiKey:     resp.Data.ApiKey,
		Status:     resp.Data.Status,
		UserID:     resp.Data.UserId,
		CreatedAt:  parseTime(resp.Data.CreatedAt),
		UpdatedAt:  parseTime(resp.Data.UpdatedAt),
	}, nil
}
