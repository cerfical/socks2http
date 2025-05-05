package socks4

import "fmt"

const (
	CommandConnect Command = 0x01
	CommandBind    Command = 0x02
)

var commands = map[Command]string{
	CommandConnect: "CONNECT",
	CommandBind:    "BIND",
}

type Command byte

func (c Command) String() string {
	if str, ok := commands[c]; ok {
		return fmt.Sprintf("%v (%v)", str, hexByte(c))
	}
	return fmt.Sprintf("Invalid Command (%v)", hexByte(c))
}

func (c Command) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}
