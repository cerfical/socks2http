package addr

import "strconv"

func ParsePort(port string) (uint16, error) {
	if port == "" {
		return 0, nil
	}

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
