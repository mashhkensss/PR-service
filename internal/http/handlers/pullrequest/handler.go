package pullrequesthandler

import (
	"log/slog"
	"net/http"

	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
	pullrequestservice "github.com/mashhkensss/PR-service/internal/service/pullrequest"
)

type Handler interface {
	CreatePullRequest(w http.ResponseWriter, r *http.Request)
	MergePullRequest(w http.ResponseWriter, r *http.Request)
	ReassignReviewer(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	service pullrequestservice.Service
	logger  *slog.Logger
}

func New(service pullrequestservice.Service, logger *slog.Logger) Handler {
	return &handler{service: service, logger: logger}
}

func logFields(r *http.Request, extra ...any) []any {
	fields := []any{"method", r.Method, "path", r.URL.Path}
	if claims, ok := mw.ClaimsFromContext(r.Context()); ok && claims.Subject != "" {
		fields = append(fields, "user_id", claims.Subject)
	}
	return append(fields, extra...)
}
