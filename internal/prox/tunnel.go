package prox

import (
	"bufio"
	"errors"
	"io"
	"net"

	"github.com/cerfical/socks2http/internal/log"
)

func tunnel(cliBufr *bufio.Reader, cliConn, servConn net.Conn, log *log.Logger) {
	errChan := make(chan error)
	go transfer(servConn, cliBufr, errChan)
	go transfer(cliConn, servConn, errChan)

	for err := range errChan {
		if err != nil {
			log.Errorf("tunneling: %v", err)
			if err := resetConn(cliConn); err != nil {
				log.Errorf("reset the client connection: %v", err)
			}
			if err := resetConn(servConn); err != nil {
				log.Errorf("reset the server connection: %v", err)
			}
			return
		}
	}
}

func transfer(dest io.Writer, src io.Reader, errChan chan<- error) {
	if _, err := io.Copy(dest, src); !errors.Is(err, net.ErrClosed) {
		errChan <- err
	}
}

func resetConn(conn net.Conn) error {
	return conn.(*net.TCPConn).SetLinger(0)
}
