package apps

import (
	"fmt"
	"strings"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/handler"
	myhandlerkafka "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/service"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	pbdocument "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant_document"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/server"
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
	svc := service.NewService(&service.Deps{
		Cache:        srv.CacheStore,
		Logger:       srv.Logger,
		Repositories: repos,
		UserAdapter:  userAdapter,
		Kafka:        mykafka,
	})
	h := handler.NewHandler(svc)

	if err := mykafka.StartConsumers([]string{"request-transaction"}, "merchant-api-key-validator", myhandlerkafka.NewMerchantKafkaHandler(svc.MerchantQueryService(), mykafka, srv.Logger)); err != nil {
		return nil, fmt.Errorf("failed to start merchant API-key validator consumer: %w", err)
	}

	srv.RegisterServices = func(gs *grpc.Server) {
		pb.RegisterMerchantQueryServiceServer(gs, h)
		pb.RegisterMerchantCommandServiceServer(gs, h)
		pbdocument.RegisterMerchantDocumentQueryServiceServer(gs, h)
		pbdocument.RegisterMerchantDocumentCommandServiceServer(gs, h)
	}

	return srv, nil
}
