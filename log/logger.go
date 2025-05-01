package log

import (
	"io"
	"os"
	"slices"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/term"
)

var Discard = New(WithLevel(LevelSilent), WithWriter(io.Discard))

func New(ops ...Option) *Logger {
	defaults := []Option{
		WithWriter(os.Stdout),
		WithLevel(LevelInfo),
	}

	l := Logger{zerolog.New(nil).
		With().Timestamp().Logger(),
	}
	for _, op := range slices.Concat(defaults, ops) {
		op(&l)
	}
	return &l
}

func WithLevel(level Level) Option {
	return func(l *Logger) {
		l.log = l.log.Level(makeZerologLevel(level))
	}
}

func WithWriter(w io.Writer) Option {
	return func(l *Logger) {
		out := w
		if isTerminal(w) {
			out = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
				w.TimeFormat = time.DateTime
				w.Out = out
			})
		}
		l.log = l.log.Output(out)
	}
}

func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		return true
	}
	return false
}

type Option func(*Logger)

type Logger struct {
	log zerolog.Logger
}

func (l *Logger) Error(msg string, err error) {
	l.logEntry(LevelError, msg, []any{"error", err})
}

func (l *Logger) Info(msg string, fields ...any) {
	l.logEntry(LevelInfo, msg, fields)
}

func (l *Logger) logEntry(level Level, msg string, fields []any) {
	entry := l.log.WithLevel(makeZerologLevel(level))
	entry.Fields(fields).
		Msg(msg)
}
