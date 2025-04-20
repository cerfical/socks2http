package socks

import "fmt"

const (
	Connect Command = 0x01
	Bind    Command = 0x02
)

var commands = map[Command]string{
	Connect: "CONNECT",
	Bind:    "BIND",
}

type Command byte

func (c Command) String() string {
	if str, ok := commands[c]; ok {
		return fmt.Sprintf("%v (%v)", str, hexByte(c))
	}
	return ""
}

func makeCommand(b byte) (c Command, isValid bool) {
	c = Command(b)
	if _, ok := commands[c]; ok {
		return c, true
	}
	return c, false
}
