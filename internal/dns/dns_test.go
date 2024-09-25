package dns

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver(t *testing.T) {
	t.Run("etc_resolv_conf", func(t *testing.T) {
		resolver, err := NewResolver([]string{})
		require.NoError(t, err)

		rrs, err := resolver.Resolve(dns.TypeA, "www.baidu.com")
		if err != nil {
			t.Fatal(err)
		}
		require.NoError(t, err)
		assert.NotNil(t, rrs)

		rrs, err = resolver.Resolve(dns.TypeAAAA, "www.baidu.com")
		if err != nil {
			t.Fatal(err)
		}
		require.NoError(t, err)
		assert.NotNil(t, rrs)

		rrs, err = resolver.Resolve(dns.TypeCNAME, "www.baidu.com")
		if err != nil {
			t.Fatal(err)
		}
		require.NoError(t, err)
		assert.NotNil(t, rrs)
	})
	t.Run("dns_server", func(t *testing.T) {
		resolver, err := NewResolver([]string{"114.114.114.114"})
		require.NoError(t, err)

		rrs, err := resolver.Resolve(dns.TypeA, "www.baidu.com")
		if err != nil {
			t.Fatal(err)
		}
		require.NoError(t, err)
		assert.NotNil(t, rrs)

		rrs, err = resolver.Resolve(dns.TypeAAAA, "www.baidu.com")
		if err != nil {
			t.Fatal(err)
		}
		require.NoError(t, err)
		assert.NotNil(t, rrs)

		rrs, err = resolver.Resolve(dns.TypeCNAME, "www.baidu.com")
		if err != nil {
			t.Fatal(err)
		}
		require.NoError(t, err)
		assert.NotNil(t, rrs)
	})
}
