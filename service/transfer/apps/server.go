package apps

import (
	"fmt"
	"strings"
	"time"

	pbai "github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbsaldo "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/server"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/outbox"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	// gRPC Clients for cross-service communication
	connSaldo, err := grpc.NewClient(viper.GetString("GRPC_SALDO_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Saldo service: %w", err)
	}

	connCard, err := grpc.NewClient(viper.GetString("GRPC_CARD_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Card service: %w", err)
	}

	connAI, err := grpc.NewClient(viper.GetString("GRPC_AI_SECURITY_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AI Security service: %w", err)
	}

	saldoClientQuery := pbsaldo.NewSaldoQueryServiceClient(connSaldo)
	saldoClientCmd := pbsaldo.NewSaldoCommandServiceClient(connSaldo)
	cardClientQuery := pbcard.NewCardQueryServiceClient(connCard)
	cardClientCmd := pbcard.NewCardCommandServiceClient(connCard)
	aiClient := pbai.NewAISecurityServiceClient(connAI)

	saldoGuard := resilience.NewDependencyGuard("saldo", 5, 30, 100, 3*time.Second, srv.Logger)
	cardGuard := resilience.NewDependencyGuard("card", 5, 30, 100, 3*time.Second, srv.Logger)
	aiSecurityGuard := resilience.NewDependencyGuard("ai_security", 5, 30, 100, 3*time.Second, srv.Logger)

	saldoAdapter := adapter.NewSaldoAdapter(saldoClientQuery, saldoClientCmd, adapter.WithDependencyGuard(saldoGuard))
	cardAdapter := adapter.NewCardAdapter(cardClientQuery, cardClientCmd, adapter.WithDependencyGuard(cardGuard))
	aiSecurityAdapter := adapter.NewAISecurityAdapter(aiClient, adapter.WithDependencyGuard(aiSecurityGuard))

	repos := repository.NewRepositories(
		srv.GormDB,
		saldoClientQuery,
		saldoClientCmd,
		cardClientQuery,
		cardClientCmd,
		repository.GuardOptions{
			Saldo: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("saldo", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
			Card: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("card", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
		},
	)

	kafkaBrokers := strings.Split(viper.GetString("KAFKA_BROKERS"), ",")
	myKafka, err := kafka.NewKafka(srv.Logger, kafkaBrokers)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	svc := service.NewService(&service.Deps{
		Kafka:             myKafka,
		Cache:             srv.CacheStore,
		Logger:            srv.Logger,
		Repositories:      repos,
		AISecurityAdapter: aiSecurityAdapter,
		CardAdapter:       cardAdapter,
		SaldoAdapter:      saldoAdapter,
	})
	svc.StartRecoveryWorker(srv.Ctx, 30*time.Second, 2*time.Minute, 100)

	outboxPub := outbox.NewPublisher(srv.GormDB, myKafka, srv.Logger)
	go outboxPub.Start(srv.Ctx)

	h := handler.NewHandler(svc)

	srv.RegisterServices = func(gs *grpc.Server) {
		pb.RegisterTransferQueryServiceServer(gs, h)
		pb.RegisterTransferCommandServiceServer(gs, h)
	}

	return srv, nil
}
