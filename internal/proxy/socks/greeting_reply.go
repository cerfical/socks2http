package socks

import (
	"bufio"
	"fmt"
	"io"
)

func ReadGreetingReply(r *bufio.Reader) (*GreetingReply, error) {
	ver, err := checkVersion(r, V5)
	if err != nil {
		return nil, err
	}
	greetRep := GreetingReply{Version: ver}

	auth, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("decode auth: %w", err)
	}
	greetRep.Auth = Auth(auth)

	return &greetRep, nil
}

type GreetingReply struct {
	Version Version
	Auth    Auth
}

func (r *GreetingReply) Write(w io.Writer) error {
	// Only SOCKS5 supports greeting replies
	if r.Version != V5 {
		return badVersion(r.Version)
	}

	bytes := []byte{byte(r.Version), byte(r.Auth)}

	_, err := w.Write(bytes)
	return err
}
