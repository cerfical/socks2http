package log

func NewLogger(logLevel LogLevel) *Logger {
	return &Logger{
		logLevel: logLevel,
	}
}

func Fatal(format string, v ...any) {
	l := Logger{}
	l.Fatal(format, v...)
}
