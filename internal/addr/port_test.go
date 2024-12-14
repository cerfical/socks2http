package addr_test

import (
	"socks2http/internal/addr"
	"socks2http/internal/test"
	"socks2http/internal/test/checks"
	"testing"
)

func TestIsValidPort(t *testing.T) {
	test := new(test.Test).
		Case("0", true).
		Case("443", true).
		Case("65535", true).
		Case("-1", false).
		Case("00", false).
		Case("65536", false).
		Case("text", false).
		Case("0.0", false).
		Case("0.", false).
		Case(" 0", false).
		Case("0 ", false).
		Case(" ", false).
		Case("", false)
	test.Assert(t, addr.IsValidPort)
}

func TestParsePort(t *testing.T) {
	test := new(test.Test).
		On("0").Want(0, nil).
		On("8080").Want(8080, nil).
		On("65535").Want(65535, nil).
		On("65536").Want(0, checks.NotNil).
		On("-1").Want(0, checks.NotNil).
		On("sss").Want(0, checks.NotNil).
		On("").Want(0, checks.NotNil)
	test.Assert(t, addr.ParsePort)
}
