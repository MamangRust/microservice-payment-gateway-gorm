package service

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	mencache "github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/redis"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
)

type Service interface {
	TopupQueryService
	TopupCommandService
}

type service struct {
	TopupQueryService
	TopupCommandService
}

type Deps struct {
	Kafka        *kafka.Kafka
	Cache        *cache.CacheStore
	Repositories repository.Repositories
	CardAdapter  adapter.CardAdapter
	SaldoAdapter adapter.SaldoAdapter
	Logger       logger.LoggerInterface
}

func NewService(deps *Deps) Service {
	cache := mencache.NewMencache(deps.Cache)

	observability, _ := observability.NewObservability("topup-service", deps.Logger)

	return &service{
		TopupQueryService:   newTopupQueryService(deps, observability, cache),
		TopupCommandService: newTopupCommandService(deps, observability, cache),
	}
}

func newTopupQueryService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.TopupQueryCache) TopupQueryService {
	return NewTopupQueryService(&topupQueryDeps{
		Cache:         cache,
		Repository:    deps.Repositories,
		Logger:        deps.Logger,
		Observability: observability,
	})
}

func newTopupCommandService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.TopupCommandCache) TopupCommandService {
	return NewTopupCommandService(&topupCommandDeps{
		Kafka:                  deps.Kafka,
		Cache:                  cache,
		CardAdapter:            deps.CardAdapter,
		TopupQueryRepository:   deps.Repositories,
		TopupCommandRepository: deps.Repositories,
		SaldoAdapter:           deps.SaldoAdapter,
		IdempotencyStore:       deps.Repositories,
		OutboxStore:            deps.Repositories,
		Logger:                 deps.Logger,
		Observability:          observability,
	})
}
