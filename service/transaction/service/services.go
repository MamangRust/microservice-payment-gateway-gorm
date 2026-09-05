package service

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	mencache "github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/redis"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
)

// Service is a struct that contains all the services
type service struct {
	TransactionQueryService
	TransactionCommandService
}

type Service interface {
	TransactionQueryService
	TransactionCommandService
}

type Deps struct {
	Kafka            *kafka.Kafka
	Repositories     repository.Repositories
	MerchantAdapter  adapter.MerchantAdapter
	CardAdapter      adapter.CardAdapter
	SaldoAdapter     adapter.SaldoAdapter
	Logger           logger.LoggerInterface
	Cache            *cache.CacheStore
	AISecurityClient ai_security.AISecurityServiceClient
}

func NewService(deps *Deps) Service {
	cache := mencache.NewMencache(deps.Cache)
	observability, _ := observability.NewObservability("transaction-service", deps.Logger)

	return &service{
		TransactionQueryService:   newTransactionQueryService(deps, observability, cache),
		TransactionCommandService: newTransactionCommandService(deps, observability, cache),
	}
}

func newTransactionQueryService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) TransactionQueryService {
	return NewTransactionQueryService(&transactionQueryServiceDeps{
		Cache:                      cache,
		TransactionQueryRepository: deps.Repositories,
		Logger:                     deps.Logger,
		Observability:              observability,
	})
}

func newTransactionCommandService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) TransactionCommandService {
	return NewTransactionCommandService(&transactionCommandServiceDeps{
		Kafka:                        deps.Kafka,
		Mencache:                     cache,
		MerchantAdapter:              deps.MerchantAdapter,
		CardAdapter:                  deps.CardAdapter,
		SaldoAdapter:                 deps.SaldoAdapter,
		TransactionCommandRepository: deps.Repositories,
		TransactionQueryRepository:   deps.Repositories,
		IdempotencyStore:             deps.Repositories,
		OutboxStore:                  deps.Repositories,
		Logger:                       deps.Logger,
		Observability:                observability,
		AISecurityClient:             deps.AISecurityClient,
	})
}
