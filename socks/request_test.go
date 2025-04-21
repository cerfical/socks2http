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
	t.Run("marks requests with a destination IPv4 address as SOCKS4", func(t *testing.T) {
		got, err := decodeSOCKSRequest([]byte{
			SOCKSVersion4, ConnectCommand, 0x04, 0x38, 127, 0, 0, 1, 0,
		})
		require.NoError(t, err)

		want := socks.V4
		assert.Equal(t, want, got.Version)
	})

	t.Run("marks requests with a destination hostname as SOCKS4a", func(t *testing.T) {
		got, err := decodeSOCKSRequest([]byte{
			SOCKSVersion4, ConnectCommand, 0x04, 0x38, 0, 0, 0, 0, 0,
			'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0,
		})
		require.NoError(t, err)

		want := socks.V4a
		assert.Equal(t, want, got.Version)
	})

	t.Run("rejects invalid SOCKS versions", func(t *testing.T) {
		_, err := decodeSOCKSRequest([]byte{0x05})
		require.Error(t, err)
	})
}

func TestReadRequest_SOCKS4(t *testing.T) {
	validCommands := map[string]struct {
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

	t.Run("rejects invalid command codes", func(t *testing.T) {
		_, err := decodeSOCKSRequest([]byte{
			SOCKSVersion4, 0x03, 0x04, 0x38, 127, 0, 0, 1, 0,
		})
		require.Error(t, err)
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

	t.Run("rejects an empty destination hostname", func(t *testing.T) {
		_, err := decodeSOCKSRequest([]byte{
			SOCKSVersion4, ConnectCommand, 0x04, 0x38, 0, 0, 0, 1, 0,
			0,
		})
		require.Error(t, err)
	})
}

func TestRequest_Write(t *testing.T) {
	t.Run("rejects invalid SOCKS versions", func(t *testing.T) {
		req := socks.NewRequest(120, socks.Connect, addr.NewHost("127.0.0.1", 1080))

		_, err := encodeSOCKSRequest(req)
		require.Error(t, err)
	})
}

func TestRequest_Write_SOCKS4(t *testing.T) {
	validCommands := map[string]struct {
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
	for name, test := range validCommands {
		t.Run(name, func(t *testing.T) {
			req := socks.NewRequest(socks.V4, test.command, addr.NewHost("127.0.0.1", 1080))

			got, err := encodeSOCKSRequest(req)
			require.NoError(t, err)

			assert.Equal(t, test.want, got[1])
		})
	}

	errors := map[string]struct {
		command socks.Command
		dstAddr *addr.Host
	}{
		"rejects an empty destination address": {
			command: socks.Connect,
			dstAddr: addr.NewHost("", 0),
		},

		"rejects non-IPv4 destination addresses": {
			command: socks.Connect,
			dstAddr: addr.NewHost("localhost", 1080),
		},

		"rejects invalid command codes": {
			command: 0x03,
			dstAddr: addr.NewHost("127.0.0.1", 1080),
		},
	}
	for name, test := range errors {
		t.Run(name, func(t *testing.T) {
			req := socks.NewRequest(socks.V4, test.command, test.dstAddr)

			_, err := encodeSOCKSRequest(req)
			require.Error(t, err)
		})
	}

	t.Run("encodes non-empty destination addresses", func(t *testing.T) {
		req := socks.NewRequest(socks.V4, socks.Connect, addr.NewHost("127.0.0.1", 1080))

		got, err := encodeSOCKSRequest(req)
		require.NoError(t, err)

		want := []byte{0x04, 0x38, 127, 0, 0, 1}
		assert.Equal(t, want, got[2:8])
	})

	t.Run("encodes a non-empty username", func(t *testing.T) {
		req := socks.NewRequest(socks.V4, socks.Connect, addr.NewHost("127.0.0.1", 1080))
		req.Username = "root"

		got, err := encodeSOCKSRequest(req)
		require.NoError(t, err)

		want := []byte{'r', 'o', 'o', 't', 0}
		assert.Equal(t, want, got[8:])
	})
}

func TestRequest_Write_SOCKS4a(t *testing.T) {
	t.Run("encodes a non-empty destination hostname", func(t *testing.T) {
		req := socks.NewRequest(socks.V4a, socks.Connect, addr.NewHost("localhost", 1080))

		got, err := encodeSOCKSRequest(req)
		require.NoError(t, err)

		want := []byte{'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't', 0}
		assert.Equal(t, want, got[9:])
	})

	t.Run("rejects an empty destination hostname", func(t *testing.T) {
		req := socks.NewRequest(socks.V4a, socks.Connect, addr.NewHost("", 1080))

		_, err := encodeSOCKSRequest(req)
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

func encodeSOCKSRequest(r *socks.Request) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
