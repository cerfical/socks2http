package addr

import "fmt"

type IPv4 [4]byte

func (a IPv4) String() string {
	return fmt.Sprintf("%v.%v.%v.%v", a[0], a[1], a[2], a[3])
}

func (a IPv4) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}
