package log

import "github.com/rs/zerolog"

// Level defines supported severity levels of log messages.
type Level zerolog.Level

const (
	// FatalLevel describe errors that cannot be handled gracefully and typically cause an application to exit.
	FatalLevel = Level(zerolog.FatalLevel)

	// ErrorLevel describe errors that can either be recovered or safely ignored.
	ErrorLevel = Level(zerolog.ErrorLevel)

	// InfoLevel provides informational messages that may be useful to the end user.
	InfoLevel = Level(zerolog.InfoLevel)
)
