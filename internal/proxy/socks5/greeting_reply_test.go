package socks5_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/socks5"
	"github.com/stretchr/testify/suite"
)

func TestGreetingReply(t *testing.T) {
	suite.Run(t, new(GreetingReplyTest))
}

type GreetingReplyTest struct {
	suite.Suite
}

func (t *GreetingReplyTest) TestRead() {
	t.Run("decodes an auth method", func() {
		g, err := decodeGreetingReply([]byte{5, 0xff})
		t.Require().NoError(err)

		t.Equal(socks5.AuthNotAcceptable, g.AuthMethod)
	})

	t.Run("rejects an invalid version", func() {
		_, err := decodeGreetingReply([]byte{0x06})
		t.Require().ErrorIs(err, socks5.ErrInvalidVersion)
	})
}

func (t *GreetingReplyTest) TestWrite() {
	t.Run("encodes a SOCKS5 version", func() {
		greet := socks5.GreetingReply{}

		got, err := encodeGreetingReply(&greet)
		t.Require().NoError(err)

		want := byte(0x05)
		t.Equal(want, got[0])
	})

	t.Run("encodes an auth method", func() {
		greet := socks5.GreetingReply{
			AuthMethod: socks5.AuthNotAcceptable,
		}

		got, err := encodeGreetingReply(&greet)
		t.Require().NoError(err)

		want := byte(0xff)
		t.Equal(want, got[1])
	})
}

func decodeGreetingReply(b []byte) (*socks5.GreetingReply, error) {
	return socks5.ReadGreetingReply(bufio.NewReader(bytes.NewReader(b)))
}

func encodeGreetingReply(r *socks5.GreetingReply) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
