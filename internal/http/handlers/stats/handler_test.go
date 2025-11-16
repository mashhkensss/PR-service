package statshandler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	statsservice "github.com/mashhkensss/PR-service/internal/service/stats"
)

type statsServiceMock struct {
	getFn func(ctx context.Context) (statsservice.AssignmentsStats, error)
}

func (m statsServiceMock) GetAssignments(ctx context.Context) (statsservice.AssignmentsStats, error) {
	if m.getFn != nil {
		return m.getFn(ctx)
	}
	return statsservice.AssignmentsStats{}, nil
}

func statsTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestGetAssignmentStats_Success(t *testing.T) {
	svc := statsServiceMock{
		getFn: func(ctx context.Context) (statsservice.AssignmentsStats, error) {
			return statsservice.AssignmentsStats{ByUser: map[string]int{"u1": 1}}, nil
		},
	}
	h := &handler{service: svc, logger: statsTestLogger()}
	req := httptest.NewRequest(http.MethodGet, "/stats/assignments", nil)
	rr := httptest.NewRecorder()
	http.HandlerFunc(h.GetAssignmentStats).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetAssignmentStats_Error(t *testing.T) {
	svc := statsServiceMock{
		getFn: func(ctx context.Context) (statsservice.AssignmentsStats, error) {
			return statsservice.AssignmentsStats{}, errors.New("boom")
		},
	}
	h := &handler{service: svc, logger: statsTestLogger()}
	req := httptest.NewRequest(http.MethodGet, "/stats/assignments", nil)
	rr := httptest.NewRecorder()
	http.HandlerFunc(h.GetAssignmentStats).ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestGetSummary(t *testing.T) {
	svc := statsServiceMock{
		getFn: func(ctx context.Context) (statsservice.AssignmentsStats, error) {
			return statsservice.AssignmentsStats{
				ByUser:        map[string]int{"u1": 2, "u2": 1},
				ByPullRequest: map[string]int{"pr-1": 2},
			}, nil
		},
	}
	h := &handler{service: svc, logger: statsTestLogger()}
	req := httptest.NewRequest(http.MethodGet, "/stats/summary", nil)
	rr := httptest.NewRecorder()
	http.HandlerFunc(h.GetSummary).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
