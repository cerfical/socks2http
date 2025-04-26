package socks4_test

import (
	"bufio"
	"bytes"
	"slices"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/socks4"
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
		reply byte
		want  socks4.ReplyCode
	}{
		"decodes a Request-Granted reply": {
			reply: RequestGranted,
			want:  socks4.Granted,
		},

		"decodes a Request-Rejected reply": {
			reply: RequestRejectedOrFailed,
			want:  socks4.Rejected,
		},

		"decodes a No-Auth reply": {
			reply: RequestRejectedNoAuth,
			want:  socks4.NoAuth,
		},

		"decodes an Auth-Failed reply": {
			reply: RequestRejectedAuthFailed,
			want:  socks4.AuthFailed,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := decodeSOCKSReply([]byte{ReplyVersion, test.reply, 0, 0, 0, 0, 0, 0})
			require.NoError(t, err)

			assert.Equal(t, test.want, got.Code)
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

	t.Run("rejects replies with unsupported reply codes", func(t *testing.T) {
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
		reply socks4.ReplyCode
		want  byte
	}{
		"encodes a Request-Granted reply": {
			reply: socks4.Granted,
			want:  RequestGranted,
		},

		"encodes a Request-Rejected reply": {
			reply: socks4.Rejected,
			want:  RequestRejectedOrFailed,
		},

		"encodes a No-Auth reply": {
			reply: socks4.NoAuth,
			want:  RequestRejectedNoAuth,
		},

		"encodes an Auth-Failed reply": {
			reply: socks4.AuthFailed,
			want:  RequestRejectedAuthFailed,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := encodeSOCKSReply(test.reply, nil)
			require.NoError(t, err)

			want := []byte{ReplyVersion, test.want, 0, 0, 0, 0, 0, 0}
			assert.Equal(t, want, got)
		})
	}

	t.Run("encodes a non-empty BIND address", func(t *testing.T) {
		got, err := encodeSOCKSReply(socks4.Granted, addr.NewHost("127.0.0.1", 1080))
		require.NoError(t, err)

		want := []byte{ReplyVersion, RequestGranted, 0x04, 0x38, 127, 0, 0, 1}
		assert.Equal(t, want, got)
	})

	t.Run("ignores an empty BIND address", func(t *testing.T) {
		got, err := encodeSOCKSReply(socks4.Granted, nil)
		require.NoError(t, err)

		want := []byte{ReplyVersion, RequestGranted, 0, 0, 0, 0, 0, 0}
		assert.Equal(t, want, got)
	})

	t.Run("rejects BIND addresses specified as non-IPv4 address", func(t *testing.T) {
		_, err := encodeSOCKSReply(socks4.Granted, addr.NewHost("localhost", 0))
		require.Error(t, err)
	})
}

func decodeSOCKSReply(b []byte) (*socks4.Reply, error) {
	return socks4.ReadReply(
		bufio.NewReader(
			bytes.NewReader(b),
		),
	)
}

func encodeSOCKSReply(s socks4.ReplyCode, bindAddr *addr.Host) ([]byte, error) {
	reply := socks4.NewReply(s, bindAddr)

	var buf bytes.Buffer
	if err := reply.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
