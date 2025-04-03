package addr

import (
	"fmt"
	"net"
)

// LookupIPv4 determines [IPv4] corresponding to the hostname.
func LookupIPv4(hostname string) (IPv4, error) {
	// assume localhost, if the hostname is not specified
	if hostname == "" {
		return IPv4{127, 0, 0, 1}, nil
	}

	ip, err := net.ResolveIPAddr("ip4", hostname)
	if err != nil {
		return IPv4{}, err
	}
	return IPv4(ip.IP.To4()), nil
}

type IPv4 [4]byte

func (a IPv4) String() string {
	return fmt.Sprintf("%v.%v.%v.%v", a[0], a[1], a[2], a[3])
}
