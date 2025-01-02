package prox

import (
	"errors"
	"io"
	"net"
)

func tunnel(cliConn, servConn net.Conn) error {
	errChan := make(chan error)
	go transfer(servConn, cliConn, errChan)
	go transfer(cliConn, servConn, errChan)

	for err := range errChan {
		if err != nil {
			resetConn(cliConn)
			resetConn(servConn)
			return err
		}
	}

	return nil
}

func transfer(dest io.Writer, src io.Reader, errChan chan<- error) {
	if _, err := io.Copy(dest, src); !errors.Is(err, net.ErrClosed) {
		errChan <- err
	}
}

func resetConn(conn net.Conn) {
	_ = conn.(*net.TCPConn).SetLinger(0)
}
