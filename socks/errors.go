package socks

import "errors"

var ErrInvalidVersion = errors.New("invalid version")
var ErrInvalidCommand = errors.New("invalid command")
var ErrInvalidReply = errors.New("invalid reply")
