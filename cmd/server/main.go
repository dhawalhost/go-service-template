package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/dhawalhost/gokit/cache"
	"github.com/dhawalhost/gokit/database"
	"github.com/dhawalhost/gokit/health"
	"github.com/dhawalhost/gokit/logger"
	"github.com/dhawalhost/gokit/middleware"
	"github.com/dhawalhost/gokit/observability"
	"github.com/dhawalhost/gokit/router"
	"github.com/dhawalhost/gokit/server"

	svcconfig "github.com/dhawalhost/go-service-template/config"
	"github.com/dhawalhost/go-service-template/internal/handler"
	"github.com/dhawalhost/go-service-template/internal/repository"
	"github.com/dhawalhost/go-service-template/internal/service"
)

func main() {
	// Step 1: handle OS signals for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Step 2: load config from env vars / yaml
	cfg := svcconfig.MustLoad()

	// Step 3: init structured logger
	log, err := logger.New(cfg.Log.Level, cfg.Log.Development)
	if err != nil {
		panic(err)
	}
	logger.SetGlobal(log)
	defer log.Sync() //nolint:errcheck

	// Step 4: init database (GORM + pgxpool from same DSN)
	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close() //nolint:errcheck

	// Step 5: run SQL migrations
	if err := database.RunMigrations(ctx, cfg.Database.DSN, cfg.Database.MigrationsPath); err != nil {
		log.Fatal("failed to run migrations", zap.Error(err))
	}

	// Step 6: init Redis cache
	redisCache, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatal("failed to connect to redis", zap.Error(err))
	}

	// Step 7: init observability (Prometheus metrics + OTel tracing)
	observability.InitMetrics(cfg.Telemetry.ServiceName)
	if cfg.Telemetry.Enabled {
		shutdownTracer, err := observability.InitTracer(ctx, cfg.Telemetry)
		if err != nil {
			log.Warn("failed to init tracer", zap.Error(err))
		} else {
			defer shutdownTracer(ctx) //nolint:errcheck
		}
	}

	// Step 8: wire dependency graph (repo -> service -> handler)
	repo := repository.NewPostgres(db.GORM)
	svc := service.New(repo, redisCache, log)
	h := handler.New(svc, log)

	// Step 9: set up health checks
	healthHandler := health.NewHandler()
	healthHandler.Register("database", db)
	healthHandler.Register("redis", redisCache)

	// Step 10: build chi router with global middleware
	r := router.New()
	r.Use(
		middleware.RequestID(),
		middleware.SecureHeaders(),
		middleware.Logger(log),
		middleware.Recovery(log),
		middleware.CORS([]string{"*"}),
		observability.Metrics(),
	)

	// Step 11: mount unauthenticated routes
	r.Get("/health/live", healthHandler.LiveHandler())
	r.Get("/health/ready", healthHandler.ReadyHandler())
	r.Handle("/metrics", observability.MetricsHandler())

	// Step 12: mount authenticated + rate-limited service routes
	r.Group(func(r chi.Router) {
		r.Use(
			middleware.RateLimit(middleware.RateLimitConfig{
				RequestsPerSecond: 100,
				Burst:             200,
				KeyFunc:           middleware.IPKeyFunc,
			}),
			middleware.JWT(middleware.JWTConfig{
				SecretKey:  []byte(cfg.JWT.Secret),
				Algorithm:  "HS256",
				ContextKey: "claims",
			}),
		)
		router.Mount(r, h)
	})

	// Step 13: start HTTP server
	srv := server.New(
		server.WithAddr(cfg.Server.Addr),
		server.WithReadTimeout(cfg.Server.ReadTimeout),
		server.WithWriteTimeout(cfg.Server.WriteTimeout),
		server.WithIdleTimeout(cfg.Server.IdleTimeout),
		server.WithShutdownTimeout(cfg.Server.ShutdownTimeout),
	)

	log.Info("starting server", zap.String("addr", cfg.Server.Addr))
	if err := srv.Run(ctx); err != nil {
		log.Fatal("server exited with error", zap.Error(err))
	}
}
