package proxy

import "slices"

var DiscardLogger Logger = discardLogger{}

func NewContextLogger(l Logger, fields ...any) Logger {
	return &contextLogger{l, fields}
}

type Logger interface {
	Error(msg string, fields ...any)
	Info(msg string, fields ...any)
}

type contextLogger struct {
	log    Logger
	fields []any
}

func (l *contextLogger) Error(msg string, fields ...any) {
	l.log.Error(msg, slices.Concat(l.fields, fields)...)
}

func (l *contextLogger) Info(msg string, fields ...any) {
	l.log.Info(msg, slices.Concat(l.fields, fields)...)
}

type discardLogger struct{}

func (l discardLogger) Error(msg string, fields ...any) {}

func (l discardLogger) Info(msg string, fields ...any) {}
