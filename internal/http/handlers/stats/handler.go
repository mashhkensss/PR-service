package statshandler

import (
	"log/slog"
	"net/http"

	"github.com/mashhkensss/PR-service/internal/http/httperror"
	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
	"github.com/mashhkensss/PR-service/internal/http/response"
	statsservice "github.com/mashhkensss/PR-service/internal/service/stats"
)

type Handler interface {
	GetAssignmentStats(w http.ResponseWriter, r *http.Request)
	GetSummary(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	service statsservice.Service
	logger  *slog.Logger
}

func New(service statsservice.Service, logger *slog.Logger) Handler {
	return &handler{service: service, logger: logger}
}

func (h *handler) GetAssignmentStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetAssignments(r.Context())
	if err != nil {
		httperror.Respond(w, err, h.logger, logFields(r)...)
		return
	}
	response.JSON(w, http.StatusOK, stats)
}

func (h *handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetAssignments(r.Context())
	if err != nil {
		httperror.Respond(w, err, h.logger, logFields(r)...)
		return
	}
	totalAssignments := 0
	for _, count := range stats.ByUser {
		totalAssignments += count
	}
	summary := struct {
		UsersCount        int `json:"users_count"`
		PullRequestsCount int `json:"pull_requests_count"`
		AssignmentsTotal  int `json:"assignments_total"`
	}{
		UsersCount:        len(stats.ByUser),
		PullRequestsCount: len(stats.ByPullRequest),
		AssignmentsTotal:  totalAssignments,
	}
	response.JSON(w, http.StatusOK, summary)
}

func logFields(r *http.Request, extra ...any) []any {
	fields := []any{"method", r.Method, "path", r.URL.Path}
	if claims, ok := mw.ClaimsFromContext(r.Context()); ok && claims.Subject != "" {
		fields = append(fields, "user_id", claims.Subject)
	}
	return append(fields, extra...)
}
