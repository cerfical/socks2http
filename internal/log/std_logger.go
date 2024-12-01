package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

type stdLogger struct {
	out      io.Writer
	logLevel LogLevel
}

func (l *stdLogger) Fatal(format string, v ...any) {
	l.putEntry(LogFatal, format, v...)
	os.Exit(1)
}

func (l *stdLogger) Error(format string, v ...any) {
	l.putEntry(LogError, format, v...)
}

func (l *stdLogger) Info(format string, v ...any) {
	l.putEntry(LogInfo, format, v...)
}

func (l *stdLogger) putEntry(level LogLevel, format string, v ...any) {
	if level <= l.logLevel {
		today := time.Now().Format(time.DateTime)

		prefix := fmt.Sprintf("[%v]: ", today)
		entry := fmt.Sprintf(format, v...)

		// log the entry and ignore any logging errors
		io.WriteString(l.out, prefix+entry+"\n")
	}
}

func (l *stdLogger) SetLevel(logLevel LogLevel) {
	l.logLevel = logLevel
}
