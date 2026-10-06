package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	pbuserrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user_role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/auth"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/hash"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	merchant_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	role_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/role/handler"
	role_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/role/repository"
	role_service "github.com/MamangRust/microservice-payment-gateway-grpc/service/role/service"
	user_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/handler"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	user_service "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/service"
	card_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/handler"
	card_service "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/service"
	saldo_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/handler"
	saldo_service "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/service"
	merchant_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/handler"
	merchant_service "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	goredis "github.com/redis/go-redis/v9"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type TestSuite struct {
	PGContainer    *tcpostgres.PostgresContainer
	RedisContainer *redis.RedisContainer
	CHContainer    *clickhouse.ClickHouseContainer
	DBURL          string
	RedisURL       string
	CHURL          string
	Ctx            context.Context
	RootDir        string

	UserAdapter     adapter.UserAdapter
	CardAdapter     adapter.CardAdapter
	MerchantAdapter adapter.MerchantAdapter
	SaldoAdapter    adapter.SaldoAdapter

	// Local gRPC Clients for Auth/Identity testing
	UserQueryClient   user.UserQueryServiceClient
	UserCommandClient user.UserCommandServiceClient
	RoleQueryClient   role.RoleQueryServiceClient
	RoleCommandClient role.RoleCommandServiceClient
	UserRoleClient    pbuserrole.UserRoleServiceClient

	// Aliases for convenience
	UserClient     *LocalUserClient
	RoleClient     *LocalRoleClient
	CardClient     *LocalCardClient
	SaldoClient    *LocalSaldoClient
	MerchantClient *LocalMerchantClient

	// Shared resources
	Logger        logger.LoggerInterface
	CacheStore    *cache.CacheStore
	Observability observability.TraceLoggerObservability
	Hashing       hash.HashPassword
	TokenManager  auth.TokenManager
}

func SetupTestSuite() (*TestSuite, error) {
	ctx := context.Background()

	// Setup PostgreSQL
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	dbURL, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres connection string: %w", err)
	}

	// Setup Redis
	redisContainer, err := redis.Run(ctx, "redis:7-alpine")
	if err != nil {
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	redisURL, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get redis connection string: %w", err)
	}

	// Setup ClickHouse
	chContainer, err := clickhouse.Run(ctx,
		"clickhouse/clickhouse-server:24.3-alpine",
		clickhouse.WithDatabase("testdb"),
		clickhouse.WithUsername("testuser"),
		clickhouse.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/ping").WithPort("8123/tcp").WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start clickhouse container: %w", err)
	}

	chURL, err := chContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get clickhouse connection string: %w", err)
	}

	ts := &TestSuite{
		PGContainer:    pgContainer,
		RedisContainer: redisContainer,
		CHContainer:    chContainer,
		DBURL:          dbURL,
		RedisURL:       redisURL,
		CHURL:          chURL,
		Ctx:            ctx,
	}

	// Find project root
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get cwd: %w", err)
	}

	root := cwd
	for {
		if _, err := os.Stat(filepath.Join(root, "justfile")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return nil, fmt.Errorf("could not find justfile in any parent directory")
		}
		root = parent
	}
	ts.RootDir = root

	// Initialize GORM client for tests
	gormDB, err := gorm.Open(gormpostgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm connection: %w", err)
	}

	// Auth helpers + Logging, Cache and Observability are initialized first so
	// every service/handler below is built with logger/cache available.
	ts.Hashing = hash.NewHashingPassword()
	ts.TokenManager, _ = auth.NewManager("test-secret-key")

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	ts.Logger, _ = logger.NewLogger("test-integration", lp)

	redisOpts, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}
	redisClient := goredis.NewClient(redisOpts)
	cacheMetrics, _ := observability.NewCacheMetrics("test-integration")
	ts.CacheStore = cache.NewCacheStore(redisClient, ts.Logger, cacheMetrics)
	ts.Observability, _ = observability.NewObservability("test-integration", ts.Logger)

	// Role handler is built first: the user repository now requires gRPC clients
	// for its role/user-role repos, which the local role client satisfies.
	roleRepos := role_repo.NewRepositories(gormDB)
	roleService := role_service.NewService(&role_service.Deps{
		Repositories: roleRepos,
		Logger:       ts.Logger,
		Cache:        ts.CacheStore,
	})
	roleHandler := role_handler.NewHandler(roleService)
	rClient := &LocalRoleClient{Handler: roleHandler}
	ts.RoleQueryClient = rClient
	ts.RoleCommandClient = rClient
	ts.RoleClient = rClient

	// User
	userRepos := user_repo.NewRepositories(&user_repo.Deps{
		Db:              gormDB,
		RoleQueryClient: rClient,
		UserRoleClient:  rClient,
		Guard:           user_repo.GuardOptions{},
	})
	userService := user_service.NewService(&user_service.Deps{
		Repositories: userRepos,
		Hash:         ts.Hashing,
		Logger:       ts.Logger,
		Cache:        ts.CacheStore,
	})
	userHandler := user_handler.NewHandler(userService)
	uClient := &LocalUserClient{Handler: userHandler}
	ts.UserQueryClient = uClient
	ts.UserCommandClient = uClient
	ts.UserClient = uClient

	// Real cross-service adapters, each backed by a local gRPC client.
	ts.UserAdapter = adapter.NewUserAdapter(uClient)

	// Card (needs the user adapter)
	cardRepos := card_repo.NewRepositories(gormDB, uClient)
	cardService := card_service.NewService(&card_service.Deps{
		Cache:           ts.CacheStore,
		Repositories:    cardRepos,
		UserAdapter:     ts.UserAdapter,
		Logger:          ts.Logger,
		Kafka:           nil,
		BillingCycleDay: 1,
	})
	cardHandler := card_handler.NewHandler(cardService)
	cardClient := &LocalCardClient{Handler: cardHandler}
	ts.CardAdapter = adapter.NewCardAdapter(cardClient, cardClient)

	// Saldo (needs the card adapter)
	saldoRepos := saldo_repo.NewRepositories(gormDB, cardClient, cardClient)
	saldoService := saldo_service.NewService(&saldo_service.Deps{
		Repositories: saldoRepos,
		CardAdapter:  ts.CardAdapter,
		Logger:       ts.Logger,
		Cache:        ts.CacheStore,
		Kafka:        nil,
	})
	saldoHandler := saldo_handler.NewHandler(saldoService)
	saldoClient := &LocalSaldoClient{Handler: saldoHandler}
	ts.SaldoAdapter = adapter.NewSaldoAdapter(saldoClient, saldoClient)

	// Merchant (needs the user adapter)
	merchantRepos := merchant_repo.NewRepositories(gormDB, uClient)
	merchantService := merchant_service.NewService(&merchant_service.Deps{
		Kafka:        nil,
		Repositories: merchantRepos,
		UserAdapter:  ts.UserAdapter,
		Logger:       ts.Logger,
		Cache:        ts.CacheStore,
	})
	merchantHandler := merchant_handler.NewHandler(merchantService)
	merchantClient := &LocalMerchantClient{Handler: merchantHandler}
	ts.MerchantAdapter = adapter.NewMerchantAdapter(merchantClient)

	// Expose local gRPC clients so per-package tests can build repositories
	// that now require cross-service gRPC clients.
	ts.UserRoleClient = rClient
	ts.CardClient = cardClient
	ts.SaldoClient = saldoClient
	ts.MerchantClient = merchantClient

	return ts, nil
}

