package addr

import "strconv"

func ParsePort(port string) (int, error) {
	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, err
	}
	return int(portNum), nil
}
