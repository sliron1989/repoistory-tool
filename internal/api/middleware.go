package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sliron1989/repoistory-tool/internal/metrics"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		duration := time.Since(start)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", sw.status)).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err, "path", r.URL.Path)
				http.Error(w, `{"error":"internal server error","code":500}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
