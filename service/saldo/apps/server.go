package apps

import (
	"context"
	"fmt"
	"time"

	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/server"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/handler"
	saldokafka "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/redis"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/service"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	// gRPC Clients for cross-service communication
	connCard, err := grpc.NewClient(viper.GetString("GRPC_CARD_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Card service: %w", err)
	}

	cardClientQuery := pbcard.NewCardQueryServiceClient(connCard)
	cardClientCmd := pbcard.NewCardCommandServiceClient(connCard)
	cardGuard := resilience.NewDependencyGuard("card", 5, 30, 100, 3*time.Second, srv.Logger)
	cardAdapter := adapter.NewCardAdapter(cardClientQuery, cardClientCmd, adapter.WithDependencyGuard(cardGuard))

	repos := repository.NewRepositories(srv.GormDB, cardAdapter)

	mykafka, err := kafka.NewKafka(srv.Logger, []string{viper.GetString("KAFKA_BROKERS")})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	svc := service.NewService(&service.Deps{
		Cache:        srv.CacheStore,
		Logger:       srv.Logger,
		Repositories: repos,
		CardAdapter:  cardAdapter,
		Kafka:        mykafka,
	})

	kafkaHandler := saldokafka.NewSaldoKafkaHandler(svc, srv.Logger, context.Background())
	err = mykafka.StartConsumers([]string{"saldo-service-topic-create-saldo"}, "saldo-service-group", kafkaHandler)
	if err != nil {
		srv.Logger.Error("Failed to start kafka consumers", zap.Error(err))
	}

	saldoCache := mencache.NewMencache(srv.CacheStore)
	cacheInvalidator := saldokafka.NewSaldoCacheInvalidator(saldoCache, srv.Logger)
	if err := mykafka.StartConsumers([]string{"stats-topic-saldo-events"}, "saldo-cache-invalidator", cacheInvalidator); err != nil {
		srv.Logger.Error("Failed to start saldo cache invalidator consumer", zap.Error(err))
	}

	h := handler.NewHandler(svc)

	srv.RegisterServices = func(gs *grpc.Server) {
		pb.RegisterSaldoQueryServiceServer(gs, h)
		pb.RegisterSaldoCommandServiceServer(gs, h)
	}

	return srv, nil
}
