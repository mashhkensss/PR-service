package middleware

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/mashhkensss/PR-service/internal/http/dto"
	"github.com/mashhkensss/PR-service/internal/http/httperror"
	"github.com/mashhkensss/PR-service/internal/idempotency"
)

const maxIdempotencyBodyBytes = 1 << 20

var errRequestBodyTooLarge = errors.New("idempotency request body too large")

func NewIdempotencyMiddleware(store idempotency.Storage, ttl time.Duration, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if store == nil || r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			body, err := readRequestBody(w, r)
			if err != nil {
				if errors.Is(err, errRequestBodyTooLarge) {
					resp := dto.NewErrorResponse(httperror.CodeInvalidInput, "request body too large")
					httperror.Write(w, http.StatusRequestEntityTooLarge, resp, log)
					return
				}
				status, resp := httperror.InvalidRequest("invalid request body")
				httperror.Write(w, status, resp, log)
				return
			}

			uri := r.URL.Path
			if raw := r.URL.RawQuery; raw != "" {
				uri = uri + "?" + raw
			}

			incomingReq := idempotency.StoredRequest{
				Method: r.Method,
				Path:   uri,
				Body:   append([]byte(nil), body...),
			}

			if storedReq, cached, ok, err := store.Get(r.Context(), key); err == nil && ok {
				if !sameRequest(storedReq, incomingReq) {
					status, resp := httperror.InvalidRequest("idempotency key replay with different request")
					httperror.Write(w, status, resp, log)
					return
				}
				writeStoredResponse(w, cached)
				return
			} else if err != nil && log != nil {
				log.Error("idempotency lookup failed", "error", err)
			}

			rec := newResponseRecorder(w)
			next.ServeHTTP(rec, r)

			resp := idempotency.StoredResponse{
				Status: rec.status,
				Body:   append([]byte(nil), rec.body.Bytes()...),
				Header: rec.header(),
			}

			if rec.status >= http.StatusBadRequest {
				return
			}

			if err := store.Save(r.Context(), key, incomingReq, ttl, resp); err != nil && log != nil {
				log.Error("idempotency save failed", "error", err)
			}
		})
	}
}

func sameRequest(a, b idempotency.StoredRequest) bool {
	if a.Method != b.Method || a.Path != b.Path {
		return false
	}
	return bytes.Equal(a.Body, b.Body)
}

func writeStoredResponse(w http.ResponseWriter, resp idempotency.StoredResponse) {
	for k, v := range resp.Header {
		w.Header().Set(k, v)
	}
	w.WriteHeader(resp.Status)
	_, _ = w.Write(resp.Body)
}

func readRequestBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	reader := http.MaxBytesReader(w, r.Body, maxIdempotencyBodyBytes)
	defer r.Body.Close()
	body, err := io.ReadAll(reader)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return nil, errRequestBodyTooLarge
		}
		return nil, err
	}
	if err := reader.Close(); err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

type responseRecorder struct {
	w      http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{w: w, body: &bytes.Buffer{}, status: http.StatusOK}
}

func (r *responseRecorder) Header() http.Header {
	return r.w.Header()
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.w.Write(b)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.w.WriteHeader(statusCode)
}

func (r *responseRecorder) header() map[string]string {
	result := make(map[string]string)
	for k, v := range r.Header() {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}
