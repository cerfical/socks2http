package log

import (
	"time"

	"github.com/rs/zerolog"
)

func New() *Logger {
	out := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.DateTime
	})

	return &Logger{
		logger: zerolog.New(out).With().
			Timestamp().
			Logger(),
	}
}

type Logger struct {
	logger zerolog.Logger
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.logger.Fatal().
		Msgf(format, v...)
}

func (l *Logger) Errorf(format string, v ...any) {
	l.logger.Error().
		Msgf(format, v...)
}

func (l *Logger) Infof(format string, v ...any) {
	l.logger.Info().
		Msgf(format, v...)
}

func (l *Logger) With() *Context {
	return &Context{l.logger.With()}
}
