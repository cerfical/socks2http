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
		{"access_rejected", 0, 91, "access rejected"},
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
