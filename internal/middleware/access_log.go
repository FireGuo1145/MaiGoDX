package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// AccessLogMiddleware writes one concise terminal line for every HTTP request
// only when DEV_MODE is enabled. Production requests remain silent by default.
func AccessLogMiddleware(next http.Handler) http.Handler {
	if !developmentModeEnabled() {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &accessLogResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)
		log.Printf(
			"[MaiGoDX] HTTP method=%s path=%q query=%q status=%d bytes=%d duration=%s remote=%q host=%q user_agent=%q",
			r.Method,
			r.URL.Path,
			r.URL.RawQuery,
			recorder.statusCode,
			recorder.bytesWritten,
			time.Since(startedAt).Round(time.Microsecond),
			r.RemoteAddr,
			r.Host,
			r.UserAgent(),
		)
	})
}

type accessLogResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (w *accessLogResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *accessLogResponseWriter) Write(body []byte) (int, error) {
	written, err := w.ResponseWriter.Write(body)
	w.bytesWritten += int64(written)
	return written, err
}

func developmentModeEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("DEV_MODE")))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
