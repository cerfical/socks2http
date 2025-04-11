package socks_test

import (
	"bufio"
	"bytes"
	"slices"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/socks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ReplyVersion = 0

	RequestGranted            = 90
	RequestRejectedOrFailed   = 91
	RequestRejectedNoAuth     = 92
	RequestRejectedAuthFailed = 93
)

func TestReadReply_SOCKS4(t *testing.T) {
	tests := map[string]struct {
		status byte
		want   socks.Status
	}{
		"decodes a Request-Granted reply": {
			status: RequestGranted,
			want:   socks.Granted,
		},

		"decodes a Request-Rejected reply": {
			status: RequestRejectedOrFailed,
			want:   socks.Rejected,
		},

		"decodes a No-Auth reply": {
			status: RequestRejectedNoAuth,
			want:   socks.NoAuth,
		},

		"decodes an Auth-Failed reply": {
			status: RequestRejectedAuthFailed,
			want:   socks.AuthFailed,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := decodeSOCKSReply([]byte{ReplyVersion, test.status, 0, 0, 0, 0, 0, 0})
			require.NoError(t, err)

			assert.Equal(t, test.want, got.Status)
		})
	}

	t.Run("decodes a non-empty BIND address", func(t *testing.T) {
		got, err := decodeSOCKSReply(slices.Concat(
			[]byte{ReplyVersion, RequestGranted, 0x04, 0x38, 127, 0, 0, 1},
		))
		require.NoError(t, err)

		want := addr.NewHost("127.0.0.1", 1080)
		assert.Equal(t, want, &got.BindAddr)
	})

	t.Run("decodes an empty BIND address to an empty hostname", func(t *testing.T) {
		got, err := decodeSOCKSReply(slices.Concat(
			[]byte{ReplyVersion, RequestGranted, 0, 0, 0, 0, 0, 0},
		))
		require.NoError(t, err)

		assert.Equal(t, "", got.BindAddr.Hostname)
	})

	t.Run("rejects replies with unsupported status", func(t *testing.T) {
		_, err := decodeSOCKSReply([]byte{ReplyVersion, 0x5e, 0, 0, 0, 0, 0, 0})
		assert.Error(t, err)
	})

	t.Run("rejects replies with unsupported reply version", func(t *testing.T) {
		_, err := decodeSOCKSReply([]byte{1})
		assert.Error(t, err)
	})
}

func TestReply_Write_SOCKS4(t *testing.T) {
	tests := map[string]struct {
		status socks.Status
		want   byte
	}{
		"encodes a Request-Granted reply": {
			status: socks.Granted,
			want:   RequestGranted,
		},

		"encodes a Request-Rejected reply": {
			status: socks.Rejected,
			want:   RequestRejectedOrFailed,
		},

		"encodes a No-Auth reply": {
			status: socks.NoAuth,
			want:   RequestRejectedNoAuth,
		},

		"encodes an Auth-Failed reply": {
			status: socks.AuthFailed,
			want:   RequestRejectedAuthFailed,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := encodeSOCKSReply(test.status, nil)
			require.NoError(t, err)

			want := []byte{ReplyVersion, test.want, 0, 0, 0, 0, 0, 0}
			assert.Equal(t, want, got)
		})
	}

	t.Run("encodes a non-empty BIND address", func(t *testing.T) {
		got, err := encodeSOCKSReply(socks.Granted, addr.NewHost("127.0.0.1", 1080))
		require.NoError(t, err)

		want := []byte{ReplyVersion, RequestGranted, 0x04, 0x38, 127, 0, 0, 1}
		assert.Equal(t, want, got)
	})

	t.Run("ignores an empty BIND address", func(t *testing.T) {
		got, err := encodeSOCKSReply(socks.Granted, nil)
		require.NoError(t, err)

		want := []byte{ReplyVersion, RequestGranted, 0, 0, 0, 0, 0, 0}
		assert.Equal(t, want, got)
	})

	t.Run("rejects BIND addresses specified as non-IPv4 address", func(t *testing.T) {
		_, err := encodeSOCKSReply(socks.Granted, addr.NewHost("localhost", 0))
		require.Error(t, err)
	})
}

func decodeSOCKSReply(b []byte) (*socks.Reply, error) {
	return socks.ReadReply(
		bufio.NewReader(
			bytes.NewReader(b),
		),
	)
}

func encodeSOCKSReply(s socks.Status, bindAddr *addr.Host) ([]byte, error) {
	reply := socks.NewReply(s, bindAddr)

	var buf bytes.Buffer
	if err := reply.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
