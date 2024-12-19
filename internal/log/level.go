package log

import "github.com/rs/zerolog"

// Level defines supported severity levels of log messages.
type Level zerolog.Level

const (
	// Fatal describe errors that cannot be handled gracefully and typically cause an application to exit.
	Fatal = Level(zerolog.FatalLevel)

	// Error describe errors that can either be recovered or safely ignored.
	Error = Level(zerolog.ErrorLevel)

	// Info provides informational messages that may be useful to the end user.
	Info = Level(zerolog.InfoLevel)

	// None disables logging.
	None = Level(zerolog.Disabled)
)
