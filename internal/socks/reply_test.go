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
		ok        bool
	}{
		{"request_granted", 0, 90, true},
		{"rejected_or_failed", 0, 91, false},
		{"rejected_no_auth", 0, 92, false},
		{"rejected_auth_failed", 0, 93, false},
		{"invalid_reply", 0, 94, false},
		{"invalid_version", 1, 90, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := bytes.NewReader([]byte{tt.ver, tt.code, 0, 0, 0, 0, 0, 0})
			err := socks.ReadReply(input)

			if tt.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
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
		{"request_granted", &socks.Reply{Code: 90}, []byte{0, socks.RequestGranted, 0, 0, 0, 0, 0, 0}},
		{"rejected_or_failed", &socks.Reply{Code: 91}, []byte{0, socks.RequestRejectedOrFailed, 0, 0, 0, 0, 0, 0}},
		{"rejected_no_auth", &socks.Reply{Code: 92}, []byte{0, socks.RequestRejectedNoAuth, 0, 0, 0, 0, 0, 0}},
		{"rejected_auth_failed", &socks.Reply{Code: 93}, []byte{0, socks.RequestRejectedAuthFailed, 0, 0, 0, 0, 0, 0}},
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
