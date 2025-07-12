package socks_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/socks"
	"github.com/stretchr/testify/suite"
)

func TestReply(t *testing.T) {
	suite.Run(t, new(ReplyTest))
}

type ReplyTest struct {
	suite.Suite
}

func (t *ReplyTest) TestRead() {
	t.Run("decodes SOCKS4 reply", func() {
		got, err := decodeReply([]byte{
			0,          // Version
			90,         // Status
			0x04, 0x38, // Bind port
			127, 0, 0, 1, // Bind address
		})
		t.Require().NoError(err)

		t.Run("decodes version", func() {
			t.Equal(socks.V4, got.Version)
		})

		t.Run("decodes status", func() {
			t.Equal(socks.StatusGranted, got.Status)
		})

		t.Run("decodes bind IPv4 address", func() {
			t.Equal("127.0.0.1", got.BindAddr.Host)
		})

		t.Run("decodes bind port", func() {
			t.Equal(uint16(1080), got.BindAddr.Port)
		})

		t.Run("decodes bind hostname", func() {
			got, err := decodeReply([]byte{
				0, 90, 0x04, 0x38, 0, 0, 0, 1,
				'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0,
			})
			t.Require().NoError(err)

			t.Equal("localhost", got.BindAddr.Host)
		})

		t.Run("decodes empty bind address", func() {
			got, err := decodeReply([]byte{0, 90, 0x04, 0x38, 0, 0, 0, 0})
			t.Require().NoError(err)

			t.Equal("", got.BindAddr.Host)
		})
	})

	t.Run("decodes SOCKS5 reply", func() {
		got, err := decodeReply([]byte{
			5,               // Version
			1,               // Status
			0,               // Reserved field
			1, 127, 0, 0, 1, // Bind address
			0x04, 0x38, // Bind port
		})
		t.Require().NoError(err)

		t.Run("decodes version", func() {
			t.Equal(socks.V5, got.Version)
		})

		t.Run("decodes status", func() {
			t.Equal(socks.StatusGeneralFailure, got.Status)
		})

		t.Run("decodes bind IPv4 address", func() {
			t.Equal("127.0.0.1", got.BindAddr.Host)
		})

		t.Run("decodes bind port", func() {
			t.Equal(uint16(1080), got.BindAddr.Port)
		})

		t.Run("decodes bind IPv6 address", func() {
			got, err := decodeReply([]byte{
				5, 1, 0,
				4, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
				0x04, 0x38,
			})
			t.Require().NoError(err)

			t.Equal("102:304:506:708:90a:b0c:d0e:f10", got.BindAddr.Host)
		})

		t.Run("decodes bind hostname", func() {
			got, err := decodeReply([]byte{
				5, 1, 0,
				0x03, 9, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't',
				0x04, 0x38,
			})
			t.Require().NoError(err)

			t.Equal("localhost", got.BindAddr.Host)
		})

		t.Run("rejects invalid address type", func() {
			_, err := decodeReply([]byte{5, 1, 0, 2})
			t.Require().Error(err)
		})

		t.Run("rejects non-zero reserved field", func() {
			_, err := decodeReply([]byte{5, 1, 1})
			t.Require().Error(err)
		})
	})

	t.Run("rejects invalid version", func() {
		_, err := decodeReply([]byte{1})
		t.ErrorIs(err, socks.ErrInvalidVersion)
	})
}

func (t *ReplyTest) TestWrite() {
	t.Run("encodes SOCKS4 reply", func() {
		r := socks.Reply{
			Version:  socks.V4,
			Status:   socks.StatusGranted,
			BindAddr: *addr.NewAddr("127.0.0.1", 1080),
		}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		t.Run("encodes version", func() {
			t.Equal(byte(0), got[0])
		})

		t.Run("encodes status", func() {
			t.Equal(byte(90), got[1])
		})

		t.Run("encodes bind IPv4 address", func() {
			t.Equal([]byte{127, 0, 0, 1}, got[4:8])
		})

		t.Run("encodes bind port", func() {
			t.Equal([]byte{0x04, 0x38}, got[2:4])
		})

		t.Run("encodes bind hostname", func() {
			r := socks.Reply{
				Version:  socks.V4,
				BindAddr: *addr.NewAddr("localhost", 0),
			}

			got, err := encodeReply(&r)
			t.Require().NoError(err)
			t.Equal([]byte{'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0}, got[8:])
		})

		t.Run("encodes empty bind address", func() {
			r := socks.Reply{
				Version:  socks.V4,
				BindAddr: *addr.NewAddr("", 0),
			}

			got, err := encodeReply(&r)
			t.Require().NoError(err)
			t.Equal([]byte{0, 0, 0, 0}, got[4:8])
		})
	})

	t.Run("encodes SOCKS5 reply", func() {
		r := socks.Reply{
			Version:  socks.V5,
			Status:   socks.StatusGeneralFailure,
			BindAddr: *addr.NewAddr("127.0.0.1", 1080),
		}

		got, err := encodeReply(&r)
		t.Require().NoError(err)

		t.Run("encodes status", func() {
			t.Equal(byte(1), got[1])
		})

		t.Run("encodes version", func() {
			t.Equal(byte(5), got[0])
		})

		t.Run("encodes zero reserved field", func() {
			t.Equal(byte(0), got[2])
		})

		t.Run("encodes bind IPv4 address", func() {
			t.Equal([]byte{0x01, 127, 0, 0, 1}, got[3:8])
		})

		t.Run("encodes bind port", func() {
			t.Equal([]byte{0x04, 0x38}, got[8:])
		})

		t.Run("encodes bind IPv6 address", func() {
			r := socks.Reply{
				Version:  socks.V5,
				BindAddr: *addr.NewAddr("102:304:506:708:90a:b0c:d0e:f10", 0),
			}

			got, err := encodeReply(&r)
			t.Require().NoError(err)

			want := []byte{
				4,
				0x01, 0x02, 0x03, 0x04,
				0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c,
				0x0d, 0x0e, 0x0f, 0x10,
			}
			t.Equal(want, got[3:20])
		})

		t.Run("encodes bind hostname", func() {
			r := socks.Reply{
				Version:  socks.V5,
				BindAddr: *addr.NewAddr("localhost", 0),
			}

			got, err := encodeReply(&r)
			t.Require().NoError(err)

			t.Equal([]byte{3, 9, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't'}, got[3:14])
		})

		t.Run("rejects a bind hostname longer than 255 bytes", func() {
			r := socks.Reply{
				Version:  socks.V5,
				BindAddr: *addr.NewAddr(strings.Repeat("a", 256), 0),
			}

			_, err := encodeReply(&r)
			t.Require().Error(err)
		})
	})

	t.Run("rejects invalid version", func() {
		r := socks.Reply{Version: 10}

		_, err := encodeReply(&r)
		t.Require().ErrorIs(err, socks.ErrInvalidVersion)
	})
}

func decodeReply(b []byte) (*socks.Reply, error) {
	return socks.ReadReply(
		bufio.NewReader(
			bytes.NewReader(b),
		),
	)
}

func encodeReply(r *socks.Reply) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
