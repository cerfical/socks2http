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
}

func NewLogger(logLevel LogLevel) Logger {
	return &stdLogger{
		out:      os.Stderr,
		logLevel: logLevel,
	}
}

func Fatal(format string, v ...any) {
	std.Fatal(format, v...)
}

func Error(format string, v ...any) {
	std.Error(format, v...)
}

func Info(format string, v ...any) {
	std.Info(format, v...)
}

var std stdLogger = stdLogger{
	out:      os.Stderr,
	logLevel: LogFatal,
}
