package socks_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/socks"
	"github.com/stretchr/testify/suite"
)

func TestGreetingReply(t *testing.T) {
	suite.Run(t, new(GreetingReplyTest))
}

type GreetingReplyTest struct {
	suite.Suite
}

func (t *GreetingReplyTest) TestRead() {
	t.Run("decodes SOCKS5 greeting reply", func() {
		got, err := decodeGreetingReply([]byte{5, 255})
		t.Require().NoError(err)

		t.Run("decodes auth method", func() {
			t.Equal(socks.AuthNotAcceptable, got.Auth)
		})

		t.Run("decodes version", func() {
			t.Equal(socks.V5, got.Version)
		})
	})

	t.Run("rejects invalid version", func() {
		_, err := decodeGreetingReply([]byte{6})
		t.Require().ErrorIs(err, socks.ErrInvalidVersion)
	})
}

func (t *GreetingReplyTest) TestWrite() {
	t.Run("encodes SOCKS5 greeting reply", func() {
		greet := socks.GreetingReply{
			Version: socks.V5,
			Auth:    socks.AuthNotAcceptable,
		}

		got, err := encodeGreetingReply(&greet)
		t.Require().NoError(err)

		t.Run("encodes version", func() {
			t.Equal(byte(5), got[0])
		})

		t.Run("encodes auth method", func() {
			t.Equal(byte(255), got[1])
		})
	})

	t.Run("rejects invalid version", func() {
		g := socks.GreetingReply{Version: 6}

		_, err := encodeGreetingReply(&g)
		t.Require().ErrorIs(err, socks.ErrInvalidVersion)
	})
}

func decodeGreetingReply(b []byte) (*socks.GreetingReply, error) {
	return socks.ReadGreetingReply(
		bufio.NewReader(
			bytes.NewReader(b),
		),
	)
}

func encodeGreetingReply(r *socks.GreetingReply) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
