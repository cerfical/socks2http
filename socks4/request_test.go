package socks4_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/socks4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	SOCKSVersion4 = 0x04

	ConnectCommand = 0x01
	BindCommand    = 0x02
)

func TestReadRequest(t *testing.T) {
	t.Run("rejects an invalid version", func(t *testing.T) {
		_, err := decodeSOCKSRequest([]byte{0x05})
		require.Error(t, err)
	})
}

func TestReadRequest_SOCKS4(t *testing.T) {
	validCommands := map[string]struct {
		input byte
		want  socks4.Command
	}{
		"decodes a CONNECT command": {
			input: ConnectCommand,
			want:  socks4.CommandConnect,
		},

		"decodes a BIND command": {
			input: BindCommand,
			want:  socks4.CommandBind,
		},
	}
	for name, test := range validCommands {
		t.Run(name, func(t *testing.T) {
			got, err := decodeSOCKSRequest([]byte{SOCKSVersion4, test.input, 0x04, 0x38, 127, 0, 0, 1, 0})
			require.NoError(t, err)

			assert.Equal(t, test.want, got.Command)
		})
	}

	t.Run("decodes a non-empty destination address", func(t *testing.T) {
		got, err := decodeSOCKSRequest([]byte{SOCKSVersion4, ConnectCommand, 0x04, 0x38, 127, 0, 0, 1, 0})
		require.NoError(t, err)

		want := addr.NewHost("127.0.0.1", 1080)
		assert.Equal(t, want, &got.DstAddr)
	})

	t.Run("decodes a non-empty username", func(t *testing.T) {
		got, err := decodeSOCKSRequest([]byte{
			SOCKSVersion4, ConnectCommand, 0x04, 0x38, 127, 0, 0, 1,
			'r', 'o', 'o', 't', 0,
		})
		require.NoError(t, err)

		want := "root"
		assert.Equal(t, want, got.Username)
	})
}

func TestReadRequest_SOCKS4a(t *testing.T) {
	t.Run("decodes a non-empty destination hostname", func(t *testing.T) {
		got, err := decodeSOCKSRequest([]byte{
			SOCKSVersion4, ConnectCommand, 0x04, 0x38, 0, 0, 0, 1, 0,
			'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0,
		})
		require.NoError(t, err)

		want := "localhost"
		assert.Equal(t, want, got.DstAddr.Hostname)
	})
}

func TestRequest_Write_SOCKS4(t *testing.T) {
	validCommands := map[string]struct {
		command socks4.Command
		want    byte
	}{
		"encodes a CONNECT command": {
			command: socks4.CommandConnect,
			want:    ConnectCommand,
		},

		"encodes a BIND command": {
			command: socks4.CommandBind,
			want:    BindCommand,
		},
	}
	for name, test := range validCommands {
		t.Run(name, func(t *testing.T) {
			req := socks4.Request{
				Command: test.command,
				DstAddr: *addr.NewHost("127.0.0.1", 1080),
			}

			got, err := encodeSOCKSRequest(&req)
			require.NoError(t, err)

			assert.Equal(t, test.want, got[1])
		})
	}

	t.Run("encodes an IPv4 destination address", func(t *testing.T) {
		req := socks4.Request{
			Command: socks4.CommandConnect,
			DstAddr: *addr.NewHost("127.0.0.1", 1080),
		}

		got, err := encodeSOCKSRequest(&req)
		require.NoError(t, err)

		want := []byte{0x04, 0x38, 127, 0, 0, 1}
		assert.Equal(t, want, got[2:8])
	})

	t.Run("encodes a non-IPv4 destination address", func(t *testing.T) {
		req := socks4.Request{
			Command: socks4.CommandConnect,
			DstAddr: *addr.NewHost("localhost", 1080),
		}

		got, err := encodeSOCKSRequest(&req)
		require.NoError(t, err)

		want := []byte{'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0}
		assert.Equal(t, want, got[9:])
	})

	t.Run("encodes an username", func(t *testing.T) {
		req := socks4.Request{
			Command: socks4.CommandConnect,
			DstAddr: *addr.NewHost("127.0.0.1", 1080),
		}
		req.Username = "root"

		got, err := encodeSOCKSRequest(&req)
		require.NoError(t, err)

		want := []byte{'r', 'o', 'o', 't', 0}
		assert.Equal(t, want, got[8:])
	})
}

func decodeSOCKSRequest(b []byte) (*socks4.Request, error) {
	return socks4.ReadRequest(
		bufio.NewReader(
			bytes.NewReader(b),
		),
	)
}

func encodeSOCKSRequest(r *socks4.Request) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
