package log

import (
	"time"

	"github.com/rs/zerolog"
)

// New creates a new console [Logger].
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

// Logger provides a simple abstraction for basic logging.
type Logger struct {
	logger zerolog.Logger
}

// Fatalf logs an [fmt]-formatted [FatalLevel] message.
func (l *Logger) Fatalf(format string, v ...any) {
	l.logger.Fatal().
		Msgf(format, v...)
}

// Errorf logs an [fmt]-formatted [ErrorLevel] message.
func (l *Logger) Errorf(format string, v ...any) {
	l.logger.Error().
		Msgf(format, v...)
}

// Infof logs an [fmt]-formatted [InfoLevel] message.
func (l *Logger) Infof(format string, v ...any) {
	l.logger.Info().
		Msgf(format, v...)
}

// WithLevel creates a child [Logger] that will only log messages of the specified lvl or higher.
func (l *Logger) WithLevel(lvl Level) *Logger {
	return &Logger{l.logger.Level(zerolog.Level(lvl))}
}

// WithAttr creates a child [Logger] that will augment all messages with a key-val pair.
func (l *Logger) WithAttr(key, val string) *Logger {
	ctx := l.logger.With().Str(key, val)
	return &Logger{ctx.Logger()}
}

// WithAttrs creates a child [Logger] that will augment all messages with a set of attrs.
func (l *Logger) WithAttrs(attrs ...string) *Logger {
	if len(attrs)%2 != 0 {
		panic("expected an even number of arguments")
	}

	ctxLogger := l
	for i := 0; i < len(attrs); i += 2 {
		ctxLogger = l.WithAttr(attrs[i], attrs[i+1])
	}
	return ctxLogger
}
