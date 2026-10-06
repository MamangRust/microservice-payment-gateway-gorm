package tests

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	statsclickhouse "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/clickhouse"
)

// StatsTables lists every table created by pkg/clickhouse/schema.sql.
var StatsTables = []string{
	"transaction_events",
	"topup_events",
	"transfer_events",
	"withdraw_events",
	"saldo_events",
	"card_events",
	"merchant_events",
}

// OpenStatsConn opens a ClickHouse connection to the test container and makes
// sure the stats schema exists before returning it.
func (ts *TestSuite) OpenStatsConn() (clickhouse.Conn, error) {
	opts, err := clickhouse.ParseDSN(ts.CHURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse clickhouse dsn: %w", err)
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	if err := statsclickhouse.ApplySchema(context.Background(), conn, ts.Logger); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return conn, nil
}

// TruncateStatsTables clears every stats table so each suite starts from a
// known state.
func TruncateStatsTables(ctx context.Context, conn clickhouse.Conn) error {
	for _, table := range StatsTables {
		if err := conn.Exec(ctx, "TRUNCATE TABLE IF EXISTS "+table); err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}
	return nil
}
