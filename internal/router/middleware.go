package router

import (
	"context"
	"net/http"

	"github.com/toddgaunt/bastion"
)

type contextKey string

const logKey = contextKey("logger")

// With is a middleware decorates the logger in the context with keys and values
func With(keyValues ...any) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logFromContext(r.Context())
			ctx := context.WithValue(r.Context(), logKey, logger.With(keyValues))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func logFromContext(ctx context.Context) bastion.Logger {
	logger, ok := ctx.Value(logKey).(bastion.Logger)
	if !ok {
		panic("no logger in context!")
	}
	return logger
}
