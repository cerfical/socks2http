package socks_test

import (
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/socks"
	"github.com/stretchr/testify/assert"
)

func TestRequest_Write(t *testing.T) {
	tests := []struct {
		name string
		r    *socks.Request
		want []byte
	}{
		{"valid_request",
			&socks.Request{socks.V4, socks.ConnectCommand, 1080, addr.IPv4Addr{127, 0, 0, 1}, "username"},
			[]byte{socks.V4, socks.ConnectCommand, 4, 56, 127, 0, 0, 1, 'u', 's', 'e', 'r', 'n', 'a', 'm', 'e', 0}},
		{"empty_request",
			&socks.Request{0, 0, 0, addr.IPv4Addr{0, 0, 0, 0}, ""},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0}},
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
