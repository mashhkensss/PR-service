package userhandler

import (
	"log/slog"
	"net/http"

	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
	userservice "github.com/mashhkensss/PR-service/internal/service/user"
)

type Handler interface {
	SetIsActive(w http.ResponseWriter, r *http.Request)
	GetReview(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	service userservice.Service
	logger  *slog.Logger
}

func New(service userservice.Service, logger *slog.Logger) Handler {
	return &handler{service: service, logger: logger}
}

func logFields(r *http.Request, extra ...any) []any {
	fields := []any{"method", r.Method, "path", r.URL.Path}
	if claims, ok := mw.ClaimsFromContext(r.Context()); ok && claims.Subject != "" {
		fields = append(fields, "user_id", claims.Subject)
	}
	return append(fields, extra...)
}
