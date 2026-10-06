package database

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestResolveClusterConfig_StrictForContextPrefix(t *testing.T) {
	viper.Reset()
	viper.Set("DB_PAYMENT_HOST", "pgbouncer_payment")
	viper.Set("DB_PAYMENT_PORT", "5432")
	viper.Set("DB_PAYMENT_NAME", "pg_payment")
	viper.Set("DB_USERNAME", "postgres")
	viper.Set("DB_PASSWORD", "password")

	got, err := resolveClusterConfig(PaymentCluster)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Host != "pgbouncer_payment" || got.Port != "5432" || got.DBName != "pg_payment" {
		t.Errorf("resolved %+v, want host=pgbouncer_payment port=5432 dbname=pg_payment", got)
	}
	if got.User != "postgres" || got.Password != "password" {
		t.Errorf("credentials did not fall back to base keys: %+v", got)
	}
}

func TestResolveClusterConfig_ErrorsOnMissingContextKeys(t *testing.T) {
	cases := []struct {
		name    string
		missing string
		setup   func()
	}{
		{
			name:    "missing host",
			missing: "DB_PAYMENT_HOST",
			setup: func() {
				viper.Set("DB_PAYMENT_PORT", "5432")
				viper.Set("DB_PAYMENT_NAME", "pg_payment")
			},
		},
		{
			name:    "missing port",
			missing: "DB_PAYMENT_PORT",
			setup: func() {
				viper.Set("DB_PAYMENT_HOST", "pgbouncer_payment")
				viper.Set("DB_PAYMENT_NAME", "pg_payment")
			},
		},
		{
			name:    "missing name",
			missing: "DB_PAYMENT_NAME",
			setup: func() {
				viper.Set("DB_PAYMENT_HOST", "pgbouncer_payment")
				viper.Set("DB_PAYMENT_PORT", "5432")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()
			// A base DB_* set must NOT rescue a broken context prefix.
			viper.Set("DB_HOST", "localhost")
			viper.Set("DB_PORT", "5432")
			viper.Set("DB_NAME", "wrong_db")
			tc.setup()

			_, err := resolveClusterConfig(PaymentCluster)
			if err == nil {
				t.Fatal("expected an error for an incomplete cluster prefix, got nil")
			}
			if !strings.Contains(err.Error(), tc.missing) {
				t.Errorf("error %q should name the missing key %s", err, tc.missing)
			}
		})
	}
}

func TestResolveClusterConfig_GenericPrefixKeepsBaseFallback(t *testing.T) {
	viper.Reset()
	viper.Set("DB_HOST", "localhost")
	viper.Set("DB_PORT", "5432")
	viper.Set("DB_NAME", "legacy_db")

	for _, prefix := range []string{"", "DB"} {
		got, err := resolveClusterConfig(prefix)
		if err != nil {
			t.Fatalf("prefix %q: unexpected error: %v", prefix, err)
		}
		if got.Host != "localhost" || got.Port != "5432" || got.DBName != "legacy_db" {
			t.Errorf("prefix %q: resolved %+v, want base DB_* values", prefix, got)
		}
	}
}

func TestClusterConfig_DSN(t *testing.T) {
	cfg := ClusterConfig{Host: "h", Port: "1", User: "u", Password: "p", DBName: "d"}
	want := "host=h port=1 user=u dbname=d password=p sslmode=disable"
	if got := cfg.DSN(); got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}
