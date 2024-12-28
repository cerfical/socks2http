package socks_test

import (
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/internal/socks"
	"github.com/stretchr/testify/assert"
)

func TestReadReply(t *testing.T) {
	tests := []struct {
		name      string
		ver, code byte
		wantErr   string
	}{
		{"access_granted", 0, 90, ""},
		{"access_denied", 0, 91, "access denied"},
		{"no_identd", 0, 92, "identd service"},
		{"auth_failed", 0, 93, "authentication failed"},
		{"invalid_reply", 0, 94, "unexpected reply code"},
		{"invalid_version", 1, 90, "unexpected version number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := bytes.NewReader([]byte{tt.ver, tt.code, 0, 0, 0, 0, 0, 0})
			err := socks.ReadReply(input)

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestReply_Write(t *testing.T) {
	tests := []struct {
		name string
		r    *socks.Reply
		want []byte
	}{
		{"access_granted", &socks.Reply{Code: 90}, []byte{0, socks.AccessGranted, 0, 0, 0, 0, 0, 0}},
		{"access_denied", &socks.Reply{Code: 91}, []byte{0, socks.AccessDenied, 0, 0, 0, 0, 0, 0}},
		{"no_identd", &socks.Reply{Code: 92}, []byte{0, socks.NoIdentd, 0, 0, 0, 0, 0, 0}},
		{"auth_failed", &socks.Reply{Code: 93}, []byte{0, socks.AuthFailed, 0, 0, 0, 0, 0, 0}},
	}

	buf := bytes.Buffer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer buf.Reset()

			err := tt.r.Write(&buf)
			got := buf.Bytes()

			assert.Equal(t, tt.want, got)
			assert.NoError(t, err)
		})
	}
}
