package addr

import (
	"fmt"
	"net"
)

// LookupIPv4 determines [IPv4Addr] corresponding to the hostname.
func LookupIPv4(hostname string) (IPv4Addr, error) {
	// assume localhost, if the hostname is not specified
	if hostname == "" {
		return IPv4Addr{127, 0, 0, 1}, nil
	}

	ip, err := net.ResolveIPAddr("ip4", hostname)
	if err != nil {
		return IPv4Addr{}, err
	}
	return IPv4Addr(ip.IP.To4()), nil
}

type IPv4Addr [4]byte

func (a IPv4Addr) String() string {
	return fmt.Sprintf("%v.%v.%v.%v", a[0], a[1], a[2], a[3])
}
