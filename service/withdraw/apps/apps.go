package apps

import (
	"fmt"
	"strings"
	"time"

	pbai "github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbsaldo "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/server"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/service"
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

	saldoAdapter := adapter.NewSaldoAdapter(saldoClientQuery, saldoClientCmd, adapter.WithDependencyGuard(saldoGuard))
	cardAdapter := adapter.NewCardAdapter(cardClientQuery, cardClientCmd, adapter.WithDependencyGuard(cardGuard))

	repos := repository.NewRepositories(srv.GormDB, cardAdapter, saldoAdapter)
	kafkaBrokers := strings.Split(viper.GetString("KAFKA_BROKERS"), ",")
	myKafka, err := kafka.NewKafka(srv.Logger, kafkaBrokers)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Daily withdrawal limit (in smallest currency unit); defaults to 10.000.000.
	dailyWithdrawalLimit := viper.GetInt64("WITHDRAW_DAILY_LIMIT")
	if dailyWithdrawalLimit == 0 {
		dailyWithdrawalLimit = 10_000_000
	}
	if dailyWithdrawalLimit < 0 {
		return nil, fmt.Errorf("WITHDRAW_DAILY_LIMIT must be >= 0, got %d", dailyWithdrawalLimit)
	}

	svc := service.NewService(&service.Deps{
		Kafka:                myKafka,
		Repositories:         repos,
		Logger:               srv.Logger,
		Cache:                srv.CacheStore,
		AISecurityClient:     aiClient,
		CardAdapter:          cardAdapter,
		SaldoAdapter:         saldoAdapter,
		DailyWithdrawalLimit: dailyWithdrawalLimit,
	})

	outboxPub := outbox.NewPublisher(srv.GormDB, myKafka, srv.Logger)
	go outboxPub.Start(srv.Ctx)

	h := handler.NewHandler(svc)

	srv.RegisterServices = func(gs *grpc.Server) {
		pb.RegisterWithdrawQueryServiceServer(gs, h)
		pb.RegisterWithdrawCommandServiceServer(gs, h)
	}

	return srv, nil
}
