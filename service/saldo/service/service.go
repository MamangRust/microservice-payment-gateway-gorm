package service

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	mencache "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/redis"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
)

type service struct {
	SaldoQueryService
	SaldoCommandService
}

type Service interface {
	SaldoQueryService
	SaldoCommandService
}

type Deps struct {
	Repositories repository.Repositories
	CardAdapter  adapter.CardAdapter
	Logger       logger.LoggerInterface
	Cache        *cache.CacheStore
	Kafka        *kafka.Kafka
}

func NewService(deps *Deps) Service {
	observability, _ := observability.NewObservability("saldo-service", deps.Logger)
	cache := mencache.NewMencache(deps.Cache)

	return &service{
		SaldoQueryService:   newSaldoQueryService(deps, observability, cache),
		SaldoCommandService: newSaldoCommandService(deps, observability, cache),
	}
}

func newSaldoQueryService(deps *Deps, observabilty observability.TraceLoggerObservability, cache mencache.Mencache) SaldoQueryService {
	return NewSaldoQueryService(&saldoQueryParams{
		Cache:         cache,
		Repository:    deps.Repositories,
		Logger:        deps.Logger,
		Observability: observabilty,
	})
}

func newSaldoCommandService(deps *Deps, observabilty observability.TraceLoggerObservability, cache mencache.Mencache) SaldoCommandService {
	return NewSaldoCommandService(&saldoCommandParams{
		Cache:                  cache,
		saldoCommandRepository: deps.Repositories,
		CardAdapter:            deps.CardAdapter,
		Logger:                 deps.Logger,
		Observability:          observabilty,
		Kafka:                  deps.Kafka,
	})
}
