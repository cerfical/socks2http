package socks

import (
	"fmt"
	"io"

	"github.com/cerfical/socks2http/addr"
)

func WriteConnect(w io.Writer, dest *addr.Addr) error {
	ipv4, err := addr.LookupIPv4(dest.Hostname)
	if err != nil {
		return fmt.Errorf("resolve address %v: %w", dest, err)
	}

	req := Request{Header{
		Version:  V4,
		Command:  RequestConnect,
		DestIP:   ipv4,
		DestPort: dest.Port,
	}, ""}

	return req.Write(w)
}

func WriteGrant(w io.Writer) error {
	rep := Reply{Code: RequestGranted}
	return rep.Write(w)
}

func WriteReject(w io.Writer) error {
	rep := Reply{Code: RequestRejectedOrFailed}
	return rep.Write(w)
}
