package maxmind

import (
	"testing"

	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	mm := setup(t)

	tests := []struct {
		Name       string
		Maddr      multiaddr.Multiaddr
		CountryISO string
		Error      bool
	}{
		{
			Name:       "China IP4",
			Maddr:      multiaddr.StringCast("/ip4/59.110.242.123/tcp/1234"),
			CountryISO: "CN",
		},
		{
			Name:       "Uruguay IP4",
			Maddr:      multiaddr.StringCast("/ip4/186.52.85.100/tcp/1234"),
			CountryISO: "UY",
		},
		{
			Name:       "DNS",
			Maddr:      multiaddr.StringCast("/dns4/google.com/tcp/80"),
			CountryISO: "US",
		},
		{
			Name:       "USA IP6",
			Maddr:      multiaddr.StringCast("/ip6/2001:4b0:85a3:0000:0000:8a2e:0370:7334/tcp/1234"),
			CountryISO: "US",
		},
		{
			Name:  "Unresolvable IP4",
			Maddr: multiaddr.StringCast("/ip4/127.0.233.233/tcp/1234"),
			Error: true,
		},
	}
	for _, tc := range tests {
		l, err := mm.Resolve([]multiaddr.Multiaddr{tc.Maddr})
		require.Equal(t, tc.Error, err != nil)
		if tc.Error {
			continue
		}
		require.Equal(t, tc.CountryISO, l.Country)
		require.NotZero(t, l.Latitude)
		require.NotZero(t, l.Longitude)
	}
}

func setup(t *testing.T) *MaxMind {
	mm, err := New("GeoLite2-City.mmdb")
	require.NoError(t, err)
	t.Cleanup(func() { _ = mm.Close() })
	return mm
}
