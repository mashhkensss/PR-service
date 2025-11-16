package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mashhkensss/PR-service/internal/config"
	apphttp "github.com/mashhkensss/PR-service/internal/http"
	healthhandler "github.com/mashhkensss/PR-service/internal/http/handlers/health"
	prhandler "github.com/mashhkensss/PR-service/internal/http/handlers/pullrequest"
	statshandler "github.com/mashhkensss/PR-service/internal/http/handlers/stats"
	teamhandler "github.com/mashhkensss/PR-service/internal/http/handlers/team"
	userhandler "github.com/mashhkensss/PR-service/internal/http/handlers/user"
	"github.com/mashhkensss/PR-service/internal/http/middleware"
	"github.com/mashhkensss/PR-service/internal/persistence/postgres"
	idempotencyrepo "github.com/mashhkensss/PR-service/internal/persistence/postgres/idempotency"
	prrepo "github.com/mashhkensss/PR-service/internal/persistence/postgres/pullrequest"
	statsrepo "github.com/mashhkensss/PR-service/internal/persistence/postgres/stats"
	teamrepo "github.com/mashhkensss/PR-service/internal/persistence/postgres/team"
	userrepo "github.com/mashhkensss/PR-service/internal/persistence/postgres/user"
	"github.com/mashhkensss/PR-service/internal/service/assignment"
	pullrequestservice "github.com/mashhkensss/PR-service/internal/service/pullrequest"
	statsservice "github.com/mashhkensss/PR-service/internal/service/stats"
	teamservice "github.com/mashhkensss/PR-service/internal/service/team"
	userservice "github.com/mashhkensss/PR-service/internal/service/user"
)

func Build(ctx context.Context, cfg config.Config, logger *slog.Logger) (http.Handler, func() error, error) {
	db, err := sql.Open("pgx", cfg.Database.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	pingCtx := ctx
	if pingCtx == nil {
		pingCtx = context.Background()
	}
	pingErr := attemptPing(pingCtx, db, cfg.Database.ConnectRetries, cfg.Database.ConnectRetryInterval)
	if pingErr != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("ping db: %w", pingErr)
	}

	txManager := postgres.NewTxManager(db)

	teamRepo := teamrepo.New(db)
	userRepo := userrepo.New(db)
	prRepo := prrepo.New(db)
	statsRepo := statsrepo.New(db)

	teamSvc := teamservice.New(teamRepo, txManager)
	userSvc := userservice.New(userRepo, prRepo)
	assigner := assignment.NewStrategy(nil)
	prSvc := pullrequestservice.New(teamRepo, userRepo, prRepo, txManager, assigner)
	statsSvc := statsservice.New(statsRepo)

	teamHandler := teamhandler.New(teamSvc, logger.With("handler", "team"))
	userHandler := userhandler.New(userSvc, logger.With("handler", "user"))
	prHandler := prhandler.New(prSvc, logger.With("handler", "pullrequest"))
	statsHandler := statshandler.New(statsSvc, logger.With("handler", "stats"))
	healthHandler := healthhandler.New(db)

	auth := middleware.NewAuthorization([]byte(cfg.Auth.AdminSecret), []byte(cfg.Auth.UserSecret))
	l := middleware.NewLogger(logger.With("component", "http"))
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.Requests, cfg.RateLimit.Interval, cfg.RateLimit.TrustForwardHeader)
	idempotencyRepo := idempotencyrepo.New(db)
	idempotency := middleware.NewIdempotencyMiddleware(idempotencyRepo, cfg.Idempotency.TTL, logger.With("component", "idempotency"))
	validator := middleware.NewValidatorMiddleware(middleware.NewTagValidator())

	router := apphttp.NewRouter(apphttp.RouterConfig{
		TeamHandler:   teamHandler,
		UserHandler:   userHandler,
		PRHandler:     prHandler,
		StatsHandler:  statsHandler,
		HealthHandler: healthHandler,
		Auth:          auth,
		Logger:        l.Middleware,
		RateLimiter:   rateLimiter.Middleware,
		Idempotency:   idempotency,
		Validator:     validator,
	})

	cleanup := func() error {
		return db.Close()
	}

	return router, cleanup, nil
}

func attemptPing(ctx context.Context, db *sql.DB, retries int, interval time.Duration) error {
	if retries <= 0 {
		retries = 1
	}
	if interval <= 0 {
		interval = time.Second
	}
	var lastErr error
	for i := 0; i < retries; i++ {
		if err := db.PingContext(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
	return lastErr
}
