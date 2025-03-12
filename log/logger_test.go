package log_test

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/cerfical/socks2http/log"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	levels := []log.Level{
		log.Info,
		log.Error,
	}

	for _, level := range levels {
		t.Run(fmt.Sprintf("%[1]v is logged if level is %[1]v or higher", level), func(t *testing.T) {
			got := writeLog(level, level)

			assert.Contains(t, got, "log message")
			assert.Regexp(t, regexp.MustCompile("error(.*)description"), got)
		})

		t.Run(fmt.Sprintf("%[1]v is not logged if level is lower", level), func(t *testing.T) {
			got := writeLog(level-1, level)
			assert.Equal(t, "", got)
		})
	}

	t.Run("silent is never logged", func(t *testing.T) {
		got := writeLog(log.Silent, log.Info)
		assert.Equal(t, "", got)
	})
}

func writeLog(logLevel, msgLevel log.Level) string {
	var buf bytes.Buffer
	l := log.New(log.WithLevel(logLevel), log.WithWriter(&buf))

	switch msgLevel {
	case log.Error:
		l.Error("log message", errors.New("description"))
	case log.Info:
		l.Info("log message", log.Fields{"error": "description"})
	}

	return buf.String()
}
