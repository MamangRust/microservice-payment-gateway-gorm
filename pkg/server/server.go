package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/dotenv"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/middleware"
	otel_pkg "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/otel"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	"github.com/grafana/pyroscope-go"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
)

type GRPCServer struct {
	Logger logger.LoggerInterface
	// GormDB is the GORM database client. Services use this for repository
	// implementations after migrating from sqlc to GORM.
	GormDB           *gorm.DB
	Ctx              context.Context
	Cancel           context.CancelFunc
	CacheStore       *cache.CacheStore
	Redis            redis.UniversalClient
	Telemetry        *otel_pkg.Telemetry
	Config           *Config
	RegisterServices func(*grpc.Server)
}

func New(cfg *Config) (*GRPCServer, error) {
	if err := dotenv.Viper(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	if err := initPyroscope(cfg); err != nil {
		log.Printf("Warning: Failed to initialize pyroscope: %v", err)
	}

	telemetry := initTelemetry(cfg)
	if err := telemetry.Init(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	cacheMetrics, err := observability.NewCacheMetrics("cache")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache metrics: %w", err)
	}

	l, err := logger.NewLogger(cfg.ServiceName, telemetry.GetLogger())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	gormDB, err := database.NewGormClientWithPrefix(l, cfg.DBCluster)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database via GORM cluster %s: %w", cfg.DBCluster, err)
	}

	if cfg.MigrationPath != "" {
		if err := database.RunMigrations(l, cfg.MigrationPath); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	redisClient, err := initRedisServer(ctx, l, cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	cacheStore := cache.NewCacheStore(redisClient, l, cacheMetrics)

	return &GRPCServer{
		Logger:     l,
		GormDB:     gormDB,
		Ctx:        ctx,
		Cancel:     cancel,
		CacheStore: cacheStore,
		Redis:      redisClient,
		Telemetry:  telemetry,
		Config:     cfg,
	}, nil
}

func (s *GRPCServer) Run() error {
	defer s.Cleanup()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.Config.Port, err)
	}

	loadMonitor := resilience.NewLoadMonitor()
	circuitBreaker := resilience.NewCircuitBreaker(100, 10, s.Logger)
	requestLimiter := resilience.NewRequestLimiter(800, s.Logger)
	resilienceHandler := middleware.NewResilienceInterceptor(loadMonitor, circuitBreaker, requestLimiter)

	grpcServer := grpc.NewServer(
		grpc.MaxConcurrentStreams(DefaultMaxConcurrentConn),
		grpc.InitialConnWindowSize(DefaultWindowSize),
		grpc.InitialWindowSize(DefaultWindowSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    DefaultKeepaliveTime,
			Timeout: DefaultKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             DefaultMinKeepaliveTime,
			PermitWithoutStream: true,
		}),
		grpc.ChainUnaryInterceptor(
			middleware.ContextMiddleware(30*time.Second, s.Logger),
			middleware.RecoveryMiddleware(s.Logger),
			middleware.PyroscopeUnaryInterceptor(),
			resilienceHandler.UnaryInterceptor(),
		),
	)

	if s.RegisterServices != nil {
		s.RegisterServices(grpcServer)
	}

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	if os.Getenv("ENABLE_REFLECTION") == "true" {
		reflection.Register(grpcServer)
		s.Logger.Info("gRPC reflection enabled")
	}

	monitoringDone := s.spawnMonitoringTask()
	cleanupDone := s.spawnCleanupTask()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	errChan := make(chan error, 1)
	go func() {
		s.Logger.Info("gRPC server starting",
			zap.Int("port", s.Config.Port),
			zap.String("address", lis.Addr().String()),
		)
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	select {
	case sig := <-sigChan:
		s.Logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	case err := <-errChan:
		s.Logger.Error("Server error", zap.Error(err))
		return err
	}

	return s.gracefulShutdown(grpcServer, healthServer, monitoringDone, cleanupDone)
}

func (s *GRPCServer) gracefulShutdown(
	grpcServer *grpc.Server,
	healthServer *health.Server,
	monitoringDone, cleanupDone <-chan struct{},
) error {
	s.Logger.Info("Starting graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer shutdownCancel()

	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	s.Cancel()

	tasksDone := make(chan struct{})
	go func() {
		<-monitoringDone
		<-cleanupDone
		close(tasksDone)
	}()

	select {
	case <-tasksDone:
		s.Logger.Info("Background tasks stopped successfully")
	case <-shutdownCtx.Done():
		s.Logger.Warn("Background tasks shutdown timeout, forcing stop")
	}

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		s.Logger.Info("gRPC server stopped gracefully")
	case <-shutdownCtx.Done():
		s.Logger.Warn("Graceful shutdown timeout, forcing stop")
		grpcServer.Stop()
	}

	s.Logger.Info("Graceful shutdown completed")
	return nil
}

