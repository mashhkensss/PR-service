package userhandler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainpr "github.com/mashhkensss/PR-service/internal/domain/pullrequest"
	domainuser "github.com/mashhkensss/PR-service/internal/domain/user"
	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
)

type userServiceMock struct {
	setFn  func(ctx context.Context, id domain.UserID, active bool) (domainuser.User, error)
	listFn func(ctx context.Context, id domain.UserID) ([]domainpr.PullRequest, error)
}

func (m userServiceMock) SetIsActive(ctx context.Context, id domain.UserID, active bool) (domainuser.User, error) {
	if m.setFn != nil {
		return m.setFn(ctx, id, active)
	}
	return domainuser.New(id, "user", "team", active)
}

func (m userServiceMock) GetReviewAssignments(ctx context.Context, id domain.UserID) ([]domainpr.PullRequest, error) {
	if m.listFn != nil {
		return m.listFn(ctx, id)
	}
	return nil, nil
}

func userTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestSetIsActive_Success(t *testing.T) {
	svc := userServiceMock{
		setFn: func(ctx context.Context, id domain.UserID, active bool) (domainuser.User, error) {
			return domainuser.New(id, "Bob", "backend", active)
		},
	}
	h := &handler{service: svc, logger: userTestLogger()}
	body := `{"user_id":"u1","is_active":true}`
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mw.NewValidatorMiddleware(mw.NewTagValidator())(http.HandlerFunc(h.SetIsActive)).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestSetIsActive_InvalidPayload(t *testing.T) {
	h := &handler{service: userServiceMock{}, logger: userTestLogger()}
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", strings.NewReader(`{"user_id":""}`))
	rr := httptest.NewRecorder()
	mw.NewValidatorMiddleware(mw.NewTagValidator())(http.HandlerFunc(h.SetIsActive)).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestSetIsActive_Unauthorized(t *testing.T) {
	h := &handler{service: userServiceMock{}, logger: userTestLogger()}
	auth := mw.NewAuthorization([]byte("admin-secret"), []byte("user-secret"))
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBufferString(`{}`))
	rr := httptest.NewRecorder()
	auth.RequireAdmin(http.HandlerFunc(h.SetIsActive)).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestGetReview_Success(t *testing.T) {
	pr, _ := domainpr.New("pr-1", "Feature", "author", time.Now())
	svc := userServiceMock{
		listFn: func(ctx context.Context, id domain.UserID) ([]domainpr.PullRequest, error) {
			return []domainpr.PullRequest{pr}, nil
		},
	}
	h := &handler{service: svc, logger: userTestLogger()}

	auth := mw.NewAuthorization([]byte("admin-secret"), []byte("user-secret"))
	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u1", nil)
	req.Header.Set("Authorization", "Bearer "+newToken([]byte("user-secret"), "u1", "user"))
	rr := httptest.NewRecorder()
	auth.RequireUserOrAdmin(http.HandlerFunc(h.GetReview)).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.UserID != "u1" {
		t.Fatalf("unexpected response %+v", resp)
	}
}

func TestGetReview_Forbidden(t *testing.T) {
	h := &handler{
		service: userServiceMock{
			listFn: func(ctx context.Context, id domain.UserID) ([]domainpr.PullRequest, error) {
				return nil, nil
			},
		},
		logger: userTestLogger(),
	}
	auth := mw.NewAuthorization([]byte("admin-secret"), []byte("user-secret"))
	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u1", nil)
	req.Header.Set("Authorization", "Bearer "+newToken([]byte("user-secret"), "other", "user"))
	rr := httptest.NewRecorder()
	auth.RequireUserOrAdmin(http.HandlerFunc(h.GetReview)).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func newToken(secret []byte, subject, role string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"sub":"%s","role":"%s","exp":%d}`, subject, role, time.Now().Add(time.Hour).Unix())))
	data := header + "." + payload
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return data + "." + signature
}
