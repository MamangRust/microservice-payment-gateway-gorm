package service

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	mencache "github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/redis"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
)

type Service interface {
	WithdrawQueryService
	WithdrawCommandService
}

type service struct {
	WithdrawQueryService
	WithdrawCommandService
}

type Deps struct {
	Kafka                *kafka.Kafka
	Repositories         repository.Repositories
	CardAdapter          adapter.CardAdapter
	SaldoAdapter         adapter.SaldoAdapter
	Logger               logger.LoggerInterface
	Cache                *cache.CacheStore
	AISecurityClient     ai_security.AISecurityServiceClient
	DailyWithdrawalLimit int64
}

func NewService(deps *Deps) Service {

	cache := mencache.NewMencache(deps.Cache)

	observability, _ := observability.NewObservability("withdraw-service", deps.Logger)

	return &service{
		WithdrawQueryService:   newWithdrawQueryService(deps, observability, cache),
		WithdrawCommandService: newWithdrawCommandService(deps, observability, cache),
	}
}

func newWithdrawQueryService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) WithdrawQueryService {
	return NewWithdrawQueryService(
		&withdrawQueryServiceDeps{
			Cache:         cache,
			Repository:    deps.Repositories,
			Logger:        deps.Logger,
			Observability: observability,
		},
	)
}

func newWithdrawCommandService(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) WithdrawCommandService {
	return NewWithdrawCommandService(
		&		withdrawCommandServiceDeps{
			Cache:                cache,
			Kafka:                deps.Kafka,
			CardAdapter:          deps.CardAdapter,
			SaldoAdapter:         deps.SaldoAdapter,
			CommandRepository:    deps.Repositories,
			QueryRepository:      deps.Repositories,
			IdempotencyStore:     deps.Repositories,
			OutboxStore:          deps.Repositories,
			Logger:               deps.Logger,
			Observability:        observability,
			AISecurityClient:     deps.AISecurityClient,
			DailyWithdrawalLimit: deps.DailyWithdrawalLimit,
		},
	)
}
