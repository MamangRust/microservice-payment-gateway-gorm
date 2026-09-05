package service

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	mencache "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/redis"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
)

type Service interface {
	CardQueryService
	CardCommandService
	BillingEngineService
}

type service struct {
	CardQueryService
	CardCommandService
	BillingEngineService
}

type Deps struct {
	Cache           *cache.CacheStore
	Repositories    *repository.Repositories
	UserAdapter     adapter.UserAdapter
	Logger          logger.LoggerInterface
	Kafka           *kafka.Kafka
	BillingCycleDay int
}

func NewService(deps *Deps) Service {
	observability, _ := observability.NewObservability("card-server", deps.Logger)

	cache := mencache.NewMencache(deps.Cache)
	billingEngine := newBillingEngine(deps, observability)

	return &service{
		CardQueryService:     newCardQuery(deps, observability, cache),
		CardCommandService:   newCardCommand(deps, observability, cache, billingEngine, deps.BillingCycleDay),
		BillingEngineService: billingEngine,
	}
}

// newCardQuery initializes a new instance of the CardQueryService.
// It takes a pointer to Deps and a mapper for CardResponse.
// It returns a pointer to CardQueryService.
func newCardQuery(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache) CardQueryService {
	return NewCardQueryService(&cardQueryServiceDeps{
		Cache:               cache,
		CardQueryRepository: deps.Repositories.CardQuery,
		UserRepository:      deps.Repositories.User,
		Logger:              deps.Logger,
		Observability:       observability,
	})
}

// newCardCommand initializes a new instance of the CardCommandService.
// It takes a pointer to Deps and a mapper for CardResponse.
// It returns a pointer to CardCommandService.
func newCardCommand(deps *Deps, observability observability.TraceLoggerObservability, cache mencache.Mencache, billingEngine BillingEngineService, billingCycleDay int) CardCommandService {
	return NewCardCommandService(&cardCommandServiceDeps{
		Cache:                 cache,
		Kafka:                 deps.Kafka,
		UserAdapter:           deps.UserAdapter,
		CardCommandRepository: deps.Repositories.CardCommand,
		BillingEngine:         billingEngine,
		BillingCycleDay:       billingCycleDay,
		Logger:                deps.Logger,
		Observability:         observability,
	})
}

func newBillingEngine(deps *Deps, observability observability.TraceLoggerObservability) BillingEngineService {
	return NewBillingEngineService(&BillingEngineServiceDeps{
		BillingCycleRepository: deps.Repositories.BillingCycle,
		Kafka:                  deps.Kafka,
		Logger:                 deps.Logger,
		Observability:          observability,
	})
}
