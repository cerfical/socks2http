package log

import (
	"log"
	"socks2http/internal/args"
	"socks2http/internal/util"
)

func Fatal(format string, v ...any) {
	util.FatalError(format, v...)
}

func Error(format string, v ...any) {
	logEntry(args.LogError, format, v...)
}

func Info(format string, v ...any) {
	logEntry(args.LogInfo, format, v...)
}

func logEntry(level uint8, format string, v ...any) {
	if level <= args.LogLevel {
		log.Printf(format+"\n", v...)
	}
}
