package database

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// ClusterConfig holds the resolved connection settings for one database cluster
// (bounded context).
type ClusterConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// DSN renders the libpq-style connection string shared by GORM and goose.
func (c ClusterConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		c.Host, c.Port, c.User, c.DBName, c.Password,
	)
}

// resolveClusterConfig resolves the connection settings for the given cluster
// prefix (see names.go), used by both NewGormClientWithPrefix and
// RunMigrations so the two can never drift apart.
//
// A bounded-context prefix (e.g. DB_PAYMENT) requires its own <PREFIX>_HOST,
// <PREFIX>_PORT and <PREFIX>_NAME: if any of them is missing the call fails
// instead of falling back, so a service can never silently connect to (or
// migrate) the wrong database. The generic "DB" prefix keeps the legacy
// fallback to the base DB_* keys. Credentials always fall back to
// DB_USERNAME / DB_PASSWORD.
func resolveClusterConfig(prefix string) (ClusterConfig, error) {
	if prefix == "" {
		prefix = "DB"
	}
	strict := prefix != "DB"

	host := viper.GetString(prefix + "_HOST")
	port := viper.GetString(prefix + "_PORT")
	dbname := viper.GetString(prefix + "_NAME")

	if strict {
		var missing []string
		if host == "" {
			missing = append(missing, prefix+"_HOST")
		}
		if port == "" {
			missing = append(missing, prefix+"_PORT")
		}
		if dbname == "" {
			missing = append(missing, prefix+"_NAME")
		}
		if len(missing) > 0 {
			return ClusterConfig{}, fmt.Errorf(
				"incomplete database configuration for cluster %s: missing %s",
				prefix, strings.Join(missing, ", "),
			)
		}
	} else {
		if host == "" {
			host = viper.GetString("DB_HOST")
		}
		if port == "" {
			port = viper.GetString("DB_PORT")
		}
		if dbname == "" {
			dbname = viper.GetString("DB_NAME")
		}
	}

	user := viper.GetString(prefix + "_USERNAME")
	if user == "" {
		user = viper.GetString("DB_USERNAME")
	}
	password := viper.GetString(prefix + "_PASSWORD")
	if password == "" {
		password = viper.GetString("DB_PASSWORD")
	}

	return ClusterConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}, nil
}
