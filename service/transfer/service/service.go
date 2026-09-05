package service

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	mencache "github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/redis"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/repository"
)

type Service interface {
	TransferQueryService
	TransferCommandService
}

type service struct {
	TransferQueryService
	TransferCommandService
}

type Deps struct {
	Kafka            *kafka.Kafka
	Cache            *cache.CacheStore
	Repositories     repository.Repositories
	CardAdapter      adapter.CardAdapter
	SaldoAdapter     adapter.SaldoAdapter
	Logger           logger.LoggerInterface
	AISecurityClient ai_security.AISecurityServiceClient
}

func NewService(deps *Deps) Service {
	cache := mencache.NewMencache(deps.Cache)

	observability, _ := observability.NewObservability("transfer-service", deps.Logger)

	return &service{
		TransferQueryService:   newTransferQueryService(deps, observability, cache),
		TransferCommandService: newTransferCommandService(deps, observability, cache),
	}
}

func newTransferQueryService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) TransferQueryService {
	return NewTransferQueryService(
		&transferQueryDeps{
			Cache:         cache,
			Repository:    deps.Repositories,
			Logger:        deps.Logger,
			Observability: observability,
		},
	)
}

func newTransferCommandService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) TransferCommandService {
	return NewTransferCommandService(
		&transferCommandDeps{
			Kafka:                     deps.Kafka,
			Cache:                     cache,
			CardAdapter:               deps.CardAdapter,
			SaldoAdapter:              deps.SaldoAdapter,
			TransferQueryRepository:   deps.Repositories,
			TransferCommandRepository: deps.Repositories,
			IdempotencyStore:          deps.Repositories,
			OutboxStore:               deps.Repositories,
			Logger:                    deps.Logger,
			Observability:             observability,
			AISecurityClient:          deps.AISecurityClient,
		},
	)
}
