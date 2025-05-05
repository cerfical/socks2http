package socks5

import "fmt"

const (
	StatusOK                   Status = 0x00
	StatusGeneralFailure       Status = 0x01
	StatusConnectionNotAllowed Status = 0x02
	StatusNetworkUnreachable   Status = 0x03
	StatusHostUnreachable      Status = 0x04
	StatusConnectionRefused    Status = 0x05
	StatusTTLExpired           Status = 0x06
	StatusCommandNotSupported  Status = 0x07
	StatusAddrTypeNotSupported Status = 0x08
)

var statuses = map[Status]string{
	StatusOK:                   "Succeeded",
	StatusGeneralFailure:       "General Failure",
	StatusConnectionNotAllowed: "Connection Not Allowed",
	StatusNetworkUnreachable:   "Network Unreachable",
	StatusHostUnreachable:      "Host Unreachable",
	StatusConnectionRefused:    "Connection Refused",
	StatusTTLExpired:           "TTL Expired",
	StatusCommandNotSupported:  "Command Not Supported",
	StatusAddrTypeNotSupported: "Address Type Not Supported",
}

type Status byte

func (s Status) String() string {
	if str, ok := statuses[s]; ok {
		return fmt.Sprintf("%v", str)
	}
	return fmt.Sprintf("%v", hexByte(s))
}

func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}
