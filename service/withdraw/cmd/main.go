package main

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/server"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/apps"
)

func main() {
	srv, err := apps.NewServer(&server.Config{
		ServiceName:    "withdraw-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		OtelEndpoint:   "otel-collector:4317",
		Port:           50060,
		DBCluster:      database.FinancialCluster,
		RedisCluster:   "REDIS_3",
		MigrationPath:  "./database/migration",
	})

	if err != nil {
		panic(err)
	}

	if err := srv.Run(); err != nil {
		panic(err)
	}
}
