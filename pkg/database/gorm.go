package database

import (
	"fmt"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewGormClient creates a GORM DB client using the default "DB" prefix.
func NewGormClient(log logger.LoggerInterface) (*gorm.DB, error) {
	return NewGormClientWithPrefix(log, "DB")
}

// NewGormClientWithPrefix creates a GORM DB client reading config from viper
// with the given prefix, falling back to unprefixed keys. This mirrors the
// config fallback of the previous pgx-based client so every service can switch to
// GORM without changing its config layout.
func NewGormClientWithPrefix(log logger.LoggerInterface, prefix string) (*gorm.DB, error) {
	if prefix == "" {
		prefix = "DB"
	}

	dbDriver := viper.GetString(fmt.Sprintf("%s_DRIVER", prefix))
	if dbDriver == "" {
		dbDriver = viper.GetString("DB_DRIVER")
	}

	if dbDriver != "postgres" && dbDriver != "pgx" && dbDriver != "" {
		log.Error("gorm postgres driver only supports PostgreSQL", zap.String("DB_DRIVER", dbDriver))
		return nil, fmt.Errorf("gorm postgres driver only supports PostgreSQL, got: %s", dbDriver)
	}

	hostKey := fmt.Sprintf("%s_HOST", prefix)
	portKey := fmt.Sprintf("%s_PORT", prefix)
	userKey := fmt.Sprintf("%s_USERNAME", prefix)
	nameKey := fmt.Sprintf("%s_NAME", prefix)
	passKey := fmt.Sprintf("%s_PASSWORD", prefix)

	host := viper.GetString(hostKey)
	if host == "" {
		host = viper.GetString("DB_HOST")
	}
	port := viper.GetString(portKey)
	if port == "" {
		port = viper.GetString("DB_PORT")
	}
	user := viper.GetString(userKey)
	if user == "" {
		user = viper.GetString("DB_USERNAME")
	}
	dbname := viper.GetString(nameKey)
	if dbname == "" {
		dbname = viper.GetString("DB_NAME")
	}
	password := viper.GetString(passKey)
	if password == "" {
		password = viper.GetString("DB_PASSWORD")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, port, user, dbname, password,
	)

	// Resolve connection pool settings (same fallback chain as the previous client).
	maxOpenConns := viper.GetInt(fmt.Sprintf("%s_MAX_OPEN_CONNS", prefix))
	if maxOpenConns <= 0 {
		maxOpenConns = viper.GetInt("DB_MAX_OPEN_CONNS")
		if maxOpenConns <= 0 {
			maxOpenConns = 100
		}
	}

	maxIdleConns := viper.GetInt(fmt.Sprintf("%s_MIN_IDLE_CONNS", prefix))
	if maxIdleConns <= 0 {
		maxIdleConns = viper.GetInt("DB_MIN_IDLE_CONNS")
		if maxIdleConns <= 0 {
			maxIdleConns = 50
		}
	}

	connMaxLifetime := viper.GetDuration(fmt.Sprintf("%s_CONN_MAX_LIFETIME", prefix))
	if connMaxLifetime == 0 {
		connMaxLifetime = viper.GetDuration("DB_CONN_MAX_LIFETIME")
		if connMaxLifetime == 0 {
			connMaxLifetime = time.Hour
		}
	}

	connMaxIdleTime := viper.GetDuration(fmt.Sprintf("%s_CONN_MAX_IDLE_TIME", prefix))
	if connMaxIdleTime == 0 {
		connMaxIdleTime = viper.GetDuration("DB_CONN_MAX_IDLE_TIME")
		if connMaxIdleTime == 0 {
			connMaxIdleTime = 30 * time.Minute
		}
	}

	gormLogLevel := gormlogger.Info
	if viper.GetString("APP_ENV") == "production" {
		gormLogLevel = gormlogger.Warn
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		log.Error("Failed to open GORM connection", zap.Error(err), zap.String("prefix", prefix))
		return nil, fmt.Errorf("failed to open gorm connection for %s: %w", prefix, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Error("Failed to get underlying sql.DB", zap.Error(err), zap.String("prefix", prefix))
		return nil, fmt.Errorf("failed to get underlying sql.DB for %s: %w", prefix, err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	if err := sqlDB.Ping(); err != nil {
		log.Error("Failed to ping database via GORM", zap.Error(err), zap.String("prefix", prefix))
		return nil, fmt.Errorf("failed to ping database %s via gorm: %w", prefix, err)
	}

	log.Debug("GORM database connection established successfully",
		zap.String("prefix", prefix),
		zap.Int("MaxOpenConns", maxOpenConns),
		zap.Int("MaxIdleConns", maxIdleConns),
		zap.Duration("ConnMaxLifetime", connMaxLifetime),
		zap.Duration("ConnMaxIdleTime", connMaxIdleTime),
	)

	return db, nil
}
