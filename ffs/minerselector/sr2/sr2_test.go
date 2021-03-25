package sr2

import (
	"context"
	"fmt"
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/lotus"
)

// TestMS is meant to be runned locally since it needs a fully
// synced Lotus node.
func TestMS(t *testing.T) {
	t.SkipNow()
	lotusToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.4KpuySIvV4n6kBEXQOle-hi1Ec3lyUmRYCknz4NQyLM"

	cb, err := lotus.NewBuilder("127.0.0.1:5555", lotusToken, 1)
	require.NoError(t, err)

	url := "https://raw.githubusercontent.com/filecoin-project/slingshot/master/miners.json"
	sr2, err := New(url, cb)
	require.NoError(t, err)

	for {
		_, err := sr2.GetMiners(1, ffs.MinerSelectorFilter{})
		require.NoError(t, err)
	}
}

func TestCustom(t *testing.T) {
	t.SkipNow()
	lotusToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.4KpuySIvV4n6kBEXQOle-hi1Ec3lyUmRYCknz4NQyLM"

	cb, err := lotus.NewBuilder("127.0.0.1:5555", lotusToken, 1)
	require.NoError(t, err)

	c, cls, err := cb(context.Background())
	require.NoError(t, err)
	defer cls()

	addr, err := address.NewFromString("t3rvsbv5yj5lil74o33bfn5mjsdlnnogukgqua5cnumtid3kgibqeer2uaipcm57iil2ndzykzq34ebp2xajwq")
	require.NoError(t, err)
	b, err := c.WalletBalance(context.Background(), addr)
	require.NoError(t, err)
	fmt.Printf("Balance: %d\n", b)
}
