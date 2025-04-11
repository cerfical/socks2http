package socks

import "fmt"

const (
	Granted    Status = 0x5a
	Rejected   Status = 0x5b
	NoAuth     Status = 0x5c
	AuthFailed Status = 0x5d
)

var statuses = map[Status]string{
	Granted:    "request granted",
	Rejected:   "request rejected or failed",
	NoAuth:     "request rejected because SOCKS server cannot connect to identd on the client",
	AuthFailed: "request rejected because the client program and identd report different user-ids",
}

type Status byte

func (s Status) String() string {
	code := printByte(byte(s))
	if s, ok := statuses[s]; ok {
		return fmt.Sprintf("%v %v", s, code)
	}
	return code
}

func makeStatus(b byte) (s Status, isValid bool) {
	s = Status(b)
	if _, ok := statuses[s]; ok {
		return s, true
	}
	return s, false
}
