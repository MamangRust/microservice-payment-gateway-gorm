package clickhouse

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

//go:embed schema.sql
var schemaSQL string

// resolveAddr returns the ClickHouse server address, preferring CLICKHOUSE_ADDR
// and falling back to CLICKHOUSE_HOST:CLICKHOUSE_PORT.
func resolveAddr() string {
	if addr := viper.GetString("CLICKHOUSE_ADDR"); addr != "" {
		return addr
	}

	host := viper.GetString("CLICKHOUSE_HOST")
	if host == "" {
		host = "clickhouse"
	}
	port := viper.GetString("CLICKHOUSE_PORT")
	if port == "" {
		port = "9000"
	}
	return fmt.Sprintf("%s:%s", host, port)
}

// EnsureDatabase creates CLICKHOUSE_DATABASE if it does not exist yet. It dials
// the server with the always-present "default" database because NewClient uses
// CLICKHOUSE_DATABASE as its default and would fail to ping on a fresh server.
func EnsureDatabase(l logger.LoggerInterface) error {
	database := viper.GetString("CLICKHOUSE_DATABASE")
	if database == "" {
		l.Debug("CLICKHOUSE_DATABASE is empty; relying on the server default database")
		return nil
	}
	if !isValidIdentifier(database) {
		return fmt.Errorf("invalid CLICKHOUSE_DATABASE %q: only letters, digits and underscores are allowed", database)
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{resolveAddr()},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: viper.GetString("CLICKHOUSE_USERNAME"),
			Password: viper.GetString("CLICKHOUSE_PASSWORD"),
		},
		DialTimeout: time.Second * 30,
	})
	if err != nil {
		return fmt.Errorf("failed to open clickhouse connection: %w", err)
	}
	defer conn.Close()

	if err := conn.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", database)); err != nil {
		return fmt.Errorf("failed to create clickhouse database %q: %w", database, err)
	}

	l.Debug("ClickHouse database ensured", zap.String("database", database))
	return nil
}

// ApplySchema executes every statement in schema.sql against the given
// connection. All statements are idempotent (CREATE TABLE IF NOT EXISTS), so it
// is safe to run on every startup.
func ApplySchema(ctx context.Context, conn clickhouse.Conn, l logger.LoggerInterface) error {
	for _, stmt := range splitStatements(schemaSQL) {
		if err := conn.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("failed to apply clickhouse schema: %w", err)
		}
	}

	l.Debug("ClickHouse schema applied")
	return nil
}

// splitStatements strips comment-only lines and returns the individual SQL
// statements, without their trailing semicolons.
func splitStatements(script string) []string {
	var cleaned strings.Builder
	for _, line := range strings.Split(script, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		cleaned.WriteString(line)
		cleaned.WriteString("\n")
	}

	var statements []string
	for _, part := range strings.Split(cleaned.String(), ";") {
		if stmt := strings.TrimSpace(part); stmt != "" {
			statements = append(statements, stmt)
		}
	}
	return statements
}

func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
		default:
			return false
		}
	}
	return true
}

func NewClient(l logger.LoggerInterface) (clickhouse.Conn, error) {
	addr := resolveAddr()

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: viper.GetString("CLICKHOUSE_DATABASE"),
			Username: viper.GetString("CLICKHOUSE_USERNAME"),
			Password: viper.GetString("CLICKHOUSE_PASSWORD"),
		},
		DialTimeout: time.Second * 30,
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})

	if err != nil {
		l.Error("Failed to open ClickHouse connection", zap.Error(err))
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			l.Error("ClickHouse exception during ping",
				zap.Uint32("code", uint32(exception.Code)),
				zap.String("message", exception.Message),
			)
		}
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	l.Debug("ClickHouse connection established successfully", zap.String("addr", addr))
	return conn, nil
}
