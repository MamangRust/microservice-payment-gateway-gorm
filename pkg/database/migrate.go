package database

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// RunMigrations executes database migrations using goose.
// path: directory containing migration files.
func RunMigrations(log logger.LoggerInterface, path string) error {
	prefix := "DB"

	host := viper.GetString(fmt.Sprintf("%s_HOST", prefix))
	if host == "" {
		host = viper.GetString("DB_HOST")
	}
	port := viper.GetString(fmt.Sprintf("%s_PORT", prefix))
	if port == "" {
		port = viper.GetString("DB_PORT")
	}
	user := viper.GetString(fmt.Sprintf("%s_USERNAME", prefix))
	if user == "" {
		user = viper.GetString("DB_USERNAME")
	}
	dbname := viper.GetString(fmt.Sprintf("%s_NAME", prefix))
	if dbname == "" {
		dbname = viper.GetString("DB_NAME")
	}
	password := viper.GetString(fmt.Sprintf("%s_PASSWORD", prefix))
	if password == "" {
		password = viper.GetString("DB_PASSWORD")
	}

	// Use pgx driver for goose
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host, port, user, dbname, password,
	)

	db, err := goose.OpenDBWithDriver("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error("Failed to close database after migrations", zap.Error(err))
		}
	}()

	log.Info("Running database migrations", zap.String("path", path), zap.String("dbname", dbname))

	if err := goose.RunContext(context.Background(), "up", db, path); err != nil {
		return fmt.Errorf("migration 'up' failed: %w", err)
	}

	log.Info("Database migrations completed successfully")
	return nil
}
