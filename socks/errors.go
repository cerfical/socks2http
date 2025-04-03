package socks

import "errors"

var ErrUnsupportedVersion = errors.New("unsupported version")
var ErrUnsupportedCommand = errors.New("unsupported command")
