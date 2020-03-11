package ldevnet

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net/http/httptest"
	"os"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-sectorbuilder"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/api/test"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/gen"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/genesis"
	"github.com/filecoin-project/lotus/lib/jsonrpc"
	"github.com/filecoin-project/lotus/miner"
	"github.com/filecoin-project/lotus/node"
	"github.com/filecoin-project/lotus/node/modules"
	modtest "github.com/filecoin-project/lotus/node/modules/testing"
	"github.com/filecoin-project/lotus/node/repo"
	"github.com/filecoin-project/lotus/storage/sbmock"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
)

var DefaultDuration = time.Millisecond * 100

func init() {
	build.SectorSizes = []uint64{1024}
	build.MinimumMinerPower = 1024
	os.Setenv("TRUST_PARAMS", "1")
	delay := os.Getenv("TEXPOWERGATE_CI")
	if delay != "" {
		DefaultDuration = time.Millisecond * 5000
		fmt.Println("Using long-delay since TEXPOWERGATE_CI is set")
	}

}

type LocalDevnet struct {
	Client *apistruct.FullNodeStruct

	numMiners int
	cancel    context.CancelFunc
	closer    func()
	done      chan struct{}
}

func (ld *LocalDevnet) Close() {
	ld.cancel()

	for i := 0; i < ld.numMiners; i++ {
		<-ld.done
	}
	close(ld.done)
	ld.closer()
}

