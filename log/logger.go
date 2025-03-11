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

func (l *Logger) WithLevel(level Level) *Logger {
	return &Logger{l.logger.Level(makeZerologLevel(level))}
}

func (l *Logger) WithAttr(key, val string) *Logger {
	ctx := l.logger.With().Str(key, val)
	return &Logger{ctx.Logger()}
}

func (l *Logger) WithAttrs(attrs ...string) *Logger {
	if len(attrs)%2 != 0 {
		panic("expected an even number of arguments")
	}

	ctxLogger := l
	for i := 0; i < len(attrs); i += 2 {
		ctxLogger = ctxLogger.WithAttr(attrs[i], attrs[i+1])
	}
	return ctxLogger
}
