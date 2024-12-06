package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	LogFatal LogLevel = iota
	LogError
	LogInfo
)

type LogLevel uint8

type Logger struct {
	logLevel LogLevel
}

func (l *Logger) Fatal(format string, v ...any) {
	l.putEntry(LogFatal, format, v...)
	os.Exit(1)
}

func (l *Logger) Error(format string, v ...any) {
	l.putEntry(LogError, format, v...)
}

func (l *Logger) Info(format string, v ...any) {
	l.putEntry(LogInfo, format, v...)
}

func (l *Logger) putEntry(level LogLevel, format string, v ...any) {
	if l == nil || level > l.logLevel {
		return
	}

	today := time.Now().Format(time.DateTime)
	prefix := fmt.Sprintf("[%v]: ", today)
	entry := prefix + fmt.Sprintf(format, v...) + "\n"

	_, _ = io.WriteString(os.Stderr, entry)
}
