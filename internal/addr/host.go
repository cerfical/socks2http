package addr

import "strconv"

type Host struct {
	Hostname string
	Port     uint16
}

func (h Host) String() string {
	var suffix string
	if h.Port != 0 {
		suffix = ":" + strconv.FormatUint(uint64(h.Port), 10)
	}
	return h.Hostname + suffix
}
