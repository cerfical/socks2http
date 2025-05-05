package socks5

import "fmt"

const (
	CommandConnect      Command = 0x01
	CommandBind         Command = 0x02
	CommandAssociateUDP Command = 0x03
)

var commands = map[Command]string{
	CommandConnect:      "CONNECT",
	CommandBind:         "BIND",
	CommandAssociateUDP: "ASSOCIATE_UDP",
}

type Command byte

func (c Command) String() string {
	if str, ok := commands[c]; ok {
		return fmt.Sprintf("%v", str)
	}
	return fmt.Sprintf("%v", hexByte(c))
}

func (c Command) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}
