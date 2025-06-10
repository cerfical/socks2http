package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/stretchr/testify/suite"
)

func TestProto(t *testing.T) {
	suite.Run(t, new(ProtoTest))
}

type ProtoTest struct {
	suite.Suite
}

func (t *ProtoTest) TestMarshalText() {
	t.Run("encodes a valid protocol name", func() {
		got, err := addr.ProtoSOCKS5.MarshalText()
		t.Require().NoError(err)

		t.Equal([]byte("socks5"), got)
	})

	t.Run("panics on an invalid protocol", func() {
		t.Panics(func() {
			p := addr.Proto(0)
			p.MarshalText()
		})
	})
}

func (t *ProtoTest) TestUnmarshalText() {
	t.Run("decodes a valid protocol name", func() {
		var got addr.Proto
		t.Require().NoError(got.UnmarshalText([]byte("socks5")))

		t.Equal(addr.ProtoSOCKS5, got)
	})

	t.Run("ignores character case", func() {
		var got addr.Proto
		t.Require().NoError(got.UnmarshalText([]byte("Socks5")))

		t.Equal(addr.ProtoSOCKS5, got)
	})

	t.Run("rejects invalid protocol names", func() {
		var got addr.Proto
		t.Require().Error(got.UnmarshalText([]byte("socks6")))
	})
}
