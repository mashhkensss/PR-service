package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	healthhandler "github.com/mashhkensss/PR-service/internal/http/handlers/health"
	prhandler "github.com/mashhkensss/PR-service/internal/http/handlers/pullrequest"
	statshandler "github.com/mashhkensss/PR-service/internal/http/handlers/stats"
	teamhandler "github.com/mashhkensss/PR-service/internal/http/handlers/team"
	userhandler "github.com/mashhkensss/PR-service/internal/http/handlers/user"
	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
)

type RouterConfig struct {
	TeamHandler   teamhandler.Handler
	UserHandler   userhandler.Handler
	PRHandler     prhandler.Handler
	StatsHandler  statshandler.Handler
	HealthHandler healthhandler.Handler

	Auth        *mw.Authorization
	Idempotency func(http.Handler) http.Handler
	RateLimiter func(http.Handler) http.Handler
	Validator   func(http.Handler) http.Handler
	Logger      func(http.Handler) http.Handler
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	if cfg.Logger != nil {
		r.Use(cfg.Logger)
	}
	if cfg.RateLimiter != nil {
		r.Use(cfg.RateLimiter)
	}
	if cfg.Validator != nil {
		r.Use(cfg.Validator)
	}
	if cfg.Idempotency != nil {
		r.Use(cfg.Idempotency)
	}

	if cfg.HealthHandler != nil {
		r.Route("/health", func(r chi.Router) {
			r.Get("/live", cfg.HealthHandler.Liveness)
			r.Get("/ready", cfg.HealthHandler.Readiness)
		})
	}

	r.Route("/team", func(r chi.Router) {
		r.With(cfg.adminOnly()).Post("/add", cfg.TeamHandler.AddTeam)
		r.With(cfg.userOrAdmin()).Get("/get", cfg.TeamHandler.GetTeam)
	})

	r.Route("/users", func(r chi.Router) {
		r.With(cfg.adminOnly()).Post("/setIsActive", cfg.UserHandler.SetIsActive)
		r.With(cfg.userOrAdmin()).Get("/getReview", cfg.UserHandler.GetReview)
	})

	r.Route("/pullRequest", func(r chi.Router) {
		r.With(cfg.adminOnly()).Post("/create", cfg.PRHandler.CreatePullRequest)
		r.With(cfg.adminOnly()).Post("/merge", cfg.PRHandler.MergePullRequest)
		r.With(cfg.adminOnly()).Post("/reassign", cfg.PRHandler.ReassignReviewer)
	})

	r.Route("/stats", func(r chi.Router) {
		r.With(cfg.adminOnly()).Get("/assignments", cfg.StatsHandler.GetAssignmentStats)
		r.With(cfg.adminOnly()).Get("/summary", cfg.StatsHandler.GetSummary)
	})

	return r
}

func (cfg RouterConfig) adminOnly() func(http.Handler) http.Handler {
	if cfg.Auth == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	return cfg.Auth.RequireAdmin
}

func (cfg RouterConfig) userOrAdmin() func(http.Handler) http.Handler {
	if cfg.Auth == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	return cfg.Auth.RequireUserOrAdmin
}
