package serv

import (
	"net"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
)

func newRequester(proto string) (requester, bool) {
	switch proto {
	case addr.HTTP:
		return httpRequester{}, true
	case addr.SOCKS4:
		return socksRequester{}, true
	default:
		return nil, false
	}
}

type requester interface {
	request(cliConn net.Conn) (request, error)
}

type request interface {
	writeReply(ok bool) error

	perform(proto string, servConn net.Conn, log *log.Logger)

	destAddr() *addr.Addr
	Close() error
}
