package log_test

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/cerfical/socks2http/internal/log"
	"github.com/stretchr/testify/suite"
)

func TestLogger(t *testing.T) {
	suite.Run(t, new(LoggerTest))
}

type LoggerTest struct {
	suite.Suite
}

func (t *LoggerTest) TestLog() {
	levels := []log.Level{
		log.LevelInfo,
		log.LevelError,
	}

	for _, level := range levels {
		t.Run(fmt.Sprintf("%[1]v is logged if log level is %[1]v or higher", level), func() {
			got := encodeLog(level, level)

			t.Contains(got, "log message")
			t.Regexp(regexp.MustCompile("error(.*)description"), got)
		})

		t.Run(fmt.Sprintf("%v is not logged if log level is lower", level), func() {
			got := encodeLog(level-1, level)
			t.Equal("", got)
		})
	}

	t.Run("silent is never logged", func() {
		got := encodeLog(log.LevelSilent, log.LevelInfo)
		t.Equal("", got)
	})
}

func encodeLog(logLevel, msgLevel log.Level) string {
	var buf bytes.Buffer
	l := log.New(log.WithLevel(logLevel), log.WithWriter(&buf))

	switch msgLevel {
	case log.LevelError:
		l.Error("log message", errors.New("description"))
	case log.LevelInfo:
		l.WithFields("error", "description").
			Info("log message")
	}

	return buf.String()
}
