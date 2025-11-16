package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mashhkensss/PR-service/internal/http/dto"
)

func TestJSONAndErrorResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		write      func(http.ResponseWriter)
		wantStatus int
		wantBody   any
	}{
		{
			name: "json payload",
			write: func(w http.ResponseWriter) {
				JSON(w, http.StatusCreated, map[string]string{"ok": "true"})
			},
			wantStatus: http.StatusCreated,
			wantBody:   map[string]string{"ok": "true"},
		},
		{
			name: "error helper",
			write: func(w http.ResponseWriter) {
				Error(w, http.StatusForbidden, "FORBIDDEN", "access denied")
			},
			wantStatus: http.StatusForbidden,
			wantBody:   dto.NewErrorResponse("FORBIDDEN", "access denied"),
		},
		{
			name: "error response helper",
			write: func(w http.ResponseWriter) {
				ErrorResponse(w, http.StatusBadRequest, dto.NewErrorResponse("INVALID", "bad input"))
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   dto.NewErrorResponse("INVALID", "bad input"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rr := httptest.NewRecorder()
			tt.write(rr)

			if rr.Code != tt.wantStatus {
				t.Fatalf("unexpected status: got %d want %d", rr.Code, tt.wantStatus)
			}
			if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("unexpected content-type: %s", ct)
			}

			var got any
			switch tt.wantBody.(type) {
			case map[string]string:
				var v map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &v); err != nil {
					t.Fatalf("unmarshal map: %v", err)
				}
				got = v
			case dto.ErrorResponse:
				var v dto.ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &v); err != nil {
					t.Fatalf("unmarshal error response: %v", err)
				}
				got = v
			default:
				t.Fatalf("unsupported expected type %T", tt.wantBody)
			}

			if !equalJSONBodies(got, tt.wantBody) {
				t.Fatalf("unexpected body: got %#v want %#v", got, tt.wantBody)
			}
		})
	}
}

func TestJSON_EncodeError(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	payload := struct {
		Ch chan int `json:"ch"`
	}{
		Ch: make(chan int),
	}

	JSON(rr, http.StatusOK, payload)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
	if body := rr.Body.String(); body != "Internal Server Error\n" {
		t.Fatalf("unexpected body: %q", body)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}

func equalJSONBodies(got, want any) bool {
	gotBytes, err := json.Marshal(got)
	if err != nil {
		return false
	}
	wantBytes, err := json.Marshal(want)
	if err != nil {
		return false
	}
	return string(gotBytes) == string(wantBytes)
}
