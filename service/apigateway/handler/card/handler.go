package cardhandler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	card_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/card"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	apimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/card"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

type DepsCard struct {
	Client      *grpc.ClientConn
	StatsClient *grpc.ClientConn

	E *echo.Echo

	Logger logger.LoggerInterface

	Cache *cache.CacheStore

	ApiHandler errors.ApiHandler
}

func RegisterCardHandler(deps *DepsCard) {
	mapper := apimapper.NewCardResponseMapper()
	cache := card_cache.NewCardMencache(deps.Cache)

	handlers := []func(){
		setupCardQueryHandler(deps, mapper.QueryMapper(), cache),
		setupCardCommandHandler(deps, mapper.CommandMapper(), cache),
		setupCardDashboardHandler(deps, mapper.DashboardMapper(), cache),
		setupCardStatsHandler(deps),
	}

	for _, h := range handlers {
		h()
	}
}

func setupCardQueryHandler(deps *DepsCard, mapper apimapper.CardQueryResponseMapper, cache card_cache.CardMencache) func() {
	return func() {
		NewCardQueryHandleApi(&cardQueryHandleApiDeps{
			client:     pb.NewCardQueryServiceClient(deps.Client),
			router:     deps.E,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}

func setupCardCommandHandler(deps *DepsCard, mapper apimapper.CardCommandResponseMapper, cache card_cache.CardMencache) func() {
	return func() {
		NewCardCommandHandleApi(&cardCommandHandleApiDeps{
			client:     pb.NewCardCommandServiceClient(deps.Client),
			router:     deps.E,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}

func setupCardDashboardHandler(deps *DepsCard, mapper apimapper.CardDashboardResponseMapper, cache card_cache.CardMencache) func() {
	return func() {
		NewCardDashboardHandleApi(&cardDashboardHandleApiDeps{
			router:     deps.E,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}
