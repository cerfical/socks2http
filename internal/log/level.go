package log

import "github.com/rs/zerolog"

type Level int

const (
	FatalLevel = Level(zerolog.FatalLevel)
	ErrorLevel = Level(zerolog.ErrorLevel)
	InfoLevel  = Level(zerolog.InfoLevel)
)
