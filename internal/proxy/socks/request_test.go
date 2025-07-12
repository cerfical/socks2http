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

func TestRequest(t *testing.T) {
	suite.Run(t, new(RequestTest))
}

type RequestTest struct {
	suite.Suite
}

func (t *RequestTest) TestRead() {
	t.Run("decodes SOCKS4 request", func() {
		got, err := decodeRequest([]byte{
			4,          // Version
			1,          // Command
			0x04, 0x38, // Port
			127, 0, 0, 1, // Address
			'r', 'o', 'o', 't', 0, // Username
		})
		t.Require().NoError(err)

		t.Run("decodes command", func() {
			t.Equal(socks.CommandConnect, got.Command)
		})

		t.Run("decodes version", func() {
			t.Equal(socks.V4, got.Version)
		})

		t.Run("decodes destination IPv4 address", func() {
			t.Equal("127.0.0.1", got.DstAddr.Host)
		})

		t.Run("decodes destination port", func() {
			t.Equal(uint16(1080), got.DstAddr.Port)
		})

		t.Run("decodes username", func() {
			t.Equal("root", got.Username)
		})

		t.Run("decodes destination hostname", func() {
			got, err := decodeRequest([]byte{
				4, 1, 0x04, 0x38, 0, 0, 0, 1, 0,
				'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0,
			})
			t.Require().NoError(err)

			t.Equal("localhost", got.DstAddr.Host)
		})

		t.Run("decodes empty destination address", func() {
			got, err := decodeRequest([]byte{
				4, 1, 0x04, 0x38,
				0, 0, 0, 0,
				0,
			})
			t.Require().NoError(err)

			t.Equal("", got.DstAddr.Host)
		})
	})

	t.Run("decodes SOCKS5 request", func() {
		got, err := decodeRequest([]byte{
			5,               // Version
			1,               // Command
			0,               // Reserved
			1, 127, 0, 0, 1, // Address
			0x04, 0x38, // Port
		})
		t.Require().NoError(err)

		t.Run("decodes command", func() {
			t.Equal(socks.CommandConnect, got.Command)
		})

		t.Run("decodes destination IPv4 address", func() {
			t.Equal("127.0.0.1", got.DstAddr.Host)
		})

		t.Run("decodes destination port", func() {
			t.Equal(uint16(1080), got.DstAddr.Port)
		})

		t.Run("decodes destination IPv6 address", func() {
			got, err := decodeRequest([]byte{
				5, 0, 0,
				4, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
				0, 0,
			})
			t.Require().NoError(err)

			t.Equal("102:304:506:708:90a:b0c:d0e:f10", got.DstAddr.Host)
		})

		t.Run("decodes destination hostname", func() {
			got, err := decodeRequest([]byte{
				5, 0, 0,
				3, 9, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't',
				0, 0,
			})
			t.Require().NoError(err)

			t.Equal("localhost", got.DstAddr.Host)
		})

		t.Run("rejects invalid address type", func() {
			_, err := decodeRequest([]byte{5, 0, 0, 2})
			t.Require().Error(err)
		})

		t.Run("rejects non-zero reserved field", func() {
			_, err := decodeRequest([]byte{5, 0, 1})
			t.Require().Error(err)
		})
	})

	t.Run("rejects invalid version", func() {
		_, err := decodeRequest([]byte{6})
		t.Require().ErrorIs(err, socks.ErrInvalidVersion)
	})
}

func (t *RequestTest) TestWrite() {
	t.Run("encodes SOCKS4 request", func() {
		r := socks.Request{
			Version:  socks.V4,
			Command:  socks.CommandConnect,
			DstAddr:  *addr.NewAddr("127.0.0.1", 1080),
			Username: "root",
		}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		t.Run("encodes version", func() {
			t.Equal(byte(4), got[0])
		})

		t.Run("encodes command", func() {
			t.Equal(byte(1), got[1])
		})

		t.Run("encodes destination IPv4 address", func() {
			t.Equal([]byte{127, 0, 0, 1}, got[4:8])
		})

		t.Run("encodes destination port", func() {
			t.Equal([]byte{0x04, 0x38}, got[2:4])
		})

		t.Run("encodes destination hostname", func() {
			r := socks.Request{
				Version: socks.V4,
				DstAddr: *addr.NewAddr("localhost", 0),
			}

			got, err := encodeRequest(&r)
			t.Require().NoError(err)
			t.Equal([]byte{'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0}, got[9:])
		})

		t.Run("encodes empty destination address", func() {
			r := socks.Request{
				Version: socks.V4,
				DstAddr: *addr.NewAddr("", 0),
			}

			got, err := encodeRequest(&r)
			t.Require().NoError(err)
			t.Equal([]byte{0, 0, 0, 0}, got[4:8])
		})

		t.Run("encodes username", func() {
			t.Equal([]byte{'r', 'o', 'o', 't', 0}, got[8:])
		})
	})

	t.Run("encodes SOCKS5 request", func() {
		r := socks.Request{
			Version: socks.V5,
			Command: socks.CommandConnect,
			DstAddr: *addr.NewAddr("127.0.0.1", 1080),
		}

		got, err := encodeRequest(&r)
		t.Require().NoError(err)

		t.Run("encodes version", func() {
			t.Equal(byte(5), got[0])
		})

		t.Run("encodes command", func() {
			t.Equal(byte(1), got[1])
		})

		t.Run("encodes zero reserved field", func() {
			t.Equal(byte(0), got[2])
		})

		t.Run("encodes destination IPv4 address", func() {
			t.Equal([]byte{0x01, 127, 0, 0, 1}, got[3:8])
		})

		t.Run("encodes destination port", func() {
			t.Equal([]byte{0x04, 0x38}, got[8:])
		})

		t.Run("encodes destination IPv6 address", func() {
			r := socks.Request{
				Version: socks.V5,
				DstAddr: *addr.NewAddr("102:304:506:708:90a:b0c:d0e:f10", 0),
			}

			got, err := encodeRequest(&r)
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

		t.Run("encodes destination hostname", func() {
			r := socks.Request{
				Version: socks.V5,
				DstAddr: *addr.NewAddr("localhost", 0),
			}

			got, err := encodeRequest(&r)
			t.Require().NoError(err)
			t.Equal([]byte{0x03, 9, 'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't'}, got[3:14])
		})

		t.Run("rejects destination hostname longer than 255 bytes", func() {
			r := socks.Request{
				Version: socks.V5,
				DstAddr: *addr.NewAddr(strings.Repeat("a", 256), 0),
			}

			_, err := encodeRequest(&r)
			t.Require().Error(err)
		})
	})

	t.Run("rejects invalid version", func() {
		r := socks.Request{Version: 10}

		_, err := encodeRequest(&r)
		t.Require().ErrorIs(err, socks.ErrInvalidVersion)
	})
}

func decodeRequest(b []byte) (*socks.Request, error) {
	return socks.ReadRequest(
		bufio.NewReader(
			bytes.NewReader(b),
		),
	)
}

func encodeRequest(r *socks.Request) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
