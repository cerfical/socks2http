package addr

import (
	"fmt"
	"strconv"
)

type Host struct {
	Hostname string
	Port     uint16
}

func (h Host) String() string {
	return h.Hostname + ":" + strconv.FormatUint(uint64(h.Port), 10)
}

type Addr struct {
	Scheme ProtoScheme
	Host
}

func (a Addr) String() string {
	return fmt.Sprintf("%v://%v", a.Scheme, a.Host)
}
