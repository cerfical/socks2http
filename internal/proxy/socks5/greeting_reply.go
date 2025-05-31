package socks5

import (
	"bufio"
	"fmt"
	"io"
)

func ReadGreetingReply(r *bufio.Reader) (*GreetingReply, error) {
	if err := checkVersion(r); err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}

	cauth, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("decode auth method: %w", err)
	}

	return &GreetingReply{AuthMethod(cauth)}, nil
}

type GreetingReply struct {
	AuthMethod AuthMethod
}

func (r *GreetingReply) Write(w io.Writer) error {
	bytes := []byte{VersionCode, byte(r.AuthMethod)}
	_, err := w.Write(bytes)
	return err
}
