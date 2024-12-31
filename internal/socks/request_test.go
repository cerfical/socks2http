package socks_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/socks"
	"github.com/stretchr/testify/assert"
)

func TestReadRequest(t *testing.T) {
	h := socks.Header{socks.V4, socks.RequestConnect, 1080, addr.IPv4Addr{127, 0, 0, 1}}
	tests := []struct {
		name  string
		input []byte
		want  *socks.Request
		ok    bool
	}{
		{"connect_no_user", []byte{4, 1, 4, 56, 127, 0, 0, 1, 0}, &socks.Request{h, ""}, true},
		{"connect_with_user", []byte{4, 1, 4, 56, 127, 0, 0, 1, 'u', 's', 'e', 'r', 0}, &socks.Request{h, "user"}, true},
		{"no_version_3", []byte{3, 1, 4, 56, 127, 0, 0, 1, 0}, nil, false},
		{"no_version_5", []byte{5, 1, 4, 56, 127, 0, 0, 1, 0}, nil, false},
		{"no_command_0", []byte{4, 0, 4, 56, 127, 0, 0, 1, 0}, nil, false},
		{"no_command_2", []byte{4, 2, 4, 56, 127, 0, 0, 1, 0}, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := socks.ReadRequest(bufio.NewReader(bytes.NewReader(tt.input)))

			assert.Equal(t, tt.want, got)
			if tt.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
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
			&socks.Request{socks.Header{socks.V4, socks.RequestConnect, 1080, addr.IPv4Addr{127, 0, 0, 1}}, "username"},
			[]byte{socks.V4, socks.RequestConnect, 4, 56, 127, 0, 0, 1, 'u', 's', 'e', 'r', 'n', 'a', 'm', 'e', 0}},
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
