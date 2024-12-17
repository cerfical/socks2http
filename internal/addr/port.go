package addr

import "strconv"

// ParsePort converts a string to a valid port number.
func ParsePort(port string) (uint16, error) {
	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(portNum), nil
}

func isValidPort(port string) bool {
	_, err := ParsePort(port)
	return err == nil
}
