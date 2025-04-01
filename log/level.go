package log

import (
	"bytes"
	"errors"
	"slices"

	"github.com/rs/zerolog"
)

const (
	Silent Level = iota
	Fatal
	Error
	Info
	Verbose
)

var levels = []levelDesc{
	{"silent", zerolog.Disabled},
	{"fatal", zerolog.FatalLevel},
	{"error", zerolog.ErrorLevel},
	{"info", zerolog.InfoLevel},
	{"verbose", zerolog.InfoLevel},
}

type levelDesc struct {
	text  string
	level zerolog.Level
}

type Level int8

func (l Level) String() string {
	text, err := l.MarshalText()
	if err != nil {
		return ""
	}
	return string(text)
}

func (l Level) MarshalText() ([]byte, error) {
	if l < Silent || l > Verbose {
		return nil, errors.New("unknown log level")
	}
	return []byte(levels[l].text), nil
}

func (l *Level) UnmarshalText(text []byte) error {
	textStr := string(bytes.ToLower(text))
	i := slices.IndexFunc(levels, func(l levelDesc) bool {
		return l.text == textStr
	})
	if i == -1 {
		return errors.New("unknown log level")
	}

	*l = Level(i)
	return nil
}

func makeZerologLevel(l Level) zerolog.Level {
	if l < Silent || l > Verbose {
		panic("unknown log level")
	}
	return levels[l].level
}
