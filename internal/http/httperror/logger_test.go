package httperror

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/http/dto"
)

func TestWrite_LogsAndResponds(t *testing.T) {
	t.Parallel()

	handler := newRecordingHandler()
	logger := slog.New(handler)
	rr := httptest.NewRecorder()

	payload := dto.NewErrorResponse("PR_EXISTS", "pull request exists")
	Write(rr, http.StatusConflict, payload, logger, "pull_request_id", "pr-1")

	if rr.Code != http.StatusConflict {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	var resp dto.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error != payload.Error {
		t.Fatalf("unexpected response payload: %+v", resp)
	}

	if len(handler.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(handler.records))
	}
	record := handler.records[0]
	if record.Message != "http_error" {
		t.Fatalf("unexpected log message: %s", record.Message)
	}
	if record.Level != slog.LevelError {
		t.Fatalf("expected error level, got %s", record.Level)
	}
	if attrInt(record.Attrs, "status") != http.StatusConflict {
		t.Fatalf("missing status attribute")
	}
	if record.Attrs["code"] != payload.Error.Code || record.Attrs["message"] != payload.Error.Message {
		t.Fatalf("log missing error info: %+v", record.Attrs)
	}
	if record.Attrs["pull_request_id"] != "pr-1" {
		t.Fatalf("log missing custom field: %+v", record.Attrs)
	}
}

func TestRespond_MapsDomainError(t *testing.T) {
	t.Parallel()

	handler := newRecordingHandler()
	logger := slog.New(handler)
	rr := httptest.NewRecorder()

	Respond(rr, domain.ErrTeamExists, logger)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	var resp dto.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error.Code != CodeTeamExists {
		t.Fatalf("expected code %s, got %s", CodeTeamExists, resp.Error.Code)
	}
	if len(handler.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(handler.records))
	}
}

func TestWrite_NoLogger(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	Write(rr, http.StatusBadRequest, dto.NewErrorResponse("INVALID", "bad input"), nil)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

type logRecord struct {
	Message string
	Level   slog.Level
	Attrs   map[string]any
}

type recordingHandler struct {
	records []logRecord
	attrs   []slog.Attr
}

func newRecordingHandler() *recordingHandler {
	return &recordingHandler{}
}

func (h *recordingHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *recordingHandler) Handle(_ context.Context, rec slog.Record) error {
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
		Level:   rec.Level,
		Attrs:   attrMap,
	})
	return nil
}

func (h *recordingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &recordingHandler{
		records: h.records,
		attrs:   append(append([]slog.Attr{}, h.attrs...), attrs...),
	}
}

func (h *recordingHandler) WithGroup(string) slog.Handler {
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
