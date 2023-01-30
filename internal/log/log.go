package log

import (
	"github.com/toddgaunt/bastion"
	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

func New() *Logger {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	return &Logger{sugar}
}

func NewNop() *Logger {
	return &Logger{zap.NewNop().Sugar()}
}

func (l *Logger) With(keyValues ...any) bastion.Logger {
	return &Logger{l.SugaredLogger.With(keyValues...)}
}