func New(numMiners int, blockDur time.Duration) (*LocalDevnet, error) {
	miners := make([]int, numMiners)
	n, sn, closer, err := rpcBuilder(1, miners)
	if err != nil {
		return nil, err
	}
	client := n[0].FullNode.(*apistruct.FullNodeStruct)
	ctx, cancel := context.WithCancel(context.Background())
	addrinfo, err := client.NetAddrsListen(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	done := make(chan struct{})
	for i := range miners {
		if err := sn[i].NetConnect(ctx, addrinfo); err != nil {
			cancel()
			return nil, err
		}
		time.Sleep(time.Second)

		mine := i == 0
		go func(i int) {
			defer func() { done <- struct{}{} }()
			for mine {
				time.Sleep(blockDur)
				if ctx.Err() != nil {
					mine = false
					continue
				}
				if err := sn[i].MineOne(context.Background()); err != nil {
					panic(err)
				}
			}
		}(i)
	}
	for i := range miners {
		for j := range miners {
			if j == i {
				continue
			}
			mainfo, err := sn[j].NetAddrsListen(ctx)
			if err != nil {
				cancel()
				return nil, err
			}
			if err := sn[i].NetConnect(ctx, mainfo); err != nil {
				cancel()
				return nil, err
			}
		}
	}

	time.Sleep(blockDur * 5) // Give time to mine at least 1 block
	return &LocalDevnet{
		Client:    client,
		closer:    closer,
		cancel:    cancel,
		done:      done,
		numMiners: numMiners,
	}, nil
}

func rpcBuilder(nFull int, storage []int) ([]test.TestNode, []test.TestStorageNode, func(), error) {
	fullApis, storaApis, err := mockSbBuilder(nFull, storage)
	if err != nil {
		return nil, nil, nil, err
	}
	fulls := make([]test.TestNode, nFull)
	storers := make([]test.TestStorageNode, len(storage))

	var closers []func()
	for i, a := range fullApis {
		rpcServer := jsonrpc.NewServer()
		rpcServer.Register("Filecoin", a)
		testServ := httptest.NewServer(rpcServer)
		closers = append(closers, func() { testServ.Close() })
		var err error
		fulls[i].FullNode, _, err = client.NewFullNodeRPC("ws://"+testServ.Listener.Addr().String(), nil)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	for i, a := range storaApis {
		rpcServer := jsonrpc.NewServer()
		rpcServer.Register("Filecoin", a)
		testServ := httptest.NewServer(rpcServer)
		closers = append(closers, func() { testServ.Close() })

		var err error
		storers[i].StorageMiner, _, err = client.NewStorageMinerRPC("ws://"+testServ.Listener.Addr().String(), nil)
		if err != nil {
			return nil, nil, nil, err
		}
		storers[i].MineOne = a.MineOne
	}

	return fulls, storers, func() {
		for _, c := range closers {
			c()
		}
	}, nil
}
func mockSbBuilder(nFull int, storage []int) ([]test.TestNode, []test.TestStorageNode, error) {
	ctx := context.Background()
	mn := mocknet.New(ctx)

	fulls := make([]test.TestNode, nFull)
	storers := make([]test.TestStorageNode, len(storage))

	gmc := &gen.GenMinerCfg{
		PreSeals: map[string]genesis.GenesisMiner{},
	}

	var storagePks []crypto.PrivKey
	for i := 0; i < len(storage); i++ {
		pk, _, err := crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		pid, err := peer.IDFromPrivateKey(pk)
		if err != nil {
			return nil, nil, err
		}
		gmc.PeerIDs = append(gmc.PeerIDs, pid)
		storagePks = append(storagePks, pk)
	}

	// PRESEAL SECTION, TRY TO REPLACE WITH BETTER IN THE FUTURE
	// TODO: would be great if there was a better way to fake the preseals
	for i := 0; i < len(storage); i++ {
		maddr, err := address.NewIDAddress(300 + uint64(i))
		if err != nil {
			return nil, nil, err
		}
		genm, err := sbmock.PreSeal(1024, maddr, 1)
		if err != nil {
			return nil, nil, err
		}

		gmc.MinerAddrs = append(gmc.MinerAddrs, maddr)
		gmc.PreSeals[maddr.String()] = *genm
	}

	// END PRESEAL SECTION

	var genbuf bytes.Buffer
	for i := 0; i < nFull; i++ {
		var genesis node.Option
		if i == 0 {
			genesis = node.Override(new(modules.Genesis), modtest.MakeGenesisMem(&genbuf, gmc))
		} else {
			genesis = node.Override(new(modules.Genesis), modules.LoadGenesis(genbuf.Bytes()))
		}

		var err error
		// TODO: Don't ignore stop
		_, err = node.New(ctx,
			node.FullAPI(&fulls[i].FullNode),
			node.Online(),
			node.Repo(repo.NewMemory(nil)),
			node.MockHost(mn),
			node.Test(),

			node.Override(new(sectorbuilder.Verifier), sbmock.MockVerifier),

			genesis,
		)
		if err != nil {
			return nil, nil, err
		}

	}

	for i, full := range storage {
		if full != 0 {
			return nil, nil, fmt.Errorf("storage nodes only supported on the first full node")
		}

		f := fulls[full]

		genMiner := gmc.MinerAddrs[i]
		wa := gmc.PreSeals[genMiner.String()].Worker

		var err error
		storers[i], err = testStorageNode(ctx, wa, genMiner, storagePks[i], f, mn, node.Options(
			node.Override(new(sectorbuilder.Interface), sbmock.NewMockSectorBuilder(5, build.SectorSizes[0])),
		))
		if err != nil {
			return nil, nil, err
		}
	}

	if err := mn.LinkAll(); err != nil {
		return nil, nil, err
	}

	return fulls, storers, nil
}

func testStorageNode(ctx context.Context, waddr address.Address, act address.Address, pk crypto.PrivKey, tnd test.TestNode, mn mocknet.Mocknet, opts node.Option) (test.TestStorageNode, error) {
	r := repo.NewMemory(nil)

	lr, err := r.Lock(repo.StorageMiner)
	if err != nil {
		return test.TestStorageNode{}, err
	}

	ks, err := lr.KeyStore()
	if err != nil {
		return test.TestStorageNode{}, err
	}

	kbytes, err := pk.Bytes()
	if err != nil {
		return test.TestStorageNode{}, err
	}

	err = ks.Put("libp2p-host", types.KeyInfo{
		Type:       "libp2p-host",
		PrivateKey: kbytes,
	})
	if err != nil {
		return test.TestStorageNode{}, err
	}

	ds, err := lr.Datastore("/metadata")
	if err != nil {
		return test.TestStorageNode{}, err
	}
	err = ds.Put(datastore.NewKey("miner-address"), act.Bytes())
	if err != nil {
		return test.TestStorageNode{}, err
	}

	err = lr.Close()
	if err != nil {
		return test.TestStorageNode{}, err
	}

	peerid, err := peer.IDFromPrivateKey(pk)
	if err != nil {
		return test.TestStorageNode{}, err
	}

	enc, err := actors.SerializeParams(&actors.UpdatePeerIDParams{PeerID: peerid})
	if err != nil {
		return test.TestStorageNode{}, err
	}

	msg := &types.Message{
		To:       act,
		From:     waddr,
		Method:   actors.MAMethods.UpdatePeerID,
		Params:   enc,
		Value:    types.NewInt(0),
		GasPrice: types.NewInt(0),
		GasLimit: types.NewInt(1000000),
	}

	_, err = tnd.MpoolPushMessage(ctx, msg)
	if err != nil {
		return test.TestStorageNode{}, err
	}

	// start node
	var minerapi api.StorageMiner

	mineBlock := make(chan struct{})
	// TODO: use stop
	_, err = node.New(ctx,
		node.StorageMiner(&minerapi),
		node.Online(),
		node.Repo(r),
		node.Test(),

		node.MockHost(mn),

		node.Override(new(api.FullNode), tnd),
		node.Override(new(*miner.Miner), miner.NewTestMiner(mineBlock, act)),

		opts,
	)
	if err != nil {
		return test.TestStorageNode{}, err
	}

	/*// Bootstrap with full node
	remoteAddrs, err := tnd.NetAddrsListen(ctx)
	require.NoError(t, err)

	err = minerapi.NetConnect(ctx, remoteAddrs)
	require.NoError(t, err)*/
	mineOne := func(ctx context.Context) error {
		select {
		case mineBlock <- struct{}{}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return test.TestStorageNode{StorageMiner: minerapi, MineOne: mineOne}, nil
}
