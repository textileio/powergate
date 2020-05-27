package tests

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/textileio/powergate/lotus"
	"github.com/textileio/powergate/util"
)

// LaunchDevnetDocker launches the devnet docker image
func LaunchDevnetDocker(t *testing.T, numMiners int, ipfsMaddr string) *dockertest.Resource {
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(fmt.Sprintf("couldn't create ipfs-pool: %s", err))
	}
	envs := []string{
		devnetEnv("NUMMINERS", strconv.Itoa(numMiners)),
		devnetEnv("SPEED", "500"),
		devnetEnv("IPFSADDR", ipfsMaddr),
	}
	repository := "textile/lotus-devnet"
	tag := "sha-77f20c7
	lotusDevnet, err := pool.RunWithOptions(&dockertest.RunOptions{Repository: repository, Tag: tag, Env: envs, Mounts: []string{"/tmp/powergate:/tmp/powergate"}})
	if err != nil {
		panic(fmt.Sprintf("couldn't run lotus-devnet container: %s", err))
	}
	if err := lotusDevnet.Expire(180); err != nil {
		panic(err)
	}
	time.Sleep(time.Second * time.Duration(2+numMiners))
	t.Cleanup(func() {
		if err := pool.Purge(lotusDevnet); err != nil {
			panic(fmt.Sprintf("couldn't purge lotus-devnet from docker pool: %s", err))
		}
	})
	debug := false
	if debug {
		go func() {
			opts := docker.LogsOptions{
				Context: context.Background(),

				Stderr:      true,
				Stdout:      true,
				Follow:      true,
				Timestamps:  true,
				RawTerminal: true,

				Container: lotusDevnet.Container.ID,

				OutputStream: os.Stdout,
			}

			if err := pool.Client.Logs(opts); err != nil {
				panic(err)
			}
		}()
	}
	return lotusDevnet
}

// CreateLocalDevnetWithIPFS creates a local devnet connected to an IPFS node.
func CreateLocalDevnetWithIPFS(t *testing.T, numMiners int, ipfsMaddr string) (*apistruct.FullNodeStruct, address.Address, []address.Address) {
	lotusDevnet := LaunchDevnetDocker(t, numMiners, ipfsMaddr)
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

// CreateLocalDevnet returns an API client that targets a local devnet with numMiners number
// of miners. Refer to http://github.com/textileio/local-devnet for more information.
func CreateLocalDevnet(t *testing.T, numMiners int) (*apistruct.FullNodeStruct, address.Address, []address.Address) {
	return CreateLocalDevnetWithIPFS(t, numMiners, "")
}

func devnetEnv(name string, value interface{}) string {
	return fmt.Sprintf("TEXLOTUSDEVNET_%s=%s", name, value)
}
