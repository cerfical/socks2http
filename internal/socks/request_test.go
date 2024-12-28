package socks_test

import (
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/socks"
	"github.com/stretchr/testify/assert"
)

func TestReadRequest(t *testing.T) {
	h := socks.Header{socks.V4, socks.ConnectCommand, 1080, addr.IPv4Addr{127, 0, 0, 1}}
	tests := []struct {
		name    string
		input   []byte
		want    *socks.Request
		wantErr string
	}{
		{"connect_no_user", []byte{4, 1, 4, 56, 127, 0, 0, 1, 0}, &socks.Request{h, ""}, ""},
		{"connect_with_user", []byte{4, 1, 4, 56, 127, 0, 0, 1, 'u', 's', 'e', 'r', 0}, &socks.Request{h, "user"}, ""},
		{"no_version_3", []byte{3, 1, 4, 56, 127, 0, 0, 1, 0}, nil, "invalid version number"},
		{"no_version_5", []byte{5, 1, 4, 56, 127, 0, 0, 1, 0}, nil, "invalid version number"},
		{"no_command_0", []byte{4, 0, 4, 56, 127, 0, 0, 1, 0}, nil, "invalid command code"},
		{"no_command_2", []byte{4, 2, 4, 56, 127, 0, 0, 1, 0}, nil, "invalid command code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := socks.ReadRequest(bytes.NewReader(tt.input))

			assert.Equal(t, tt.want, got)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequest_Write(t *testing.T) {
	tests := []struct {
		name string
		r    *socks.Request
		want []byte
	}{
		{"valid_request",
			&socks.Request{socks.Header{socks.V4, socks.ConnectCommand, 1080, addr.IPv4Addr{127, 0, 0, 1}}, "username"},
			[]byte{socks.V4, socks.ConnectCommand, 4, 56, 127, 0, 0, 1, 'u', 's', 'e', 'r', 'n', 'a', 'm', 'e', 0}},
		{"empty_request",
			&socks.Request{},
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
