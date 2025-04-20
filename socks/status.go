package socks

import "fmt"

const (
	Granted    Status = 0x5a
	Rejected   Status = 0x5b
	NoAuth     Status = 0x5c
	AuthFailed Status = 0x5d
)

var statuses = map[Status]string{
	Granted:    "Request Granted",
	Rejected:   "Request Rejected",
	NoAuth:     "No Authentication Service",
	AuthFailed: "Authentication Failed",
}

type Status byte

func (s Status) String() string {
	if str, ok := statuses[s]; ok {
		return fmt.Sprintf("%v (%v)", str, hexByte(s))
	}
	return ""
}

func makeStatus(b byte) (s Status, isValid bool) {
	s = Status(b)
	if _, ok := statuses[s]; ok {
		return s, true
	}
	return s, false
}
