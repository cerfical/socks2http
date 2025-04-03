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
	SOCKSVersion4  = 0x04
	ConnectCommand = 0x01
)

var host = addr.NewHost("127.0.0.1", 1080)

func TestReadRequest(t *testing.T) {
	okTests := map[string]struct {
		input []byte
		want  *socks.Request
	}{
		"correctly decodes a SOCKS4 request": {
			input: []byte{SOCKSVersion4, ConnectCommand, 0x04, 0x38, 127, 0, 0, 1, 0},
			want:  socks.NewRequest(socks.V4, socks.Connect, host),
		},
	}
	for name, test := range okTests {
		t.Run(name, func(t *testing.T) {
			r := bufio.NewReader(bytes.NewReader(test.input))

			got, err := socks.ReadRequest(r)
			require.NoError(t, err)

			assert.Equal(t, test.want, got)
		})
	}

	failTests := map[string]struct {
		input []byte
		err   error
	}{
		"rejects unsupported SOCKS versions": {
			input: []byte{123},
			err:   socks.ErrInvalidVersion,
		},

		"rejects unsupported SOCKS4 commands": {
			input: []byte{SOCKSVersion4, 123, 0x04, 0x38, 127, 0, 0, 1, 0},
			err:   socks.ErrInvalidCommand,
		},
	}
	for name, test := range failTests {
		t.Run(name, func(t *testing.T) {
			r := bufio.NewReader(bytes.NewReader(test.input))

			_, err := socks.ReadRequest(r)
			require.ErrorIs(t, err, test.err)
		})
	}
}

func TestRequest_Write(t *testing.T) {
	tests := map[string]struct {
		req  *socks.Request
		want []byte
	}{
		"correctly encodes a SOCKS4 request": {
			req:  socks.NewRequest(socks.V4, socks.Connect, host),
			want: []byte{SOCKSVersion4, ConnectCommand, 0x04, 0x38, 127, 0, 0, 1, 0},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got bytes.Buffer
			require.NoError(t, test.req.Write(&got))

			assert.Equal(t, test.want, got.Bytes())
		})
	}
}
