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

func (t *ProtoTest) TestParse() {
	t.Run("parses protocol from valid protocol name", func() {
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

func (t *ProtoTest) TestString() {
	t.Run("prints protocol name in upper case", func() {
		got, err := addr.ProtoSOCKS5.MarshalText()
		t.Require().NoError(err)

		t.Equal([]byte("SOCKS5"), got)
	})

	t.Run("panics on invalid protocol", func() {
		t.Panics(func() {
			p := addr.Proto(0)
			p.MarshalText()
		})
	})
}

func (t *ProtoTest) TestTextMarshalUnmarshal() {
	t.Run("marshalling converts protocol to string", func() {
		proto := addr.ProtoSOCKS4

		data, err := proto.MarshalText()
		t.Require().NoError(err)

		t.Equal("SOCKS4", string(data))
	})

	t.Run("unmarshalling parses protocol from string", func() {
		input := "socks4"

		var proto addr.Proto
		err := proto.UnmarshalText([]byte(input))
		t.Require().NoError(err)

		t.Equal(addr.ProtoSOCKS4, proto)
	})
}
