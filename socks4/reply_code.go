package socks4

import "fmt"

const (
	Granted    ReplyCode = 0x5a
	Rejected   ReplyCode = 0x5b
	NoAuth     ReplyCode = 0x5c
	AuthFailed ReplyCode = 0x5d
)

var replyCodes = map[ReplyCode]string{
	Granted:    "Request Granted",
	Rejected:   "Request Rejected",
	NoAuth:     "No Authentication Service",
	AuthFailed: "Authentication Failed",
}

type ReplyCode byte

func (c ReplyCode) String() string {
	if str, ok := replyCodes[c]; ok {
		return fmt.Sprintf("%v (%v)", str, hexByte(c))
	}
	return ""
}

func makeReplyCode(b byte) (c ReplyCode, isValid bool) {
	c = ReplyCode(b)
	if _, ok := replyCodes[c]; ok {
		return c, true
	}
	return c, false
}
