package socks5_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/socks5"
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
		got, err := decodeReply([]byte{5, 0x01, 0, 3, 0, 0, 0})
		t.Require().NoError(err)

		want := socks5.StatusGeneralFailure
		t.Equal(want, got.Status)
	})

	t.Run("decodes a bind IPv4 address", func() {
		got, err := decodeReply([]byte{5, 0, 0, 0x01, 127, 0, 0, 1, 0, 0})
		t.Require().NoError(err)

		want := "127.0.0.1"
		t.Equal(want, got.BindAddr.Host)
	})

	t.Run("decodes a bind IPv6 address", func() {
		got, err := decodeReply([]byte{5, 0, 0,
			0x04, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
			0, 0,
		})
		t.Require().NoError(err)

		want := "102:304:506:708:90a:b0c:d0e:f10"
		t.Equal(want, got.BindAddr.Host)
	})

	t.Run("decodes a bind hostname", func() {
		got, err := decodeReply([]byte{5, 0, 0, 0x03, 9, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0, 0})
		t.Require().NoError(err)

		want := "localhost"
		t.Equal(want, got.BindAddr.Host)
	})

	t.Run("decodes a bind port", func() {
		got, err := decodeReply([]byte{5, 0, 0, 3, 0, 0x04, 0x38})
		t.Require().NoError(err)

		want := uint16(1080)
		t.Equal(want, got.BindAddr.Port)
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
			_, err := decodeReply(test.input)
			t.Require().ErrorIs(err, test.want)
		})
	}
}

func (t *ReplyTest) TestWrite() {
	t.Run("encodes a status", func() {
		r := socks5.Reply{
			Status: socks5.StatusGeneralFailure,
		}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := byte(0x01)
		t.Equal(want, got[1])
	})

	t.Run("encodes a SOCKS5 version", func() {
		r := socks5.Reply{}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := byte(0x05)
		t.Equal(want, got[0])
	})

	t.Run("encodes a zero-valued reserved field", func() {
		r := socks5.Reply{}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := byte(0x00)
		t.Equal(want, got[2])
	})

	t.Run("encodes a bind IPv4 address", func() {
		r := socks5.Reply{
			BindAddr: *addr.NewAddr("127.0.0.1", 0),
		}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := []byte{0x01, 127, 0, 0, 1}
		t.Equal(want, got[3:8])
	})

	t.Run("encodes a bind IPv6 address", func() {
		r := socks5.Reply{
			BindAddr: *addr.NewAddr("102:304:506:708:90a:b0c:d0e:f10", 0),
		}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := []byte{
			0x04, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		}
		t.Equal(want, got[3:20])
	})

	t.Run("encodes a bind hostname", func() {
		r := socks5.Reply{
			BindAddr: *addr.NewAddr("localhost", 0),
		}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := []byte{0x03, 9, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't'}
		t.Equal(want, got[3:14])
	})

	t.Run("encodes a bind port", func() {
		r := socks5.Reply{
			BindAddr: *addr.NewAddr("", 1080),
		}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		want := []byte{0x04, 0x38}
		t.Equal(want, got[5:])
	})

	t.Run("rejects a bind hostname longer than 255 bytes", func() {
		r := socks5.Reply{
			BindAddr: *addr.NewAddr(strings.Repeat("a", 256), 0),
		}

		_, err := encodeReply(&r)
		t.Require().ErrorIs(err, socks5.ErrHostnameTooLong)
	})
}

func decodeReply(b []byte) (*socks5.Reply, error) {
	return socks5.ReadReply(bufio.NewReader(bytes.NewReader(b)))
}

func encodeReply(r *socks5.Reply) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
