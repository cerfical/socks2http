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
	RequestGranted = 90
	ReplyVersion   = 0
)

func TestReadReply(t *testing.T) {
	okTests := map[string]struct {
		input []byte
		want  socks.Reply
	}{
		"request_granted": {
			input: []byte{ReplyVersion, RequestGranted, 0, 0, 0, 0, 0, 0},
			want:  socks.Granted,
		},
	}
	for name, test := range okTests {
		t.Run(name, func(t *testing.T) {
			input := bytes.NewReader(test.input)

			got, err := socks.ReadReply(bufio.NewReader(input))
			require.NoError(t, err)

			assert.Equal(t, test.want, got)
		})
	}

	failTests := map[string]struct {
		input []byte
	}{
		"rejects replies with invalid version code":    {[]byte{1}},
		"rejects replies with unsupported status code": {[]byte{ReplyVersion, 123, 0, 0, 0, 0, 0, 0}},
	}
	for name, test := range failTests {
		t.Run(name, func(t *testing.T) {
			input := bytes.NewReader(test.input)

			_, err := socks.ReadReply(bufio.NewReader(input))
			assert.Error(t, err)
		})
	}
}

func TestReply_String(t *testing.T) {
	tests := map[string]struct {
		reply socks.Reply
		want  string
	}{
		"prints supported replies as reply description followed by reply code in hex": {socks.Granted, "request granted (0x5a)"},
		"prints unsupported replies as reply code in hex":                             {0x17, "(0x17)"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.reply.String()
			assert.Equal(t, test.want, got)
		})
	}
}

func TestReply_Write(t *testing.T) {
	tests := map[string]struct {
		reply socks.Reply
		want  []byte
	}{
		"correctly encodes a SOCKS4 reply": {socks.Granted, []byte{0, RequestGranted, 0, 0, 0, 0, 0, 0}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got bytes.Buffer
			require.NoError(t, test.reply.Write(&got))

			assert.Equal(t, test.want, got.Bytes())
		})
	}
}
