package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/service/email/config"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/email/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/email/mailer"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/dotenv"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/kafka"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	otel_pkg "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/otel"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	telemetry := otel_pkg.NewTelemetry(otel_pkg.Config{
		ServiceName:            "email-service",
		ServiceVersion:         "v1.0.0",
		Environment:            "production",
		Endpoint:               "otel-collector:4317",
		Insecure:               true,
		EnableRuntimeMetrics:   true,
		RuntimeMetricsInterval: 15 * time.Second,
	})

	if err := telemetry.Init(context.Background()); err != nil {
		return
	}

	logger, err := logger.NewLogger("email-service", telemetry.GetLogger())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logger: %v\n", err)
		os.Exit(1)
	}

	if err := dotenv.Viper(); err != nil {
		logger.Fatal("Failed to load .env file", zap.Error(err))
	}

	ctx := context.Background()

	cfg := config.Config{
		KafkaBrokers: strings.Split(viper.GetString("KAFKA_BROKERS"), ","),
		SMTPServer:   viper.GetString("SMTP_SERVER"),
		SMTPPort:     viper.GetInt("SMTP_PORT"),
		SMTPUser:     viper.GetString("SMTP_USER"),
		SMTPPass:     viper.GetString("SMTP_PASS"),
	}

	defer func() {
		if err := telemetry.Shutdown(ctx); err != nil {
			logger.Fatal("Failed to shutdown tracer provider", zap.Error(err))
		}
	}()

	m := mailer.NewMailer(
		ctx,
		cfg.SMTPServer,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPass,
		logger,
	)

	h := handler.NewEmailHandler(ctx, logger, m)

	myKafka, err := kafka.NewKafka(logger, cfg.KafkaBrokers)
	if err != nil {
		logger.Fatal("Failed to create Kafka producer", zap.Error(err))
	}
	err = myKafka.StartConsumers([]string{
			"email-service-topic-auth-register",
			"email-service-topic-auth-forgot-password",
			"email-service-topic-auth-verify-code-success",
			"email-service-topic-saldo-create",
			"email-service-topic-topup-create",
			"email-service-topic-transaction-create",
			"email-service-topic-transfer-create",
			"email-service-topic-withdraw-create",
			"email-service-topic-withdraw-update",
			"email-service-topic-merchant-create",
			"email-service-topic-merchant-update-status",
			"email-service-topic-merchant-document-create",
			"email-service-topic-merchant-document-update-status",
		}, "email-service-group", h)

	if err != nil {
		logger.Fatal("Failed to start Kafka consumers", zap.Error(err))
	}

	logger.Info("Email service started", zap.String("service", "email-service"))

	select {}

}
