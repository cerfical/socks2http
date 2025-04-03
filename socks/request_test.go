package socks_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/socks"
	"github.com/stretchr/testify/assert"
)

var host = addr.NewHost("127.0.0.1", 1080)

func TestReadRequest(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  *socks.Request
		ok    bool
	}{
		{"connect_no_user", []byte{4, 1, 4, 56, 127, 0, 0, 1, 0}, socks.NewRequest(socks.V4, socks.Connect, host), true},
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
			socks.NewRequest(socks.V4, socks.Connect, host),
			[]byte{socks.V4, socks.Connect, 4, 56, 127, 0, 0, 1, 0}},
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
