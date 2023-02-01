package log

import (
	"context"
	"net/http"
)

type contextKey string

const logKey = contextKey("logger")

// With is a middleware decorates the logger in the context with keys and values
func With(keyValues ...any) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := From(r.Context())
			ctx := context.WithValue(r.Context(), logKey, logger.With(keyValues))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Middleware(logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), logKey, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func From(ctx context.Context) Logger {
	logger, ok := ctx.Value(logKey).(Logger)
	if !ok {
		panic("no logger in context!")
	}
	return logger
}
