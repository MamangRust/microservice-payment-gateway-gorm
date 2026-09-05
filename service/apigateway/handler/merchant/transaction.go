package merchanthandler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	merchant_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/merchant"
	errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	apimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/merchant"
	"github.com/labstack/echo/v4"
)

type merchantTransactionHandleApi struct {
	client pb.MerchantTransactionServiceClient

	logger logger.LoggerInterface
	mapper apimapper.MerchantTransactionResponseMapper

	cache merchant_cache.MerchantMencache

	apiHandler errors.ApiHandler
}

type merchantTransactionHandleDeps struct {
	client pb.MerchantTransactionServiceClient
	router *echo.Echo

	logger logger.LoggerInterface
	mapper apimapper.MerchantTransactionResponseMapper

	cache merchant_cache.MerchantMencache

	apiHandler errors.ApiHandler
}

func NewMerchantTransactionHandleApi(params *merchantTransactionHandleDeps) *merchantTransactionHandleApi {

	merchantHandler := &merchantTransactionHandleApi{
		client:     params.client,
		logger:     params.logger,
		mapper:     params.mapper,
		cache:      params.cache,
		apiHandler: params.apiHandler,
	}

	// routerMerchant := params.router.Group("/api/merchant-transaction")

	// Transaction analytical endpoints are handled by StatsReader
	// This handler is for domain-specific transaction operations if needed

	return merchantHandler
}
