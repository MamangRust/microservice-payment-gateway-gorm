package transferhandler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	transfer_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/transfer"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	apimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/transfer"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

type DepsTransfer struct {
	Client      *grpc.ClientConn
	StatsClient *grpc.ClientConn

	E *echo.Echo

	Logger logger.LoggerInterface

	Cache *cache.CacheStore

	ApiHandler errors.ApiHandler
}

func RegisterTransferHandler(deps *DepsTransfer) {
	mapper := apimapper.NewTransferResponseMapper()
	cache := transfer_cache.NewTransferMencache(deps.Cache)

	handlers := []func(){
		setupTransferQueryHandler(deps, mapper.QueryMapper(), cache),
		setupTransferCommandHandler(deps, mapper.CommandMapper(), cache),
		setupTransferStatsHandler(deps),
	}

	for _, h := range handlers {
		h()
	}
}

func setupTransferQueryHandler(deps *DepsTransfer, mapper apimapper.TransferQueryResponseMapper, cache transfer_cache.TransferMencache) func() {
	return func() {
		NewTransferQueryHandleApi(&transferQueryHandleDeps{
			client:     pb.NewTransferQueryServiceClient(deps.Client),
			router:     deps.E,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}

func setupTransferCommandHandler(deps *DepsTransfer, mapper apimapper.TransferCommandResponseMapper, cache transfer_cache.TransferMencache) func() {
	return func() {
		NewTransferCommandHandleApi(&transferCommandHandleDeps{
			client:     pb.NewTransferCommandServiceClient(deps.Client),
			router:     deps.E,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}
