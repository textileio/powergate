package ldevnet

import (
	"bytes"
	"context"
	"crypto/rand"
	"io/ioutil"
	"net/http/httptest"
	"testing"
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
	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	build.SectorSizes = []uint64{1024}
	build.MinimumMinerPower = 1024
}

type LocalDevnet struct {
	Client *apistruct.FullNodeStruct
	Miner  address.Address

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

func New(t *testing.T, numMiners int) (*LocalDevnet, error) {
	miners := make([]int, numMiners)
	n, sn, closer := rpcBuilder(t, 1, miners)
	client := n[0].FullNode.(*apistruct.FullNodeStruct)
	miner := sn[0]
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

		mine := true
		go func(i int) {
			defer func() { done <- struct{}{} }()
			for mine {
				time.Sleep(10 * time.Millisecond)
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
	minerAddr, err := miner.ActorAddress(ctx)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond * 500) // Give time to mine at least 1 block
	return &LocalDevnet{
		Client:    client,
		Miner:     minerAddr,
		closer:    closer,
		cancel:    cancel,
		done:      done,
		numMiners: numMiners,
	}, nil
}

func rpcBuilder(t *testing.T, nFull int, storage []int) ([]test.TestNode, []test.TestStorageNode, func()) {
	fullApis, storaApis := mockSbBuilder(t, nFull, storage)
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
			t.Fatal(err)
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
			t.Fatal(err)
		}
		storers[i].MineOne = a.MineOne
	}

	return fulls, storers, func() {
		for _, c := range closers {
			c()
		}
	}
}
func mockSbBuilder(t *testing.T, nFull int, storage []int) ([]test.TestNode, []test.TestStorageNode) {
	ctx := context.Background()
	mn := mocknet.New(ctx)

	fulls := make([]test.TestNode, nFull)
	storers := make([]test.TestStorageNode, len(storage))

	pk, _, err := crypto.GenerateEd25519Key(rand.Reader)
	require.NoError(t, err)

	gmc := &gen.GenMinerCfg{
		PreSeals: map[string]genesis.GenesisMiner{},
	}
	for i := 0; i < len(storage); i++ {
		pid, err := peer.IDFromPrivateKey(pk)
		require.NoError(t, err)
		gmc.PeerIDs = append(gmc.PeerIDs, pid)
	}

	// PRESEAL SECTION, TRY TO REPLACE WITH BETTER IN THE FUTURE
	// TODO: would be great if there was a better way to fake the preseals
	var presealDirs []string
	for i := 0; i < len(storage); i++ {
		maddr, err := address.NewIDAddress(300 + uint64(i))
		if err != nil {
			t.Fatal(err)
		}
		tdir, err := ioutil.TempDir("", "preseal-memgen")
		if err != nil {
			t.Fatal(err)
		}
		genm, err := sbmock.PreSeal(1024, maddr, 1)
		if err != nil {
			t.Fatal(err)
		}

		presealDirs = append(presealDirs, tdir)
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
			t.Fatal(err)
		}

	}

	for i, full := range storage {
		if full != 0 {
			t.Fatal("storage nodes only supported on the first full node")
		}

		f := fulls[full]

		genMiner := gmc.MinerAddrs[i]
		wa := gmc.PreSeals[genMiner.String()].Worker

		storers[i] = testStorageNode(ctx, t, wa, genMiner, pk, f, mn, node.Options(
			node.Override(new(sectorbuilder.Interface), sbmock.NewMockSectorBuilder(5, build.SectorSizes[0])),
		))
	}

	if err := mn.LinkAll(); err != nil {
		t.Fatal(err)
	}

	return fulls, storers
}

func testStorageNode(ctx context.Context, t *testing.T, waddr address.Address, act address.Address, pk crypto.PrivKey, tnd test.TestNode, mn mocknet.Mocknet, opts node.Option) test.TestStorageNode {
	r := repo.NewMemory(nil)

	lr, err := r.Lock(repo.StorageMiner)
	require.NoError(t, err)

	ks, err := lr.KeyStore()
	require.NoError(t, err)

	kbytes, err := pk.Bytes()
	require.NoError(t, err)

	err = ks.Put("libp2p-host", types.KeyInfo{
		Type:       "libp2p-host",
		PrivateKey: kbytes,
	})
	require.NoError(t, err)

	ds, err := lr.Datastore("/metadata")
	require.NoError(t, err)
	err = ds.Put(datastore.NewKey("miner-address"), act.Bytes())
	require.NoError(t, err)

	err = lr.Close()
	require.NoError(t, err)

	peerid, err := peer.IDFromPrivateKey(pk)
	require.NoError(t, err)

	enc, err := actors.SerializeParams(&actors.UpdatePeerIDParams{PeerID: peerid})
	require.NoError(t, err)

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
	require.NoError(t, err)

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
		t.Fatalf("failed to construct node: %v", err)
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

	return test.TestStorageNode{StorageMiner: minerapi, MineOne: mineOne}
}
