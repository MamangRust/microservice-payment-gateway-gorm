package withdrawhandler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	withdraw_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/withdraw"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	apimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/withdraw"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

type DepsWithdraw struct {
	Client      *grpc.ClientConn
	StatsClient *grpc.ClientConn

	E *echo.Echo

	Logger logger.LoggerInterface

	Cache *cache.CacheStore

	ApiHandler errors.ApiHandler
}

func RegisterWithdrawHandler(deps *DepsWithdraw) {
	if deps.Client == nil {
		panic("RegisterWithdrawHandler: deps.Client is nil")
	}
	mapper := apimapper.NewWithdrawResponseMapper()

	cache := withdraw_cache.NewWithdrawMencache(deps.Cache)

	handlers := []func(){
		setupWithdrawQueryHandler(deps, mapper.QueryMapper(), cache),
		setupWithdrawCommandHandler(deps, mapper.CommandMapper(), cache),
		setupWithdrawStatsHandler(deps),
	}

	for _, h := range handlers {
		h()
	}
}

func setupWithdrawQueryHandler(deps *DepsWithdraw, mapper apimapper.WithdrawQueryResponseMapper, cache withdraw_cache.WithdrawMencache) func() {
	return func() {
		NewWithdrawQueryHandleApi(&withdrawQueryHandleDeps{
			client:     pb.NewWithdrawQueryServiceClient(deps.Client),
			router:     deps.E,
			logger:     deps.Logger,
			mapper:     mapper,
			apiHandler: deps.ApiHandler,
			cache:      cache,
		})
	}
}

func setupWithdrawCommandHandler(deps *DepsWithdraw, mapper apimapper.WithdrawCommandResponseMapper, cache withdraw_cache.WithdrawMencache) func() {
	return func() {
		NewWithdrawCommandHandleApi(&withdrawCommandHandleDeps{
			client:     pb.NewWithdrawCommandServiceClient(deps.Client),
			router:     deps.E,
			logger:     deps.Logger,
			mapper:     mapper,
			apiHandler: deps.ApiHandler,
			cache:      cache,
		})
	}
}
