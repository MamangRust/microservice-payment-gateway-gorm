package saldohandler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	saldo_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	apimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/saldo"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

type DepsSaldo struct {
	Client      *grpc.ClientConn
	StatsClient *grpc.ClientConn
	E           *echo.Echo

	Logger logger.LoggerInterface

	Cache *cache.CacheStore

	ApiHandler errors.ApiHandler
}

func RegisterSaldoHandler(deps *DepsSaldo) {
	mapper := apimapper.NewSaldoResponseMapper()

	cache := saldo_cache.NewSaldoMencache(deps.Cache)

	handlers := []func(){
		setupSaldoQueryHandler(deps, mapper.QueryMapper(), cache),
		setupSaldoCommandHandler(deps, mapper.CommandMapper(), cache),
		setupSaldoStatsHandler(deps),
	}

	for _, h := range handlers {
		h()
	}
}

func setupSaldoQueryHandler(deps *DepsSaldo, mapper apimapper.SaldoQueryResponseMapper, cache saldo_cache.SaldoMencache) func() {
	return func() {
		NewSaldoQueryHandleApi(
			&saldoQueryHandleDeps{
				client:     pb.NewSaldoQueryServiceClient(deps.Client),
				router:     deps.E,
				logger:     deps.Logger,
				mapper:     mapper,
				cache:      cache,
				apiHandler: deps.ApiHandler,
			},
		)
	}
}

func setupSaldoCommandHandler(deps *DepsSaldo, mapper apimapper.SaldoCommandResponseMapper, cache saldo_cache.SaldoMencache) func() {
	return func() {
		NewSaldoCommandHandleApi(
			&saldoCommandHandleDeps{
				client:     pb.NewSaldoCommandServiceClient(deps.Client),
				router:     deps.E,
				logger:     deps.Logger,
				mapper:     mapper,
				apiHandler: deps.ApiHandler,
				cache:      cache,
			},
		)
	}
}
