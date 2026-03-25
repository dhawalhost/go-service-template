package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/dhawalhost/gokit/cache"
	"github.com/dhawalhost/gokit/database"
	"github.com/dhawalhost/gokit/health"
	"github.com/dhawalhost/gokit/logger"
	"github.com/dhawalhost/gokit/middleware"
	"github.com/dhawalhost/gokit/observability"
	"github.com/dhawalhost/gokit/ratelimit"
	"github.com/dhawalhost/gokit/server"

	svcconfig "github.com/dhawalhost/go-service-template/config"
	"github.com/dhawalhost/go-service-template/internal/handler"
	"github.com/dhawalhost/go-service-template/internal/repository"
	"github.com/dhawalhost/go-service-template/internal/service"
)

func main() {
	// Handle OS signals for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load config from env vars / yaml.
	cfg := svcconfig.MustLoad()

	// Init structured logger.
	log, err := logger.New(cfg.Log.Level, cfg.Log.Development)
	if err != nil {
		panic(err)
	}
	logger.SetGlobal(log)
	defer log.Sync() //nolint:errcheck

	// Init database (GORM + pgxpool from same DSN).
	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close() //nolint:errcheck

	// Run SQL migrations.
	if err := database.RunMigrations(ctx, cfg.Database.DSN, cfg.Database.MigrationsPath); err != nil {
		log.Fatal("failed to run migrations", zap.Error(err))
	}

	// Init Redis cache.
	redisCache, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatal("failed to connect to redis", zap.Error(err))
	}

	// Init observability (Prometheus metrics + OTel tracing).
	observability.InitMetrics(cfg.Telemetry.ServiceName)
	if cfg.Telemetry.Enabled {
		shutdownTracer, err := observability.InitTracer(ctx, cfg.Telemetry)
		if err != nil {
			log.Warn("failed to init tracer", zap.Error(err))
		} else {
			defer shutdownTracer(ctx) //nolint:errcheck
		}
	}

	// Wire dependency graph: repo -> service -> handler.
	repo := repository.NewPostgres(db.GORM)
	svc := service.New(repo, redisCache, log)
	h := handler.New(svc, log)

	// Set up health checks.
	healthHandler := health.NewHandler()
	healthHandler.Register("database", db)
	healthHandler.Register("redis", redisCache)

	// Build HTTP server.
	srv := server.New(
		server.WithAddr(cfg.Server.Addr),
		server.WithReadTimeout(cfg.Server.ReadTimeout),
		server.WithWriteTimeout(cfg.Server.WriteTimeout),
		server.WithIdleTimeout(cfg.Server.IdleTimeout),
		server.WithShutdownTimeout(cfg.Server.ShutdownTimeout),
	)

	// Register global middleware.
	srv.Use(
		middleware.RequestID(),
		middleware.TenantID(),
		middleware.SecureHeaders(),
		middleware.Logger(log),
		middleware.Recovery(log),
		middleware.CORS([]string{"*"}),
		observability.Metrics(),
	)

	// Mount unauthenticated routes.
	srv.Mount("/health/live", healthHandler.LiveHandler())
	srv.Mount("/health/ready", healthHandler.ReadyHandler())
	srv.Mount("/metrics", observability.MetricsHandler())

	// Mount authenticated + rate-limited service routes.
	// Use an in-memory store for rate limiting (swap for ratelimit.NewRedisStore for
	// distributed rate limiting in production).
	rateLimitStore := ratelimit.NewInMemoryStore()
	srv.Mount(h.Pattern(), middleware.RateLimit(middleware.RateLimitConfig{
		RequestsPerSecond: 100,
		Burst:             200,
		KeyFunc:           middleware.IPKeyFunc,
		Store:             rateLimitStore,
	})(middleware.JWT(middleware.JWTConfig{
		SecretKey:  []byte(cfg.JWT.Secret),
		Algorithm:  "HS256",
		ContextKey: "claims",
	})(h.Router())))

	log.Info("starting server", zap.String("addr", cfg.Server.Addr))
	if err := srv.Run(ctx); err != nil {
		log.Fatal("server exited with error", zap.Error(err))
	}
}
