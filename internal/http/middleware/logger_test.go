package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func TestLoggerMiddleware_RecordsRequest(t *testing.T) {
	t.Parallel()

	handler := newLogCaptureHandler()
	logger := slog.New(handler)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/teams?foo=bar", nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	mw := NewLogger(logger).Middleware
	chained := chimw.RequestID(mw(next))
	chained.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}

	if len(handler.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(handler.records))
	}
	record := handler.records[0]
	if record.Message != "http_request" {
		t.Fatalf("unexpected log message: %s", record.Message)
	}
	if record.Attrs["method"] != http.MethodGet {
		t.Fatalf("missing method attribute")
	}
	if record.Attrs["path"] != "/teams" {
		t.Fatalf("unexpected path: %s", record.Attrs["path"])
	}
	if attrInt(record.Attrs, "status") != http.StatusAccepted {
		t.Fatalf("unexpected status attribute: %v", record.Attrs["status"])
	}
	if _, ok := record.Attrs["duration"]; !ok {
		t.Fatalf("missing duration attribute")
	}
	reqID, _ := record.Attrs["request_id"].(string)
	if reqID == "" {
		t.Fatalf("request_id not logged")
	}
	clientIP, _ := record.Attrs["client_ip"].(string)
	if clientIP == "" {
		t.Fatalf("client_ip not logged")
	}
}

type logCaptureHandler struct {
	records []logRecord
	attrs   []slog.Attr
}

type logRecord struct {
	Message string
	Attrs   map[string]any
}

func newLogCaptureHandler() *logCaptureHandler {
	return &logCaptureHandler{}
}

func (h *logCaptureHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *logCaptureHandler) Handle(_ context.Context, rec slog.Record) error {
	attrMap := make(map[string]any)
	rec.Attrs(func(attr slog.Attr) bool {
		attrMap[attr.Key] = attr.Value.Any()
		return true
	})
	for _, attr := range h.attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}
	h.records = append(h.records, logRecord{
		Message: rec.Message,
		Attrs:   attrMap,
	})
	return nil
}

func (h *logCaptureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logCaptureHandler{
		records: h.records,
		attrs:   append(append([]slog.Attr{}, h.attrs...), attrs...),
	}
}

func (h *logCaptureHandler) WithGroup(string) slog.Handler {
	return h
}

func attrInt(attrs map[string]any, key string) int {
	val, ok := attrs[key]
	if !ok {
		return -1
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	default:
		return -1
	}
}
