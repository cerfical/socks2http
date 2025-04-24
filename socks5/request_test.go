package socks5_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/socks5"
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
		got, err := decodeRequest([]byte{5, 0x01, 0, 3, 0, 0, 0})
		t.Require().NoError(err)

		want := socks5.CommandConnect
		t.Equal(want, got.Command)
	})

	t.Run("decodes a destination IPv4 address", func() {
		got, err := decodeRequest([]byte{5, 0, 0, 0x01, 127, 0, 0, 1, 0, 0})
		t.Require().NoError(err)

		want := "127.0.0.1"
		t.Equal(want, got.DstAddr.Hostname)
	})

	t.Run("decodes a destination hostname", func() {
		got, err := decodeRequest([]byte{5, 0, 0, 0x03, 9, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0, 0})
		t.Require().NoError(err)

		want := "localhost"
		t.Equal(want, got.DstAddr.Hostname)
	})

	t.Run("decodes a destination port", func() {
		got, err := decodeRequest([]byte{5, 0, 0, 3, 0, 0x04, 0x38})
		t.Require().NoError(err)

		want := uint16(1080)
		t.Equal(want, got.DstAddr.Port)
	})

	errors := map[string]struct {
		input []byte
		want  error
	}{
		"rejects an invalid version": {
			input: []byte{0x06},
			want:  socks5.ErrInvalidVersion,
		},

		"rejects an invalid address type": {
			input: []byte{5, 0, 0, 0x02},
			want:  socks5.ErrInvalidAddrType,
		},

		"rejects reserved field values other than 0": {
			input: []byte{5, 0, 0x01},
			want:  socks5.ErrNonZeroReservedField,
		},
	}
	for name, test := range errors {
		t.Run(name, func() {
			_, err := decodeRequest(test.input)
			t.Require().ErrorIs(err, test.want)
		})
	}
}

func (t *RequestTest) TestWrite() {
	t.Run("encodes a command", func() {
		r := socks5.Request{
			Command: socks5.CommandConnect,
		}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		want := byte(0x01)
		t.Equal(want, got[1])
	})

	t.Run("encodes a SOCKS5 version", func() {
		r := socks5.Request{}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		want := byte(0x05)
		t.Equal(want, got[0])
	})

	t.Run("encodes a zero-valued reserved field", func() {
		r := socks5.Request{}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		want := byte(0x00)
		t.Equal(want, got[2])
	})

	t.Run("encodes a destination IPv4 address", func() {
		r := socks5.Request{
			DstAddr: *addr.NewHost("127.0.0.1", 0),
		}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		want := []byte{0x01, 127, 0, 0, 1}
		t.Equal(want, got[3:8])
	})

	t.Run("encodes a destination hostname", func() {
		r := socks5.Request{
			DstAddr: *addr.NewHost("localhost", 0),
		}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		want := []byte{0x03, 9, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't'}
		t.Equal(want, got[3:14])
	})

	t.Run("encodes a destination port", func() {
		r := socks5.Request{
			DstAddr: *addr.NewHost("", 1080),
		}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		want := []byte{0x04, 0x38}
		t.Equal(want, got[5:])
	})

	t.Run("rejects a destination hostname longer than 255 bytes", func() {
		r := socks5.Request{
			DstAddr: *addr.NewHost(strings.Repeat("a", 256), 0),
		}

		_, err := encodeRequest(&r)
		t.Require().ErrorIs(err, socks5.ErrHostnameTooLong)
	})
}

func decodeRequest(b []byte) (*socks5.Request, error) {
	return socks5.ReadRequest(bufio.NewReader(bytes.NewReader(b)))
}

func encodeRequest(r *socks5.Request) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
