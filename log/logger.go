package log

import (
	"io"
	"os"
	"slices"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/term"
)

var Discard = New(WithLevel(Silent), WithWriter(io.Discard))

func New(ops ...Option) *Logger {
	defaults := []Option{
		WithLogger(&Logger{zerolog.New(nil).
			With().Timestamp().Logger(),
		}),
		WithWriter(os.Stdout),
		WithLevel(Info),
	}

	var l Logger
	for _, op := range slices.Concat(defaults, ops) {
		op(&l)
	}
	return &l
}

func WithLogger(l *Logger) Option {
	return func(ll *Logger) {
		ll.log = l.log
	}
}

func WithFields(f Fields) Option {
	return func(l *Logger) {
		l.log = l.log.With().Fields(f).Logger()
	}
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

type Fields map[string]any

type Logger struct {
	log zerolog.Logger
}

func (l *Logger) Fatal(msg string, err error) {
	l.logEntry(Fatal, msg, nil, err)
	os.Exit(1)
}

func (l *Logger) Error(msg string, err error) {
	l.logEntry(Error, msg, nil, err)
}

func (l *Logger) Info(msg string, f Fields) {
	l.logEntry(Info, msg, f, nil)
}

func (l *Logger) logEntry(level Level, msg string, f Fields, err error) {
	entry := l.log.WithLevel(makeZerologLevel(level))
	if err != nil {
		entry = entry.Err(err)
	}

	entry.Fields(map[string]any(f)).
		Msg(msg)
}
