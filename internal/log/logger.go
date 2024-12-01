package log

import "os"

const (
	LogFatal LogLevel = iota
	LogError
	LogInfo
)

type LogLevel uint8

type Logger interface {
	Fatal(format string, v ...any)
	Error(format string, v ...any)
	Info(format string, v ...any)

	SetLevel(logLevel LogLevel)
}

func NewLogger() Logger {
	return &stdLogger{
		out: os.Stderr,
	}
}
