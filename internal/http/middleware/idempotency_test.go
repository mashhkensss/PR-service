package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/mashhkensss/PR-service/internal/idempotency"
)

type memoryStore struct {
	mu      sync.Mutex
	entries map[string]struct {
		req  idempotency.StoredRequest
		resp idempotency.StoredResponse
	}
}

func newMemoryStore() *memoryStore {
	return &memoryStore{entries: make(map[string]struct {
		req  idempotency.StoredRequest
		resp idempotency.StoredResponse
	})}
}

func (m *memoryStore) Get(_ context.Context, key string) (idempotency.StoredRequest, idempotency.StoredResponse, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.entries[key]
	if !ok {
		return idempotency.StoredRequest{}, idempotency.StoredResponse{}, false, nil
	}
	return entry.req, entry.resp, true, nil
}

func (m *memoryStore) Save(_ context.Context, key string, req idempotency.StoredRequest, _ time.Duration, resp idempotency.StoredResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[key] = struct {
		req  idempotency.StoredRequest
		resp idempotency.StoredResponse
	}{req: req, resp: resp}
	return nil
}

func TestIdempotencyMiddlewareCachesResponse(t *testing.T) {
	store := newMemoryStore()
	wrapped := NewIdempotencyMiddleware(store, time.Minute, nil)

	calls := 0
	handler := wrapped(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/pr?team=backend", bytes.NewBufferString(`{"a":1}`))
	req.Header.Set("Idempotency-Key", "key-1")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if calls != 1 {
		t.Fatalf("expected handler to be called once, got %d", calls)
	}
	if rr.Result().StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status %d", rr.Result().StatusCode)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/pr?team=backend", bytes.NewBufferString(`{"a":1}`))
	req2.Header.Set("Idempotency-Key", "key-1")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if calls != 1 {
		t.Fatalf("handler must not be called on replay, got %d", calls)
	}
	if rr2.Result().StatusCode != http.StatusCreated {
		t.Fatalf("unexpected replay status %d", rr2.Result().StatusCode)
	}
	if body := rr2.Body.String(); body != "created" {
		t.Fatalf("unexpected body %q", body)
	}
}

func TestIdempotencyMiddlewareRejectsMismatchedRequest(t *testing.T) {
	store := newMemoryStore()
	wrapped := NewIdempotencyMiddleware(store, time.Minute, nil)
	handler := wrapped(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/resource", bytes.NewBufferString(`{"a":1}`))
	req.Header.Set("Idempotency-Key", "key-2")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", rr.Result().StatusCode)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/resource", bytes.NewBufferString(`{"a":2}`))
	req2.Header.Set("Idempotency-Key", "key-2")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for mismatched request, got %d", rr2.Result().StatusCode)
	}
}
