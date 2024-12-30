package serv

import (
	"net"
	"strconv"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/socks"
)

type socksRequester struct{}

func (socksRequester) request(cliConn net.Conn) (request, error) {
	req, err := socks.ReadRequest(cliConn)
	if err != nil {
		return nil, err
	}

	return &socksRequest{addr.Addr{
		Hostname: req.DestIP.String(),
		Port:     req.DestPort,
	}, cliConn, req}, nil
}

type socksRequest struct {
	dest    addr.Addr
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

	for err := range tunnel(r.cliConn, servConn) {
		log.Errorf("%v", err)
	}
}

func (r *socksRequest) destAddr() *addr.Addr {
	return &r.dest
}

func (r *socksRequest) Close() error {
	return nil
}
