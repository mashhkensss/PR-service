package pullrequesthandler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainpr "github.com/mashhkensss/PR-service/internal/domain/pullrequest"
	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
)

type prServiceMock struct {
	createFn   func(ctx context.Context, pr domainpr.PullRequest) (domainpr.PullRequest, error)
	mergeFn    func(ctx context.Context, id domain.PullRequestID, ts time.Time) (domainpr.PullRequest, error)
	reassignFn func(ctx context.Context, id domain.PullRequestID, old domain.UserID) (domainpr.PullRequest, domain.UserID, error)
}

func (m prServiceMock) Create(ctx context.Context, pr domainpr.PullRequest) (domainpr.PullRequest, error) {
	if m.createFn != nil {
		return m.createFn(ctx, pr)
	}
	return pr, nil
}

func (m prServiceMock) Merge(ctx context.Context, id domain.PullRequestID, ts time.Time) (domainpr.PullRequest, error) {
	if m.mergeFn != nil {
		return m.mergeFn(ctx, id, ts)
	}
	return domainpr.New(id, "name", "author", ts)
}

func (m prServiceMock) Reassign(ctx context.Context, id domain.PullRequestID, old domain.UserID) (domainpr.PullRequest, domain.UserID, error) {
	if m.reassignFn != nil {
		return m.reassignFn(ctx, id, old)
	}
	pr, _ := domainpr.New(id, "name", "author", time.Now())
	return pr, "new", nil
}

func prTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCreatePullRequest_Success(t *testing.T) {
	pr, _ := domainpr.New("pr-1", "Feature", "author", time.Now())
	h := &handler{
		service: prServiceMock{
			createFn: func(ctx context.Context, in domainpr.PullRequest) (domainpr.PullRequest, error) {
				return pr, nil
			},
		},
		logger: prTestLogger(),
	}
	body := `{"pull_request_id":"pr-1","pull_request_name":"Feature","author_id":"author"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mw.NewValidatorMiddleware(mw.NewTagValidator())(http.HandlerFunc(h.CreatePullRequest)).ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}

func TestCreatePullRequest_ServiceError(t *testing.T) {
	h := &handler{
		service: prServiceMock{
			createFn: func(ctx context.Context, pr domainpr.PullRequest) (domainpr.PullRequest, error) {
				return domainpr.PullRequest{}, domain.ErrPullRequestExists
			},
		},
		logger: prTestLogger(),
	}
	body := `{"pull_request_id":"pr-1","pull_request_name":"Feature","author_id":"author"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mw.NewValidatorMiddleware(mw.NewTagValidator())(http.HandlerFunc(h.CreatePullRequest)).ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestCreatePullRequest_InvalidJSON(t *testing.T) {
	h := &handler{service: prServiceMock{}, logger: prTestLogger()}
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", strings.NewReader("{"))
	rr := httptest.NewRecorder()
	mw.NewValidatorMiddleware(mw.NewTagValidator())(http.HandlerFunc(h.CreatePullRequest)).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestMergePullRequest_Success(t *testing.T) {
	pr, _ := domainpr.New("pr-1", "Feature", "author", time.Now())
	h := &handler{
		service: prServiceMock{
			mergeFn: func(ctx context.Context, id domain.PullRequestID, ts time.Time) (domainpr.PullRequest, error) {
				return pr, nil
			},
		},
		logger: prTestLogger(),
	}
	body := `{"pull_request_id":"pr-1"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mw.NewValidatorMiddleware(mw.NewTagValidator())(http.HandlerFunc(h.MergePullRequest)).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestMergePullRequest_InvalidJSON(t *testing.T) {
	h := &handler{service: prServiceMock{}, logger: prTestLogger()}
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", strings.NewReader("{"))
	rr := httptest.NewRecorder()
	mw.NewValidatorMiddleware(mw.NewTagValidator())(http.HandlerFunc(h.MergePullRequest)).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestReassignReviewer_Success(t *testing.T) {
	pr, _ := domainpr.New("pr-1", "Feature", "author", time.Now())
	h := &handler{
		service: prServiceMock{
			reassignFn: func(ctx context.Context, id domain.PullRequestID, old domain.UserID) (domainpr.PullRequest, domain.UserID, error) {
				return pr, "new", nil
			},
		},
		logger: prTestLogger(),
	}
	body := `{"pull_request_id":"pr-1","old_user_id":"old"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mw.NewValidatorMiddleware(mw.NewTagValidator())(http.HandlerFunc(h.ReassignReviewer)).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestReassignReviewer_ServiceError(t *testing.T) {
	h := &handler{
		service: prServiceMock{
			reassignFn: func(ctx context.Context, id domain.PullRequestID, old domain.UserID) (domainpr.PullRequest, domain.UserID, error) {
				return domainpr.PullRequest{}, "", domain.ErrNoActiveCandidate
			},
		},
		logger: prTestLogger(),
	}
	body := `{"pull_request_id":"pr-1","old_user_id":"old"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mw.NewValidatorMiddleware(mw.NewTagValidator())(http.HandlerFunc(h.ReassignReviewer)).ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}
