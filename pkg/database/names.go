package database

// Bounded-context database names. One database per context; the same
// identifiers are used by POSTGRES_DB in Compose/Kubernetes so a single value
// stays consistent across env, code and dashboards.
const (
	IdentityDB  = "pg_identity"
	PaymentDB   = "pg_payment"
	FinancialDB = "pg_financial"
)

// Cluster prefixes passed to NewGormClientWithPrefix and RunMigrations. Each
// prefix resolves the <PREFIX>_HOST / _PORT / _NAME env keys that select the
// bounded-context database. These must match the keys defined in .env,
// docker.env and the Kubernetes app-config ConfigMap; a prefix that no longer
// exists silently degrades to the base DB_* keys (which are not defined),
// leaving the connection with an empty host/port/dbname.
const (
	IdentityCluster  = "DB_IDENTITY"
	PaymentCluster   = "DB_PAYMENT"
	FinancialCluster = "DB_FINANCIAL"
)
