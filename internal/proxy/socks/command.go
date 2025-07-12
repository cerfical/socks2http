package socks

import "fmt"

const (
	CommandConnect   Command = 0x01
	CommandBind      Command = 0x02
	CommandAssociate Command = 0x03
)

var commandNames = map[Command]string{
	CommandConnect:   "CONNECT",
	CommandBind:      "BIND",
	CommandAssociate: "ASSOCIATE",
}

type Command byte

func (c Command) String() string {
	if str, ok := commandNames[c]; ok {
		return fmt.Sprintf("%v", str)
	}
	return fmt.Sprintf("%#02x", byte(c))
}

func (c Command) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}
