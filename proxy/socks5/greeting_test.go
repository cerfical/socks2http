package socks5_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/proxy/socks5"
	"github.com/stretchr/testify/suite"
)

func TestGreeting(t *testing.T) {
	suite.Run(t, new(GreetingTest))
}

type GreetingTest struct {
	suite.Suite
}

func (t *GreetingTest) TestRead() {
	t.Run("decodes auth methods", func() {
		g, err := decodeGreeting([]byte{5, 0x01, 0x00})
		t.Require().NoError(err)

		want := []socks5.AuthMethod{socks5.AuthNone}
		t.Equal(want, g.AuthMethods)
	})

	t.Run("rejects an invalid version", func() {
		_, err := decodeGreeting([]byte{0x06})
		t.Require().ErrorIs(err, socks5.ErrInvalidVersion)
	})
}

func (t *GreetingTest) TestWrite() {
	t.Run("encodes a SOCKS5 version", func() {
		g := socks5.Greeting{}

		got, err := encodeGreeting(&g)
		t.Require().NoError(err)

		want := byte(0x05)
		t.Equal(want, got[0])
	})

	t.Run("encodes auth methods", func() {
		g := socks5.Greeting{
			AuthMethods: []socks5.AuthMethod{
				socks5.AuthNone,
			},
		}

		got, err := encodeGreeting(&g)
		t.Require().NoError(err)

		want := []byte{1, 0x00}
		t.Equal(want, got[1:])
	})

	t.Run("rejects more than 255 auth methods", func() {
		g := socks5.Greeting{
			AuthMethods: make([]socks5.AuthMethod, 256),
		}

		_, err := encodeGreeting(&g)
		t.Require().ErrorIs(err, socks5.ErrTooManyAuthMethods)
	})
}

func decodeGreeting(b []byte) (*socks5.Greeting, error) {
	return socks5.ReadGreeting(bufio.NewReader(bytes.NewReader(b)))
}

func encodeGreeting(g *socks5.Greeting) ([]byte, error) {
	var buf bytes.Buffer
	if err := g.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
