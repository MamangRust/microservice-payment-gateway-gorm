package apps

import (
	"fmt"
	"strings"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/server"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/service"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	// Establish GRPC connection to User service
	userConn, err := grpc.NewClient(viper.GetString("GRPC_USER_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to User service: %w", err)
	}

	userQueryClient := user.NewUserQueryServiceClient(userConn)
	userGuard := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, srv.Logger)
	userAdapter := adapter.NewUserAdapter(userQueryClient, adapter.WithDependencyGuard(userGuard))

	kafkaBrokers := strings.Split(viper.GetString("KAFKA_BROKERS"), ",")
	mykafka, err := kafka.NewKafka(srv.Logger, kafkaBrokers)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	repos := repository.NewRepositories(srv.GormDB, userQueryClient)
	billingCycleDay := viper.GetInt("BILLING_CYCLE_DAY")
	if billingCycleDay == 0 {
		billingCycleDay = 1
	}
	if billingCycleDay < 1 || billingCycleDay > 31 {
		return nil, fmt.Errorf("BILLING_CYCLE_DAY must be between 1 and 31, got %d", billingCycleDay)
	}

	svc := service.NewService(&service.Deps{
		Cache:           srv.CacheStore,
		Logger:          srv.Logger,
		Repositories:    repos,
		UserAdapter:     userAdapter,
		Kafka:           mykafka,
		BillingCycleDay: billingCycleDay,
	})
	h := handler.NewHandler(svc)

	srv.RegisterServices = func(gs *grpc.Server) {
		card.RegisterCardQueryServiceServer(gs, h)
		card.RegisterCardCommandServiceServer(gs, h)
	}

	return srv, nil
}
