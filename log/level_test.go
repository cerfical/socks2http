package log_test

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/cerfical/socks2http/log"
	"github.com/stretchr/testify/assert"
)

var knownLogLevels = []string{
	"silent",
	"error",
	"info",
	"verbose",
}

func TestLevel_String(t *testing.T) {
	t.Run("returns a valid string representation", func(t *testing.T) {
		config := quick.Config{
			Values: func(v []reflect.Value, r *rand.Rand) {
				i := randR(r, int(log.Silent), int(log.Verbose)+1)
				v[0] = reflect.ValueOf(log.Level(i))
			},
		}

		err := quick.Check(func(l log.Level) bool {
			return assert.Contains(t, knownLogLevels, l.String())
		}, &config)

		assert.NoError(t, err)
	})

	t.Run("returns an empty string on unknown log level", func(t *testing.T) {
		l := log.Level(log.Verbose + 1)
		assert.Equal(t, "", l.String())
	})
}

func TestLevel_UnmarshalText(t *testing.T) {
	t.Run("preserves input text when followed by marshalling", func(t *testing.T) {
		config := quick.Config{
			Values: func(v []reflect.Value, r *rand.Rand) {
				i := randR(r, 0, len(knownLogLevels))
				v[0] = reflect.ValueOf(knownLogLevels[i])
			},
		}

		err := quick.Check(func(text string) bool {
			var l log.Level
			if !assert.NoError(t, l.UnmarshalText([]byte(text))) {
				return false
			}

			got, err := l.MarshalText()
			if !assert.NoError(t, err) {
				return false
			}

			return assert.Equal(t, []byte(text), got)
		}, &config)

		assert.NoError(t, err)
	})

	t.Run("reports an error on invalid input", func(t *testing.T) {
		var l log.Level
		assert.Error(t, l.UnmarshalText([]byte("some-invalid-log-level")))
	})
}

func TestLevel_MarshalText(t *testing.T) {
	t.Run("preserves input value when followed by unmarshalling", func(t *testing.T) {
		config := quick.Config{
			Values: func(v []reflect.Value, r *rand.Rand) {
				i := randR(r, int(log.Silent), int(log.Verbose)+1)
				v[0] = reflect.ValueOf(log.Level(i))
			},
		}

		err := quick.Check(func(l log.Level) bool {
			text, err := l.MarshalText()
			if !assert.NoError(t, err) {
				return false
			}

			var got log.Level
			if !assert.NoError(t, got.UnmarshalText(text)) {
				return false
			}

			return assert.Equal(t, l, got)
		}, &config)

		assert.NoError(t, err)
	})

	t.Run("reports an error on invalid input", func(t *testing.T) {
		l := log.Verbose + 1

		_, err := l.MarshalText()
		assert.Error(t, err)
	})
}

func randR(r *rand.Rand, min, max int) int {
	return r.Intn(max-min) + min
}
