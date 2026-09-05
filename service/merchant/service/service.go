package service

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"

	mencache "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/redis"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
)

// Service exposes all merchant-related domain services.
type Service interface {
	MerchantQueryService() MerchantQueryService
	MerchantTransactionService() MerchantTransactionService
	MerchantCommandService() MerchantCommandService
	MerchantDocumentCommandService() MerchantDocumentCommandService
	MerchantDocumentQueryService() MerchantDocumentQueryService
}

type service struct {
	merchantQuery           MerchantQueryService
	merchantTransaction     MerchantTransactionService
	merchantCommand         MerchantCommandService
	merchantDocumentCommand MerchantDocumentCommandService
	merchantDocumentQuery   MerchantDocumentQueryService
}

// Deps holds shared dependencies for merchant services.
type Deps struct {
	Kafka        *kafka.Kafka
	Repositories repository.Repositories
	UserAdapter  adapter.UserAdapter
	Logger       logger.LoggerInterface
	Cache        *cache.CacheStore
}

// NewService wires and initializes all merchant services.
func NewService(deps *Deps) Service {
	observability, _ := observability.NewObservability("merchant-service", deps.Logger)
	cache := mencache.NewMencache(deps.Cache)

	return &service{
		merchantQuery:           newMerchantQueryService(deps, observability, cache),
		merchantTransaction:     newMerchantTransactionService(deps, observability, cache),
		merchantCommand:         newMerchantCommandService(deps, observability, cache),
		merchantDocumentCommand: newMerchantDocumentCommandService(deps, observability, cache),
		merchantDocumentQuery:   newMerchantDocumentQueryService(deps, observability, cache),
	}
}

func (s *service) MerchantQueryService() MerchantQueryService {
	return s.merchantQuery
}
func (s *service) MerchantTransactionService() MerchantTransactionService {
	return s.merchantTransaction
}
func (s *service) MerchantCommandService() MerchantCommandService {
	return s.merchantCommand
}
func (s *service) MerchantDocumentCommandService() MerchantDocumentCommandService {
	return s.merchantDocumentCommand
}
func (s *service) MerchantDocumentQueryService() MerchantDocumentQueryService {
	return s.merchantDocumentQuery
}

func newMerchantQueryService(
	deps *Deps,
	observability observability.TraceLoggerObservability,
	cache mencache.Mencache,
) MerchantQueryService {
	return NewMerchantQueryService(&merchantQueryDeps{
		Repository:    deps.Repositories,
		Cache:         cache,
		Logger:        deps.Logger,
		Observability: observability,
	})
}

func newMerchantDocumentQueryService(
	deps *Deps,
	observability observability.TraceLoggerObservability,
	cache mencache.Mencache,
) MerchantDocumentQueryService {
	return NewMerchantDocumentQueryService(&merchantDocumentQueryDeps{
		Repository:    deps.Repositories,
		Cache:         cache,
		Logger:        deps.Logger,
		Observability: observability,
	})
}

func newMerchantTransactionService(
	deps *Deps,
	observability observability.TraceLoggerObservability,
	cache mencache.Mencache,
) MerchantTransactionService {
	return NewMerchantTransactionService(&merchantTransactionDeps{
		Repository:    deps.Repositories,
		Cache:         cache,
		Logger:        deps.Logger,
		Observability: observability,
	})
}

func newMerchantCommandService(
	deps *Deps,
	observability observability.TraceLoggerObservability,
	cache mencache.Mencache,
) MerchantCommandService {
	return NewMerchantCommandService(&merchantCommandServiceDeps{
		Kafka:                     deps.Kafka,
		UserAdapter:               deps.UserAdapter,
		MerchantQueryRepository:   deps.Repositories,
		MerchantCommandRepository: deps.Repositories,
		Logger:                    deps.Logger,
		Observability:             observability,
		Cache:                     cache,
	})
}

func newMerchantDocumentCommandService(
	deps *Deps,
	observability observability.TraceLoggerObservability,
	cache mencache.Mencache,
) MerchantDocumentCommandService {
	return NewMerchantDocumentCommandService(&merchantDocumentCommandDeps{
		Kafka:                   deps.Kafka,
		CommandRepository:       deps.Repositories,
		MerchantQueryRepository: deps.Repositories,
		UserAdapter:             deps.UserAdapter,
		Logger:                  deps.Logger,
		Observability:           observability,
		Cache:                   cache,
	})
}
