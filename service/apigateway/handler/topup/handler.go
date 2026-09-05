package topuphandler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	topup_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/topup"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	apimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/topup"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

type DepsTopup struct {
	Client      *grpc.ClientConn
	StatsClient *grpc.ClientConn

	E *echo.Echo

	Logger logger.LoggerInterface

	Cache *cache.CacheStore

	ApiHandler errors.ApiHandler
}

func RegisterTopupHandler(deps *DepsTopup) {
	mapper := apimapper.NewTopupResponseMapper()

	cache := topup_cache.NewTopupMencache(deps.Cache)

	handlers := []func(){
		setupTopupQueryHandler(deps, mapper.QueryMapper(), cache),
		setupTopupCommandHandler(deps, mapper.CommandMapper(), cache),
		setupTopupStatsHandler(deps),
	}

	for _, h := range handlers {
		h()
	}
}

func setupTopupQueryHandler(deps *DepsTopup, mapper apimapper.TopupQueryResponseMapper, cache topup_cache.TopupMencache) func() {
	return func() {
		NewTopupQueryHandleApi(
			&topupQueryHandleDeps{
				client:     pb.NewTopupQueryServiceClient(deps.Client),
				router:     deps.E,
				logger:     deps.Logger,
				mapper:     mapper,
				cache:      cache,
				apiHandler: deps.ApiHandler,
			},
		)
	}
}

func setupTopupCommandHandler(deps *DepsTopup, mapper apimapper.TopupCommandResponseMapper, cache topup_cache.TopupMencache) func() {
	return func() {
		NewTopupCommandHandleApi(
			&topupCommandHandleDeps{
				client:     pb.NewTopupCommandServiceClient(deps.Client),
				router:     deps.E,
				logger:     deps.Logger,
				mapper:     mapper,
				apiHandler: deps.ApiHandler,
				cache:      cache,
			},
		)
	}
}
