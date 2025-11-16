package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	rl := NewRateLimiter(2, time.Hour, false)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Result().StatusCode != http.StatusOK {
			t.Fatalf("expected request %d to pass, got %d", i, rec.Result().StatusCode)
		}
	}
}

func TestRateLimiter_BlocksAfterLimit(t *testing.T) {
	rl := NewRateLimiter(1, time.Second, false)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/limit", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Result().StatusCode != http.StatusOK {
		t.Fatalf("first call should succeed, got %d", rec1.Result().StatusCode)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/limit", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Result().StatusCode != http.StatusTooManyRequests {
		t.Fatalf("second call should be rate limited, got %d", rec2.Result().StatusCode)
	}
	if header := rec2.Header().Get("Retry-After"); header == "" {
		t.Fatalf("missing Retry-After header")
	}
}

func TestRateLimiter_TrustForwardedHeader(t *testing.T) {
	rl := NewRateLimiter(1, time.Hour, true)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	reqA := httptest.NewRequest(http.MethodGet, "/xff", nil)
	reqA.Header.Set("X-Forwarded-For", "10.0.0.1")
	reqA.RemoteAddr = "192.0.2.1:1234"
	recA := httptest.NewRecorder()
	handler.ServeHTTP(recA, reqA)
	if recA.Result().StatusCode != http.StatusOK {
		t.Fatalf("request A should pass, got %d", recA.Result().StatusCode)
	}

	reqB := httptest.NewRequest(http.MethodGet, "/xff", nil)
	reqB.Header.Set("X-Forwarded-For", "10.0.0.2")
	reqB.RemoteAddr = "192.0.2.1:1234"
	recB := httptest.NewRecorder()
	handler.ServeHTTP(recB, reqB)
	if recB.Result().StatusCode != http.StatusOK {
		t.Fatalf("request from another forwarded IP should pass, got %d", recB.Result().StatusCode)
	}
}
