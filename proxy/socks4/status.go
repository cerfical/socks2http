package socks4

import "fmt"

const (
	StatusGranted          Status = 0x5a
	StatusRejectedOrFailed Status = 0x5b
	StatusNoAuthService    Status = 0x5c
	StatusAuthFailed       Status = 0x5d
)

var statuses = map[Status]string{
	StatusGranted:          "Request Granted",
	StatusRejectedOrFailed: "Request Rejected or Failed",
	StatusNoAuthService:    "No Authentication Service",
	StatusAuthFailed:       "Authentication Failed",
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
