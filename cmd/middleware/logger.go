package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Logger is a custom logging middleware using log/slog.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture the status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process the request
		next.ServeHTTP(lrw, r)

		// Log request details after processing
		duration := time.Since(start)
		slog.Info("HTTP request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("user_agent", r.UserAgent()),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("ReqID", GetReqID(r.Context())),
			slog.Int("status", lrw.statusCode),
			slog.String("status_text", http.StatusText(lrw.statusCode)),
			slog.Duration("duration", duration),
		)
	})
}

// loggingResponseWriter is a custom response writer to capture the status code.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code for logging.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
