package middleware

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
)

// LoggingMiddleware wraps the gorilla/myhandlers logging functionality to add custom fields.
func LoggingMiddleware(next http.Handler) http.Handler {
	return handlers.CustomLoggingHandler(log.Writer(), next, logFormatter)
}

// logFormatter is a custom formatter to include additional fields for logging.
func logFormatter(writer io.Writer, params handlers.LogFormatterParams) {
	// Extract custom fields like ReqID from the request context
	reqID := GetReqID(params.Request.Context())

	// Calculate the duration
	duration := time.Since(params.TimeStamp)

	// Log the request details
	log.Printf(
		`HTTP request - method=%s path=%s remote_addr=%s user_agent=%q req_id=%s status=%d bytes=%d duration=%s`,
		params.Request.Method,
		params.URL.Path,
		params.Request.RemoteAddr,
		params.Request.UserAgent(),
		reqID,
		params.StatusCode,
		params.Size,
		duration,
	)
}
