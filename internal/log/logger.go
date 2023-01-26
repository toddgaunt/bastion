package log

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type contextKey string

const key = contextKey("logger")

type Logger interface {
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Infow(msg string, keyValues ...any)
	Warnw(msg string, keyValues ...any)
	Errorw(msg string, keyValues ...any)
	Fatalw(msg string, keyValues ...any)
	With(keyValues ...any) Logger
}

type wrapper struct {
	*zap.SugaredLogger
}

func New() Logger {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	return &wrapper{sugar}
}

func (l *wrapper) With(keyValues ...any) Logger {
	return &wrapper{l.SugaredLogger.With(keyValues...)}
}

// With decorates the logger in the context with keys and values
func With(keyValues ...any) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := From(r.Context())
			logger = logger.With(keyValues)
			ctx := context.WithValue(r.Context(), key, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Middleware adds a loger to the context
func Middleware(logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), key, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// From returns the logger in the request context, or a no-op logger if
// none was provided.
func From(ctx context.Context) Logger {
	logger, ok := ctx.Value(key).(Logger)
	if !ok {
		return &wrapper{zap.NewNop().Sugar()}
	}
	return logger
}
