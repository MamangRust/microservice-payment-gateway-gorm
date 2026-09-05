package main

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/server"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/auth/apps"
)

func main() {
	srv, err := apps.NewServer(&server.Config{
		ServiceName:    "auth-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		OtelEndpoint:   "otel-collector:4317",
		Port:           50051,
		DBCluster:      "DB_A",
		RedisCluster:   "REDIS_1",
		MigrationPath:  "./database/migration",
	})

	if err != nil {
		panic(err)
	}

	if err := srv.Run(); err != nil {
		panic(err)
	}
}
