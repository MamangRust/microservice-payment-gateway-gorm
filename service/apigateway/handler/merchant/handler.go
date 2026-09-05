package merchanthandler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	merchant_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/merchant"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	apimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/merchant"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

type DepsMerchant struct {
	Client      *grpc.ClientConn
	StatsClient *grpc.ClientConn
	E           *echo.Echo
	Logger      logger.LoggerInterface
	Cache       *cache.CacheStore

	ApiHandler errors.ApiHandler
}

func RegisterMerchantHandler(deps *DepsMerchant) {
	mapper := apimapper.NewMerchantResponseMapper()

	cache := merchant_cache.NewMerchantMencache(deps.Cache)

	handlers := []func(){
		setupMerchantQueryHandler(deps, mapper.QueryMapper(), cache),
		setupMerchantCommandHandler(deps, mapper.CommandMapper(), cache),
		setupMerchantTransactionHandler(deps, mapper.TransactionMapper(), cache),
		setupMerchantStatsHandler(deps),
	}

	for _, h := range handlers {
		h()
	}
}

func setupMerchantQueryHandler(deps *DepsMerchant, mapper apimapper.MerchantQueryResponseMapper, cache merchant_cache.MerchantMencache) func() {
	return func() {
		NewMerchantQueryHandleApi(&merchantQueryHandleDeps{
			client:     pb.NewMerchantQueryServiceClient(deps.Client),
			router:     deps.E,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}

func setupMerchantCommandHandler(deps *DepsMerchant, mapper apimapper.MerchantCommandResponseMapper, cache merchant_cache.MerchantMencache) func() {
	return func() {
		NewMerchantCommandHandleApi(&merchantCommandHandleDeps{
			client:     pb.NewMerchantCommandServiceClient(deps.Client),
			router:     deps.E,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}

func setupMerchantTransactionHandler(deps *DepsMerchant, mapper apimapper.MerchantTransactionResponseMapper, cache merchant_cache.MerchantMencache) func() {
	return func() {
		NewMerchantTransactionHandleApi(&merchantTransactionHandleDeps{
			client:     pb.NewMerchantTransactionServiceClient(deps.Client),
			router:     deps.E,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}
