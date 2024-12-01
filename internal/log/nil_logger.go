package log

func NilLogger() Logger {
	return nilLogger{}
}

type nilLogger struct{}

func (nilLogger) Fatal(format string, v ...any) {}

func (nilLogger) Error(format string, v ...any) {}

func (nilLogger) Info(format string, v ...any) {}

func (nilLogger) SetLevel(logLevel LogLevel) {}
