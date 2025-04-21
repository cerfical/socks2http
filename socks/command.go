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

func decodeCommand(b byte) (Command, bool) {
	if _, ok := commands[Command(b)]; ok {
		return Command(b), true
	}
	return 0, false
}

func encodeCommand(c Command) (byte, bool) {
	if _, ok := commands[c]; ok {
		return byte(c), true
	}
	return 0, false
}

type Command byte

func (c Command) String() string {
	if str, ok := commands[c]; ok {
		return fmt.Sprintf("%v (%v)", str, hexByte(c))
	}
	return ""
}
