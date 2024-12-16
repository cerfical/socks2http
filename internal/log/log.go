package log

func Fatal(format string, v ...any) {
	logger.Fatalf(format, v...)
}

func Errorf(format string, v ...any) {
	logger.Errorf(format, v...)
}

func Infof(format string, v ...any) {
	logger.Infof(format, v...)
}

var logger = New(InfoLevel)
