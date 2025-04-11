package socks_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/socks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	SOCKSVersion4 = 0x04

	ConnectCommand = 0x01
	BindCommand    = 0x02
)

func TestReadRequest(t *testing.T) {
	t.Run("rejects unsupported SOCKS versions", func(t *testing.T) {
		_, err := decodeSOCKSRequest([]byte{0x05})
		require.Error(t, err)
	})
}

func TestReadRequest_SOCKS4(t *testing.T) {
	tests := map[string]struct {
		input byte
		want  socks.Command
	}{
		"decodes a CONNECT command": {
			input: ConnectCommand,
			want:  socks.Connect,
		},

		"decodes a BIND command": {
			input: BindCommand,
			want:  socks.Bind,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := decodeSOCKSRequest([]byte{SOCKSVersion4, test.input, 0x04, 0x38, 127, 0, 0, 1, 0})
			require.NoError(t, err)

			want := socks.NewRequest(socks.V4, test.want, addr.NewHost("127.0.0.1", 1080))
			assert.Equal(t, want, got)
		})
	}

	t.Run("rejects unsupported commands", func(t *testing.T) {
		_, err := decodeSOCKSReply([]byte{SOCKSVersion4, 0x03, 0x04, 0x38, 127, 0, 0, 1, 0})
		require.Error(t, err)
	})
}

func TestRequest_Write_SOCKS4(t *testing.T) {
	tests := map[string]struct {
		command socks.Command
		want    byte
	}{
		"encodes a CONNECT command": {
			command: socks.Connect,
			want:    ConnectCommand,
		},

		"encodes a BIND command": {
			command: socks.Bind,
			want:    BindCommand,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := encodeSOCKSRequest(socks.V4, test.command, addr.NewHost("0.0.0.0", 0))
			require.NoError(t, err)

			want := []byte{SOCKSVersion4, test.want, 0, 0, 0, 0, 0, 0, 0}
			assert.Equal(t, want, got)
		})
	}

	t.Run("encodes a non-empty destination address", func(t *testing.T) {
		got, err := encodeSOCKSRequest(socks.V4, socks.Connect, addr.NewHost("127.0.0.1", 1080))
		require.NoError(t, err)

		want := []byte{SOCKSVersion4, ConnectCommand, 0x04, 0x38, 127, 0, 0, 1, 0}
		assert.Equal(t, want, got)
	})

	t.Run("rejects an empty destination address", func(t *testing.T) {
		_, err := encodeSOCKSRequest(socks.V4, socks.Connect, nil)
		require.Error(t, err)
	})

	t.Run("rejects destination addresses specified as non-IPv4 address", func(t *testing.T) {
		_, err := encodeSOCKSRequest(socks.V4, socks.Connect, addr.NewHost("localhost", 0))
		require.Error(t, err)
	})
}

func decodeSOCKSRequest(b []byte) (*socks.Request, error) {
	return socks.ReadRequest(
		bufio.NewReader(
			bytes.NewReader(b),
		),
	)
}

func encodeSOCKSRequest(v socks.Version, c socks.Command, dstAddr *addr.Host) ([]byte, error) {
	req := socks.NewRequest(v, c, dstAddr)

	var buf bytes.Buffer
	if err := req.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
