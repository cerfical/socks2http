package socks

import "fmt"

const (
	Connect Command = 0x01
)

var commands = map[Command]string{
	Connect: "CONNECT",
}

func makeCommand(b byte) (c Command, isValid bool) {
	c = Command(b)
	if _, ok := commands[c]; ok {
		return c, true
	}
	return c, false
}

type Command byte

func (c Command) String() string {
	code := fmt.Sprintf("(%#02x)", byte(c))
	if s, ok := commands[c]; ok {
		return fmt.Sprintf("%v %v", s, code)
	}
	return code
}
