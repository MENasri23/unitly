package middleware

import (
	"net/http"
	"time"
	"unitly/internal/log"

	"github.com/go-chi/chi/v5/middleware"
)

func Logger(log log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				if recover := recover(); recover != nil {
					log.Debug("panic occured", "recover", recover)
				}
				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}

				fields := []any{
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"duration_ms", time.Since(start).Milliseconds(),
					"client_ip", middleware.GetClientIP(r.Context()),
					"bytes_written", ww.BytesWritten(),
					"scheme", scheme,
				}

				if reqID := middleware.GetReqID(r.Context()); reqID != "" {
					fields = append(fields, "request_id", reqID)
				}

				status := ww.Status()
				if status >= 500 {
					log.Error("HTTP request failed", fields...)
				} else if status >= 400 {
					log.Warn("HTTP request client error", fields...)
				} else {
					log.Info("HTTP request processed", fields...)
				}
			}()

			// Pass control to the next handler down the chain
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
