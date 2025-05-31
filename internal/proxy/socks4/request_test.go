package socks4_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/socks4"
	"github.com/stretchr/testify/suite"
)

func TestRequest(t *testing.T) {
	suite.Run(t, new(RequestTest))
}

type RequestTest struct {
	suite.Suite
}

func (t *RequestTest) TestRead() {
	t.Run("decodes a command", func() {
		got, err := decodeRequest([]byte{4, 0x01, 0, 0, 127, 0, 0, 1, 0})
		t.Require().NoError(err)

		want := socks4.CommandConnect
		t.Equal(want, got.Command)
	})

	t.Run("decodes a destination IPv4 address", func() {
		got, err := decodeRequest([]byte{4, 0, 0, 0, 127, 0, 0, 1, 0})
		t.Require().NoError(err)

		want := "127.0.0.1"
		t.Equal(want, got.DstAddr.Host)
	})

	t.Run("decodes a destination hostname", func() {
		got, err := decodeRequest([]byte{
			4, 0, 0, 0, 0, 0, 0, 1, 0,
			'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0,
		})
		t.Require().NoError(err)

		want := "localhost"
		t.Equal(want, got.DstAddr.Host)
	})

	t.Run("decodes an empty destination address", func() {
		got, err := decodeRequest([]byte{4, 0, 0, 0, 0x00, 0x00, 0x00, 0x00, 0})
		t.Require().NoError(err)

		want := ""
		t.Equal(want, got.DstAddr.Host)
	})

	t.Run("decodes a destination port", func() {
		got, err := decodeRequest([]byte{4, 0, 0x04, 0x38, 127, 0, 0, 1, 0})
		t.Require().NoError(err)

		want := uint16(1080)
		t.Equal(want, got.DstAddr.Port)
	})

	t.Run("decodes a username", func() {
		got, err := decodeRequest([]byte{4, 0, 0, 0, 127, 0, 0, 1, 'r', 'o', 'o', 't', 0})
		t.Require().NoError(err)

		want := "root"
		t.Equal(want, got.Username)
	})

	t.Run("rejects an invalid version", func() {
		_, err := decodeRequest([]byte{0x05})
		t.Require().Error(err)
	})
}

func (t *RequestTest) TestWrite() {
	t.Run("encodes a command", func() {
		r := socks4.Request{
			Command: socks4.CommandConnect,
		}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		want := byte(0x01)
		t.Equal(want, got[1])
	})

	t.Run("encodes a SOCKS4 version", func() {
		r := socks4.Request{}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		want := byte(0x04)
		t.Equal(want, got[0])
	})

	t.Run("encodes a destination IPv4 address", func() {
		r := socks4.Request{
			DstAddr: *addr.New("127.0.0.1", 0),
		}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		want := []byte{127, 0, 0, 1}
		t.Equal(want, got[4:8])
	})

	t.Run("encodes a destination hostname", func() {
		r := socks4.Request{
			DstAddr: *addr.New("localhost", 0),
		}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		want := []byte{'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0}
		t.Equal(want, got[9:])
	})

	t.Run("encodes an destination address", func() {
		r := socks4.Request{
			DstAddr: *addr.New("", 0),
		}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		want := []byte{0, 0, 0, 0}
		t.Equal(want, got[4:8])
	})

	t.Run("encodes a destination port", func() {
		r := socks4.Request{
			DstAddr: *addr.New("", 1080),
		}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		want := []byte{0x04, 0x38}
		t.Equal(want, got[2:4])
	})

	t.Run("encodes a username", func() {
		req := socks4.Request{Username: "root"}

		got, err := encodeRequest(&req)
		t.Require().NoError(err)

		want := []byte{'r', 'o', 'o', 't', 0}
		t.Equal(want, got[8:])
	})
}

func decodeRequest(b []byte) (*socks4.Request, error) {
	return socks4.ReadRequest(
		bufio.NewReader(
			bytes.NewReader(b),
		),
	)
}

func encodeRequest(r *socks4.Request) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
