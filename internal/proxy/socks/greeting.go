package socks

import (
	"bufio"
	"fmt"
	"io"
	"math"
)

func ReadGreeting(r *bufio.Reader) (*Greeting, error) {
	ver, err := checkVersion(r, V5)
	if err != nil {
		return nil, err
	}

	nauth, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("decode auths count: %w", err)
	}

	greet := Greeting{
		Version: Version(ver),
		Auth:    make([]Auth, 0, nauth),
	}

	for range nauth {
		m, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("decode auths: %w", err)
		}
		greet.Auth = append(greet.Auth, Auth(m))
	}

	return &greet, nil
}

type Greeting struct {
	Version Version
	Auth    []Auth
}

func (g *Greeting) Write(w io.Writer) error {
	// Only SOCKS5 supports greetings
	if g.Version != V5 {
		return badVersion(g.Version)
	}

	if len(g.Auth) > math.MaxUint8 {
		return fmt.Errorf("too many auth methods (%v)", len(g.Auth))
	}

	// Wirte a header with a version code and a number of auth methods
	bytes := make([]byte, 0, 2+len(g.Auth))
	bytes = append(bytes, byte(V5))
	bytes = append(bytes, byte(len(g.Auth)))

	// Write the auth methods
	for _, m := range g.Auth {
		bytes = append(bytes, byte(m))
	}

	_, err := w.Write(bytes)
	return err
}