func (s *GRPCServer) Cleanup() {
	s.Logger.Info("Cleaning up resources...")

	if s.GormDB != nil {
		if sqlDB, err := s.GormDB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				s.Logger.Error("Failed to close GORM database connection", zap.Error(err))
			} else {
				s.Logger.Info("GORM database connection closed")
			}
		}
	}

	if s.Redis != nil {
		if err := s.Redis.Close(); err != nil {
			s.Logger.Error("Failed to close Redis connection", zap.Error(err))
		} else {
			s.Logger.Info("Redis connection closed")
		}
	}

	if s.Telemetry != nil {
		if err := s.Telemetry.Shutdown(context.Background()); err != nil {
			s.Logger.Error("Failed to shutdown telemetry", zap.Error(err))
		} else {
			s.Logger.Info("Telemetry shutdown successfully")
		}
	}

	s.Logger.Info("Cleanup completed")
}

func initPyroscope(cfg *Config) error {
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: cfg.ServiceName,
		ServerAddress:   os.Getenv("PYROSCOPE_SERVER"),
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
		},
		Tags: map[string]string{
			"service": cfg.ServiceName,
			"env":     cfg.Environment,
			"version": cfg.ServiceVersion,
		},
	})
	return err
}

func initTelemetry(cfg *Config) *otel_pkg.Telemetry {
	return otel_pkg.NewTelemetry(otel_pkg.Config{
		ServiceName:            cfg.ServiceName,
		ServiceVersion:         cfg.ServiceVersion,
		Environment:            cfg.Environment,
		Endpoint:               cfg.OtelEndpoint,
		Insecure:               true,
		EnableRuntimeMetrics:   true,
		RuntimeMetricsInterval: 15 * time.Second,
	})
}

func initRedisServer(ctx context.Context, logger logger.LoggerInterface, cfg *Config) (redis.UniversalClient, error) {
	prefix := cfg.RedisCluster
	if prefix == "" {
		prefix = "REDIS"
	}

	hostKey := fmt.Sprintf("%s_HOST", prefix)
	portKey := fmt.Sprintf("%s_PORT", prefix)
	addrsKey := fmt.Sprintf("%s_ADDRS", prefix)
	passKey := fmt.Sprintf("%s_PASSWORD", prefix)
	dbKey := fmt.Sprintf("%s_DB", prefix)

	var addrs []string
	if val := viper.GetString(addrsKey); val != "" {
		addrs = strings.Split(val, ",")
	} else if val := viper.GetString("REDIS_ADDRS"); val != "" {
		addrs = strings.Split(val, ",")
	} else {
		host := viper.GetString(hostKey)
		if host == "" {
			host = viper.GetString("REDIS_HOST")
		}
		port := viper.GetString(portKey)
		if port == "" {
			port = viper.GetString("REDIS_PORT")
		}
		addrs = []string{fmt.Sprintf("%s:%s", host, port)}
	}

	password := viper.GetString(passKey)
	if password == "" {
		password = viper.GetString("REDIS_PASSWORD")
	}
	db := viper.GetInt(dbKey)
	if !viper.IsSet(dbKey) {
		db = viper.GetInt("REDIS_DB")
	}

	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        addrs,
		Password:     password,
		DB:           db,
		DialTimeout:  RedisDialTimeout,
		ReadTimeout:  RedisReadTimeout,
		WriteTimeout: RedisWriteTimeout,
		PoolSize:     RedisPoolSize,
		MinIdleConns: RedisMinIdleConns,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}

func (s *GRPCServer) spawnMonitoringTask() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(MonitoringInterval)
		defer ticker.Stop()
		for {
			select {
			case <-s.Ctx.Done():
				return
			case <-ticker.C:
				s.monitorCache()
			}
		}
	}()
	return done
}

func (s *GRPCServer) monitorCache() {
	refCount := s.CacheStore.GetRefCount()
	stats, err := s.CacheStore.GetStats(s.Ctx)
	if err != nil {
		s.Logger.Error("Failed to get cache stats", zap.Error(err))
		return
	}
	logLevel := zap.InfoLevel
	if refCount > CacheRefCountThreshold {
		logLevel = zap.WarnLevel
	}
	if ce := s.Logger.Check(logLevel, "Cache statistics"); ce != nil {
		ce.Write(
			zap.Int64("ref_count", refCount),
			zap.Int64("total_keys", stats.TotalKeys),
			zap.Float64("hit_rate", stats.HitRate),
			zap.String("memory_used", stats.MemoryUsedHuman),
		)
	}
}

func (s *GRPCServer) spawnCleanupTask() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(CleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-s.Ctx.Done():
				return
			case <-ticker.C:
				s.CacheStore.ClearExpired(s.Ctx)
			}
		}
	}()
	return done
}
