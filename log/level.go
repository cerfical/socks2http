package log

import (
	"errors"

	"github.com/rs/zerolog"
)

const (
	Fatal = Level(zerolog.FatalLevel)
	Error = Level(zerolog.ErrorLevel)
	Info  = Level(zerolog.InfoLevel)
	None  = Level(zerolog.Disabled)
)

type Level zerolog.Level

func (l *Level) UnmarshalText(text []byte) error {
	switch text := string(text); text {
	case "fatal":
		*l = Fatal
	case "error":
		*l = Error
	case "info":
		*l = Info
	case "none":
		*l = None
	default:
		return errors.New("unknown log level")
	}
	return nil
}

func (l Level) MarshalText() ([]byte, error) {
	var text string
	switch l {
	case Fatal:
		text = "fatal"
	case Error:
		text = "error"
	case Info:
		text = "info"
	case None:
		text = "none"
	default:
		return nil, errors.New("unknown log level")
	}
	return []byte(text), nil
}
