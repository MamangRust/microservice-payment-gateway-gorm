package database

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

// RunMigrations executes database migrations using goose against the
// bounded-context database selected by prefix (see names.go), so migrations for
// e.g. card land in pg_payment rather than an arbitrary database.
//
// The connection settings come from resolveClusterConfig, so migrations target
// exactly the same database as NewGormClientWithPrefix. path is the directory
// containing the goose migration files.
func RunMigrations(log logger.LoggerInterface, prefix, path string) error {
	cluster, err := resolveClusterConfig(prefix)
	if err != nil {
		return err
	}

	// Use pgx driver for goose
	db, err := goose.OpenDBWithDriver("pgx", cluster.DSN())
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error("Failed to close database after migrations", zap.Error(err))
		}
	}()

	log.Info("Running database migrations", zap.String("path", path), zap.String("cluster", prefix), zap.String("dbname", cluster.DBName))

	if err := goose.RunContext(context.Background(), "up", db, path); err != nil {
		return fmt.Errorf("migration 'up' failed: %w", err)
	}

	log.Info("Database migrations completed successfully")
	return nil
}
