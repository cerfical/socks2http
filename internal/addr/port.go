package addr

import "strconv"

func IsValidPort(port string) bool {
	_, err := ParsePort(port)
	return err == nil
}

func ParsePort(port string) (uint16, error) {
	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(portNum), nil
}
