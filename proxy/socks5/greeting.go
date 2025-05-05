package socks5

import (
	"bufio"
	"fmt"
	"io"
	"math"
)

func ReadGreeting(r *bufio.Reader) (*Greeting, error) {
	if err := checkVersion(r); err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}

	nauth, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("decode auth method count: %w", err)
	}

	authMethods := make([]AuthMethod, 0, nauth)
	for range nauth {
		m, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("decode auth methods: %w", err)
		}
		authMethods = append(authMethods, AuthMethod(m))
	}

	return &Greeting{authMethods}, nil
}

type Greeting struct {
	AuthMethods []AuthMethod
}

func (g *Greeting) Write(w io.Writer) error {
	if len(g.AuthMethods) > math.MaxUint8 {
		return fmt.Errorf("%w (%v)", ErrTooManyAuthMethods, len(g.AuthMethods))
	}

	// Wirte a header with a version code and a number of auth methods
	bytes := make([]byte, 0, 2+len(g.AuthMethods))
	bytes = append(bytes, VersionCode)
	bytes = append(bytes, byte(len(g.AuthMethods)))

	// Write the auth methods
	for _, m := range g.AuthMethods {
		bytes = append(bytes, byte(m))
	}

	_, err := w.Write(bytes)
	return err
}
