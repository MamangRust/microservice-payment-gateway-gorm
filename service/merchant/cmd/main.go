package main

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/server"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/apps"
)

func main() {
	srv, err := apps.NewServer(&server.Config{
		ServiceName:    "merchant-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		OtelEndpoint:   "otel-collector:4317",
		Port:           50054,
		DBCluster:      database.PaymentCluster,
		RedisCluster:   "REDIS_2",
		MigrationPath:  "./database/migration",
	})

	if err != nil {
		panic(err)
	}

	if err := srv.Run(); err != nil {
		panic(err)
	}
}
