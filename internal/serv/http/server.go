package http

import (
	"net/http"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/prox"
)

func Run(servAddr *addr.Addr, proxy *prox.Proxy, logger *log.Logger) error {
	s := server{proxy, logger}
	if err := http.ListenAndServe(servAddr.Host(), &s); err != nil {
		return err
	}
	return nil
}

type server struct {
	proxy  *prox.Proxy
	logger *log.Logger
}

func (s *server) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	logger := s.logger.With().
		Str("uri", req.RequestURI).
		Logger()

	logger.With().
		Str("method", req.Method).
		Str("proto", req.Proto).Logger().
		Infof("new request")

	clientConn, _, err := wr.(http.Hijacker).Hijack()
	if err != nil {
		logger.Errorf("opening a client connection: %v", err)
		return
	}

	defer func() {
		if err := clientConn.Close(); err != nil {
			logger.Errorf("closing client connection: %v", err)
		}
	}()

	reqHandler := requestHandler{
		clientConn: clientConn,
		proxy:      s.proxy,
		logger:     logger,
		request:    req,
	}
	reqHandler.run()
}
