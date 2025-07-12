package socks

import "fmt"

const (
	StatusGranted              Status = 0x00
	StatusGeneralFailure       Status = 0x01
	StatusConnectionNotAllowed Status = 0x02
	StatusNetworkUnreachable   Status = 0x03
	StatusHostUnreachable      Status = 0x04
	StatusConnectionRefused    Status = 0x05
	StatusTTLExpired           Status = 0x06
	StatusCommandNotSupported  Status = 0x07
	StatusAddrTypeNotSupported Status = 0x08

	StatusNoIdentService Status = iota
	StatusIdentAuthFailure
)

const (
	v4StatusGranted          byte = 0x5a
	v4StatusRejectedOrFailed byte = 0x5b
	v4StatusNoIdentService   byte = 0x5c
	v4StatusIdentAuthFailure byte = 0x5d

	v5StatusFirst byte = byte(StatusGranted)
	v5StatusLast  byte = byte(StatusAddrTypeNotSupported)
)

var (
	statusText = map[Status]string{
		StatusGranted:              "Granted",
		StatusGeneralFailure:       "General Failure",
		StatusConnectionNotAllowed: "Connection Not Allowed",
		StatusNetworkUnreachable:   "Network Unreachable",
		StatusHostUnreachable:      "Host Unreachable",
		StatusConnectionRefused:    "Connection Refused",
		StatusTTLExpired:           "TTL Expired",
		StatusCommandNotSupported:  "Command Not Supported",
		StatusAddrTypeNotSupported: "Address Type Not Supported",

		StatusNoIdentService:   "No Ident Service",
		StatusIdentAuthFailure: "Ident Authentication Failure",
	}

	statusToV4Code = map[Status]byte{
		StatusGranted:          v4StatusGranted,
		StatusNoIdentService:   v4StatusNoIdentService,
		StatusIdentAuthFailure: v4StatusIdentAuthFailure,
	}

	v4CodeToStatus = map[byte]Status{
		v4StatusGranted:          StatusGranted,
		v4StatusRejectedOrFailed: StatusGeneralFailure,
		v4StatusNoIdentService:   StatusNoIdentService,
		v4StatusIdentAuthFailure: StatusIdentAuthFailure,
	}
)

type Status byte

func (s Status) String() string {
	if str, ok := statusText[s]; ok {
		return fmt.Sprintf("%v", str)
	}
	return fmt.Sprintf("%#02x", byte(s))
}

func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func encodeStatus(s Status, v Version) byte {
	if v == V4 {
		if c, ok := statusToV4Code[s]; ok {
			return c
		}
		return v4StatusRejectedOrFailed
	}
	return byte(s)
}

func decodeStatus(b byte) Status {
	if b >= v5StatusFirst && b <= v5StatusLast {
		return Status(b)
	}
	if s, ok := v4CodeToStatus[b]; ok {
		return s
	}
	return Status(b)
}
