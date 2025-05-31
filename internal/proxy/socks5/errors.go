package socks5

import "errors"

var (
	ErrTooManyAuthMethods   = errors.New("too many auth methods")
	ErrHostnameTooLong      = errors.New("hostname too long")
	ErrInvalidVersion       = errors.New("invalid version")
	ErrInvalidAddrType      = errors.New("invalid address type")
	ErrNonZeroReservedField = errors.New("reserved field has a non-zero value")
)