// GormDB returns a GORM DB connection for the test database.
func (ts *TestSuite) GormDB() (*gorm.DB, error) {
	return gorm.Open(gormpostgres.Open(ts.DBURL), &gorm.Config{})
}

func (ts *TestSuite) RunMigrations(serviceNames ...string) error {
	var relPaths []string
	for _, name := range serviceNames {
		relPaths = append(relPaths, filepath.Join("service", name, "database", "migration"))
	}
	return ts.RunAllMigrations(ts.RootDir, relPaths)
}

func (ts *TestSuite) RunServiceMigrations(serviceName string) error {
	return ts.RunMigrations(serviceName)
}

func (ts *TestSuite) RunAllMigrations(root string, relPaths []string) error {
	db, err := goose.OpenDBWithDriver("pgx", ts.DBURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	// Collect all SQL migration files from all service directories
	var allFiles []string
	seen := make(map[string]bool)
	for _, relPath := range relPaths {
		absPath := filepath.Join(root, relPath)
		if seen[absPath] {
			continue
		}
		seen[absPath] = true

		entries, err := os.ReadDir(absPath)
		if err != nil {
			// Skip missing directories
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
				allFiles = append(allFiles, filepath.Join(absPath, entry.Name()))
			}
		}
	}

	if len(allFiles) == 0 {
		return nil
	}

	// Sort all migration files by filename (timestamp prefix ensures correct order)
	sort.Strings(allFiles)

	// Create a temp directory and symlink/copy all sorted migrations there
	tmpDir, err := os.MkdirTemp("", "goose-migrations-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	for i, src := range allFiles {
		// Goose requires sequential version numbering, so use zero-padded index
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", src, err)
		}
		destName := fmt.Sprintf("%04d_%s", i+1, filepath.Base(src))
		if err := os.WriteFile(filepath.Join(tmpDir, destName), data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", destName, err)
		}
	}

	if err := goose.RunContext(ts.Ctx, "up", db, tmpDir); err != nil {
		return fmt.Errorf("goose migration failed: %w", err)
	}

	return nil
}

func (ts *TestSuite) Teardown() {
	if ts.PGContainer != nil {
		if err := ts.PGContainer.Terminate(ts.Ctx); err != nil {
			log.Printf("failed to terminate postgres container: %v", err)
		}
	}
	if ts.RedisContainer != nil {
		if err := ts.RedisContainer.Terminate(ts.Ctx); err != nil {
			log.Printf("failed to terminate redis container: %v", err)
		}
	}
	if ts.CHContainer != nil {
		if err := ts.CHContainer.Terminate(ts.Ctx); err != nil {
			log.Printf("failed to terminate clickhouse container: %v", err)
		}
	}
}

// SeedRole seeds a role into the database using GORM.
func (ts *TestSuite) SeedRole(db *gorm.DB, name string) (*models.Role, error) {
	role := &models.Role{RoleName: name}
	err := db.Create(role).Error
	return role, err
}

// SeedUser seeds a user into the database using GORM.
func (ts *TestSuite) SeedUser(db *gorm.DB, firstname, lastname, email, password string) (*models.User, error) {
	user := &models.User{
		Firstname: firstname,
		Lastname:  lastname,
		Email:     email,
		Password:  password,
	}
	err := db.Create(user).Error
	return user, err
}

// SeedUserRole seeds a user-role association into the database using GORM.
func (ts *TestSuite) SeedUserRole(db *gorm.DB, userID, roleID int) error {
	userRole := &models.UserRole{
		UserID: int32(userID),
		RoleID: int32(roleID),
	}
	return db.Create(userRole).Error
}
