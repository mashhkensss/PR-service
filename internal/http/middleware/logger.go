package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type Logger struct {
	logger *slog.Logger
}

func NewLogger(l *slog.Logger) *Logger {
	return &Logger{logger: l}
}

func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := newResponseRecorder(w)
		next.ServeHTTP(rec, r)
		if l.logger != nil {
			reqID := chimw.GetReqID(r.Context())
			clientIP := remoteAddr(r)
			l.logger.Info(
				"http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"request_id", reqID,
				"client_ip", clientIP,
				"status", rec.status,
				"duration", time.Since(start),
			)
		}
	})
}

func remoteAddr(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
