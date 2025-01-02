package prox

import (
	"bufio"
	"io"

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
	parseRequest(r *bufio.Reader) (request, error)
}

type request interface {
	isConnect() bool
	destAddr() *addr.Addr

	logAttrs(log *log.Logger) *log.Logger
	writeGrant(w io.Writer) error
	writeReject(w io.Writer) error

	write(w io.Writer) error
	writeProxy(w io.Writer) error

	Close() error
}
