package internalhttp

import (
	"fmt"
	"net/http"
	"time"

	logger "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
)

func loggingMiddleware(logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			latency := time.Since(start)
			clientIP := r.RemoteAddr
			method := r.Method
			path := r.URL.RequestURI()
			proto := r.Proto
			ua := r.UserAgent()
			code := ww.statusCode

			logLine := fmt.Sprintf("%s [%s] %s %s %s %d %v \"%s\"",
				clientIP,
				start.Format("13/Jul/1996:10:00:00 -0700"),
				method, path, proto, code, latency, ua,
			)
			logger.Info(logLine)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
