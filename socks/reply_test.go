package socks_test

import (
	"bufio"
	"bytes"
	"testing"

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

func TestReadReply(t *testing.T) {
	okTests := map[string]struct {
		input byte
		want  socks.Status
	}{
		"decodes a SOCKS4 GRANTED status": {
			input: RequestGranted,
			want:  socks.Granted,
		},

		"decodes a SOCKS4 REJECTED status": {
			input: RequestRejectedOrFailed,
			want:  socks.Rejected,
		},

		"decodes a SOCKS4 NO-AUTH status": {
			input: RequestRejectedNoAuth,
			want:  socks.NoAuth,
		},

		"decodes a SOCKS4 AUTH-FAILED status": {
			input: RequestRejectedAuthFailed,
			want:  socks.AuthFailed,
		},
	}
	for name, test := range okTests {
		t.Run(name, func(t *testing.T) {
			input := bytes.NewReader([]byte{ReplyVersion, test.input, 0, 0, 0, 0, 0, 0})

			got, err := socks.ReadReply(bufio.NewReader(input))
			require.NoError(t, err)

			assert.Equal(t, test.want, got.Status)
		})
	}

	failTests := map[string]struct {
		input []byte
	}{
		"rejects replies with unsupported reply version": {
			input: []byte{1},
		},

		"rejects replies with unsupported status": {
			input: []byte{ReplyVersion, 0x5e, 0, 0, 0, 0, 0, 0},
		},
	}
	for name, test := range failTests {
		t.Run(name, func(t *testing.T) {
			input := bytes.NewReader(test.input)

			_, err := socks.ReadReply(bufio.NewReader(input))
			assert.Error(t, err)
		})
	}
}

func TestReply_Write(t *testing.T) {
	tests := map[string]struct {
		input socks.Status
		want  byte
	}{
		"encodes a SOCKS4 GRANTED status": {
			input: socks.Granted,
			want:  RequestGranted,
		},

		"encodes a SOCKS4 REJECTED status": {
			input: socks.Rejected,
			want:  RequestRejectedOrFailed,
		},

		"encodes a SOCKS4 NO-AUTH status": {
			input: socks.NoAuth,
			want:  RequestRejectedNoAuth,
		},

		"encodes a SOCKS4 AUTH-FAILED status": {
			input: socks.AuthFailed,
			want:  RequestRejectedAuthFailed,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			want := []byte{0, test.want, 0, 0, 0, 0, 0, 0}
			reply := socks.NewReply(test.input)

			var got bytes.Buffer
			require.NoError(t, reply.Write(&got))

			assert.Equal(t, want, got.Bytes())
		})
	}
}
