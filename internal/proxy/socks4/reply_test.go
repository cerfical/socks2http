package socks4_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/socks4"
	"github.com/stretchr/testify/suite"
)

func TestReply(t *testing.T) {
	suite.Run(t, new(ReplyTest))
}

type ReplyTest struct {
	suite.Suite
}

func (t *ReplyTest) TestRead() {
	t.Run("decodes a status", func() {
		got, err := decodeReply([]byte{0, 90, 0, 0, 0, 0, 0, 0})
		t.Require().NoError(err)

		want := socks4.StatusGranted
		t.Equal(want, got.Status)
	})

	t.Run("decodes a bind IPv4 address", func() {
		got, err := decodeReply([]byte{0, 0, 0, 0, 127, 0, 0, 1})
		t.Require().NoError(err)

		want := "127.0.0.1"
		t.Equal(want, got.BindAddr.Host)
	})

	t.Run("decodes a bind hostname", func() {
		got, err := decodeReply([]byte{
			0, 0, 0, 0, 0, 0, 0, 1,
			'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0,
		})
		t.Require().NoError(err)

		want := "localhost"
		t.Equal(want, got.BindAddr.Host)
	})

	t.Run("decodes an empty bind address", func() {
		got, err := decodeReply([]byte{0, 0, 0, 0, 0x00, 0x00, 0x00, 0x00})
		t.Require().NoError(err)

		want := ""
		t.Equal(want, got.BindAddr.Host)
	})

	t.Run("decodes a bind port", func() {
		got, err := decodeReply([]byte{0, 0, 0x04, 0x38, 0, 0, 0, 0})
		t.Require().NoError(err)

		want := uint16(1080)
		t.Equal(want, got.BindAddr.Port)
	})

	t.Run("rejects an invalid reply version", func() {
		_, err := decodeReply([]byte{0x01})
		t.Error(err)
	})
}

func (t *ReplyTest) TestWrite() {
	t.Run("encodes a status", func() {
		r := socks4.Reply{
			Status: socks4.StatusGranted,
		}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := byte(90)
		t.Equal(want, got[1])
	})

	t.Run("encodes a reply version", func() {
		r := socks4.Reply{}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := byte(0x00)
		t.Equal(want, got[0])
	})

	t.Run("encodes a bind IPv4 address", func() {
		r := socks4.Reply{
			BindAddr: *addr.New("127.0.0.1", 0),
		}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := []byte{127, 0, 0, 1}
		t.Equal(want, got[4:8])
	})

	t.Run("encodes a bind hostname", func() {
		r := socks4.Reply{
			BindAddr: *addr.New("localhost", 0),
		}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := []byte{'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0}
		t.Equal(want, got[8:])
	})

	t.Run("encodes an empty bind address", func() {
		r := socks4.Reply{
			BindAddr: *addr.New("", 0),
		}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := []byte{0, 0, 0, 0}
		t.Equal(want, got[4:8])
	})

	t.Run("encodes a bind port", func() {
		r := socks4.Reply{
			BindAddr: *addr.New("", 1080),
		}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := []byte{0x04, 0x38}
		t.Equal(want, got[2:4])
	})
}

func decodeReply(b []byte) (*socks4.Reply, error) {
	return socks4.ReadReply(
		bufio.NewReader(
			bytes.NewReader(b),
		),
	)
}

func encodeReply(r *socks4.Reply) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
