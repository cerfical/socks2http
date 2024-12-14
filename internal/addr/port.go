package addr

import (
	"errors"
	"strconv"
	"strings"
)

func IsValidPort(port string) bool {
	_, err := ParsePort(port)
	return err == nil
}

func ParsePort(port string) (uint16, error) {
	if strings.HasPrefix(port, "0") && port != "0" {
		return 0, errors.New("port number must not start with zeros")
	}

	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(portNum), nil
}
