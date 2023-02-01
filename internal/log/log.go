package log

import (
	"go.uber.org/zap"
)

type Level int

const (
	Debug = Level(0)
	Info  = Level(1)
	Warn  = Level(2)
	Error = Level(3)
	Fatal = Level(4)
)

type Logger interface {
	Print(n Level, args ...any)
	Printf(n Level, format string, args ...any)
	With(keyValues ...any) Logger
}

type wrapper struct {
	logger *zap.SugaredLogger
}

func New() Logger {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	return &wrapper{sugar}
}

func NewNop() Logger {
	return &wrapper{zap.NewNop().Sugar()}
}

func (l *wrapper) Print(n Level, args ...any) {
	switch n {
	case Debug:
		l.logger.Debug(args...)
	case Info:
		l.logger.Info(args...)
	case Warn:
		l.logger.Warn(args...)
	case Error:
		l.logger.Error(args...)
	case Fatal:
		l.logger.Fatal(args...)
	}
}

func (l *wrapper) Printf(n Level, format string, args ...any) {
	switch n {
	case Debug:
		l.logger.Debugf(format, args...)
	case Info:
		l.logger.Infof(format, args...)
	case Warn:
		l.logger.Warnf(format, args...)
	case Error:
		l.logger.Errorf(format, args...)
	case Fatal:
		l.logger.Fatalf(format, args...)
	}
}

func (l *wrapper) With(keyValues ...any) Logger {
	return &wrapper{l.logger.With(keyValues...)}
}
