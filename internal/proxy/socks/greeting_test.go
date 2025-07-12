package socks_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/socks"
	"github.com/stretchr/testify/suite"
)

func TestGreeting(t *testing.T) {
	suite.Run(t, new(GreetingTest))
}

type GreetingTest struct {
	suite.Suite
}

func (t *GreetingTest) TestRead() {
	t.Run("decodes SOCKS5 greeting", func() {
		got, err := decodeGreeting([]byte{5, 1, 0})
		t.Require().NoError(err)

		t.Run("decodes version", func() {
			t.Equal(socks.V5, got.Version)
		})

		t.Run("decodes auth methods", func() {
			t.Equal([]socks.Auth{socks.AuthNone}, got.Auth)
		})
	})

	t.Run("rejects invalid version", func() {
		_, err := decodeGreeting([]byte{6})
		t.Require().ErrorIs(err, socks.ErrInvalidVersion)
	})
}

func (t *GreetingTest) TestWrite() {
	t.Run("encodes SOCKS5 greeting", func() {
		g := socks.Greeting{
			Version: socks.V5,
			Auth: []socks.Auth{
				socks.AuthNone,
			},
		}

		got, err := encodeGreeting(&g)
		t.Require().NoError(err)

		t.Run("encodes version", func() {
			t.Equal(byte(0x05), got[0])
		})

		t.Run("encodes auth methods", func() {
			t.Equal([]byte{1, 0}, got[1:])
		})

		t.Run("rejects more than 255 auth methods", func() {
			g := socks.Greeting{
				Version: socks.V5,
				Auth:    make([]socks.Auth, 256),
			}

			_, err := encodeGreeting(&g)
			t.Require().Error(err)
		})
	})

	t.Run("rejects invalid version", func() {
		g := socks.Greeting{Version: 6}

		_, err := encodeGreeting(&g)
		t.Require().ErrorIs(err, socks.ErrInvalidVersion)
	})
}

func decodeGreeting(b []byte) (*socks.Greeting, error) {
	return socks.ReadGreeting(
		bufio.NewReader(
			bytes.NewReader(b),
		),
	)
}

func encodeGreeting(g *socks.Greeting) ([]byte, error) {
	var buf bytes.Buffer
	if err := g.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
