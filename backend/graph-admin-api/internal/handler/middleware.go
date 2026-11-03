package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/logattr"
)

var quietPaths = map[string]struct{}{
	"/healthz": {},
	"/readyz":  {},
}

func accessLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			level := slog.LevelInfo
			if _, quiet := quietPaths[r.URL.Path]; quiet {
				level = slog.LevelDebug
			}

			if ww.Status() >= http.StatusInternalServerError {
				level = slog.LevelError
			}

			log.LogAttrs(r.Context(), level, "http request",
				slog.String("method", r.Method),
				slog.String("http.route", chi.RouteContext(r.Context()).RoutePattern()),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int64("duration_ms", time.Since(started).Milliseconds()),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}

func recoverPanics(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}

				log.ErrorContext(r.Context(), "panic recovered",
					logattr.Error(fmt.Errorf("panic: %v", recovered)),
					logattr.Stack(debug.Stack()),
				)
				w.WriteHeader(http.StatusInternalServerError)
			}()

			next.ServeHTTP(w, r)
		})
	}
}
