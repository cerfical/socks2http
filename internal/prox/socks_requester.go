package prox

import (
	"bufio"
	"net"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/socks"
)

type socksRequester struct{}

func (socksRequester) request(cliConn net.Conn) (request, error) {
	cliRead := bufio.NewReaderSize(cliConn, 17)
	req, err := socks.ReadRequest(cliRead)
	if err != nil {
		return nil, err
	}

	return &socksRequest{addr.Addr{
		Hostname: req.DestIP.String(),
		Port:     req.DestPort,
	}, cliRead, cliConn, req}, nil
}

type socksRequest struct {
	dest    addr.Addr
	cliBufr *bufio.Reader
	cliConn net.Conn
	*socks.Request
}

func (r *socksRequest) writeReply(ok bool) error {
	rep := socks.Reply{}
	if ok {
		rep.Code = socks.RequestGranted
	} else {
		rep.Code = socks.RequestRejectedOrFailed
	}
	return rep.Write(r.cliConn)
}

func (r *socksRequest) perform(_ string, servConn net.Conn, log *log.Logger) {
	log.WithAttrs(
		"command", "CONNECT",
		"host", r.destAddr().Host(),
	).Infof("incoming request")
	tunnel(r.cliBufr, r.cliConn, servConn, log)
}

func (r *socksRequest) destAddr() *addr.Addr {
	return &r.dest
}

func (r *socksRequest) Close() error {
	return nil
}
