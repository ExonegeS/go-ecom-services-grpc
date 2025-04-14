package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Middleware func(next http.Handler) http.Handler

func NewMiddlewareChain(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			next = x(next)
		}
		return next
	}
}

func NewTimeoutContextMW(timeoutInSec int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutInSec))
				defer cancel()

				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
			})

	}
}

func NewLoggerMW(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				next.ServeHTTP(w, r)
				duration := time.Since(start)
				var status string
				if w, ok := w.(http.ResponseWriter); ok {
					status = w.Header().Get("Status")
				}
				logger.Info("request",
					slog.String("method", r.Method),
					slog.String("url", r.URL.String()),
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("user_agent", r.UserAgent()),
					slog.String("status", status),
					slog.String("duration", duration.String()),
				)
			},
		)
	}
}

func NewCORS(CORS_URLS string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				allowedOrigins := strings.Split(CORS_URLS, ",")
				origin := r.Header.Get("Origin")

				allowOrigin := ""
				for _, o := range allowedOrigins {
					if strings.TrimSpace(o) == origin {
						allowOrigin = origin
						break
					}
				}

				if allowOrigin != "" {
					w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				// Handle preflight OPTIONS request
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusOK)
					return
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}

func RecoveryMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
