package log

import (
	"log"
	"socks2http/internal/args"
)

func Fatal(format string, v ...any) {
	log.Fatalf(format+"\n", v...)
}

func Error(format string, v ...any) {
	logEntry(args.LogError, format, v...)
}

func Info(format string, v ...any) {
	logEntry(args.LogInfo, format, v...)
}

func logEntry(level uint8, format string, v ...any) {
	log.Printf(format+"\n", v...)
}
