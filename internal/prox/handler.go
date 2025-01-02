package prox

import (
	"bufio"
	"io"
	"net"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
)

func newHandler(proto string) (handler, bool) {
	switch proto {
	case addr.HTTP:
		return httpHandler{}, true
	case addr.SOCKS4:
		return socksHandler{}, true
	default:
		return nil, false
	}
}

type handler interface {
	readRequest(r *bufio.Reader) (request, error)
}

type request interface {
	destAddr() *addr.Addr
	logAttrs(log *log.Logger) *log.Logger

	writeGrant(w io.Writer) error
	writeReject(w io.Writer) error
	do(cliConn, servConn net.Conn, proxyProto string) error

	Close() error
}
