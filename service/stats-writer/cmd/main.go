package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/clickhouse"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/dotenv"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-writer/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-writer/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-writer/usecase"
	"github.com/spf13/viper"
	"strings"
	"go.uber.org/zap"
)

func main() {
	if err := dotenv.Viper(); err != nil {
		zap.L().Error("Failed to load configuration", zap.Error(err))
	}
	log, _ := logger.NewLogger("stats-writer", nil)

	chConn, err := clickhouse.NewClient(log)
	if err != nil {
		log.Fatal("Failed to connect to ClickHouse", zap.Error(err))
	}

	brokers := strings.Split(viper.GetString("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"kafka-1:9092", "kafka-2:9092", "kafka-3:9092"}
	}
	k, err := kafka.NewKafka(log, brokers)
	if err != nil {
		log.Fatal("Failed to create Kafka producer", zap.Error(err))
	}

	// Dependency Injection
	repo := repository.NewClickhouseRepository(chConn, log)
	uc := usecase.NewStatsUseCase(repo)
	statsHandler := handler.NewStatsHandler(uc, log)

	// Start Consumer
	topics := []string{
		"payment.transaction.created",
		"stats-topic-transaction-events",
		"stats-topic-topup-events",
		"stats-topic-transfer-events",
		"stats-topic-withdraw-events",
		"stats-topic-saldo-events",
		"stats-topic-merchant-events",
		"stats-topic-card-events",
	}
	if err := k.StartConsumers(topics, "stats-writer-group", statsHandler); err != nil {
		log.Fatal("Failed to start Kafka consumers", zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Stats Writer...")
	if err := uc.Close(); err != nil {
		log.Error("Failed to close stats usecase", zap.Error(err))
	}
}
