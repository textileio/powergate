package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ory/dockertest"
	"github.com/textileio/lotus-client/api/apistruct"
	"github.com/textileio/powergate/lotus"
	"github.com/textileio/powergate/util"
)

func CreateLocalDevnet(t *testing.T, numMiners int) (*apistruct.FullNodeStruct, address.Address, []address.Address) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(fmt.Sprintf("couldn't create ipfs-pool: %s", err))
	}
	envNumMiners := fmt.Sprintf("TEXLOTUSDEVNET_NUMMINERS=%d", numMiners)
	lotusDevnet, err := pool.RunWithOptions(&dockertest.RunOptions{Repository: "textile/lotus-devnet", Tag: "lotus-testnet3-d41aeb", Env: []string{envNumMiners}, Mounts: []string{"/tmp/powergate:/tmp/powergate"}})
	if err != nil {
		panic(fmt.Sprintf("couldn't run lotus-devnet container: %s", err))
	}
	lotusDevnet.Expire(180)
	time.Sleep(time.Second * 3)
	t.Cleanup(func() {
		if err := pool.Purge(lotusDevnet); err != nil {
			panic(fmt.Sprintf("couldn't purge lotus-devnet from docker pool: %s", err))
		}
	})

	c, cls, err := lotus.New(util.MustParseAddr("/ip4/127.0.0.1/tcp/"+lotusDevnet.GetPort("7777/tcp")), "")
	if err != nil {
		panic(err)
	}
	t.Cleanup(func() { cls() })
	ctx := context.Background()
	addr, err := c.WalletDefaultAddress(ctx)
	if err != nil {
		t.Fatal(err)
	}

	miners, err := c.StateListMiners(ctx, types.EmptyTSK)
	if err != nil {
		t.Fatal(err)
	}

	return c, addr, miners
}
