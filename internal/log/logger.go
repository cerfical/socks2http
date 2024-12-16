package log

import (
	"time"

	"github.com/rs/zerolog"
)

func New(lvl Level) *Logger {
	out := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.DateTime
	})

	return &Logger{
		log: zerolog.New(out).
			Level(zerolog.Level(lvl)).
			With().
			Timestamp().
			Logger(),
	}
}

type Logger struct {
	log zerolog.Logger
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.log.Fatal().
		Msgf(format, v...)
}

func (l *Logger) Errorf(format string, v ...any) {
	l.log.Error().
		Msgf(format, v...)
}

func (l *Logger) Infof(format string, v ...any) {
	l.log.Info().
		Msgf(format, v...)
}
