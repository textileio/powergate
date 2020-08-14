package integrationtest

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/deals"
	dealsModule "github.com/textileio/powergate/deals/module"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/cidlogger"
	"github.com/textileio/powergate/ffs/coreipfs"
	"github.com/textileio/powergate/ffs/filcold"
	"github.com/textileio/powergate/ffs/manager"
	"github.com/textileio/powergate/ffs/minerselector/fixed"
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/filchain"
	paych "github.com/textileio/powergate/paych/lotus"
	"github.com/textileio/powergate/tests"
	txndstr "github.com/textileio/powergate/txndstransform"
	"github.com/textileio/powergate/util"
	walletModule "github.com/textileio/powergate/wallet/module"
)

const (
	iWalletBal int64 = 4000000000000000
)

// RequireIpfsUnpinnedCid checks that a cid is unpinned in the IPFS node.
func RequireIpfsUnpinnedCid(ctx context.Context, t require.TestingT, cid cid.Cid, ipfsAPI *httpapi.HttpApi) {
	pins, err := ipfsAPI.Pin().Ls(ctx)
	require.NoError(t, err)
	for _, p := range pins {
		require.NotEqual(t, cid, p.Path().Cid(), "Cid isn't unpined from IPFS node")
	}
}

// RequireIpfsPinnedCid checks that a cid is pinned in the IPFS node.
func RequireIpfsPinnedCid(ctx context.Context, t require.TestingT, cid cid.Cid, ipfsAPI *httpapi.HttpApi) {
	pins, err := ipfsAPI.Pin().Ls(ctx)
	require.NoError(t, err)

	pinned := false
	for _, p := range pins {
		if p.Path().Cid() == cid {
			pinned = true
			break
		}
	}
	require.True(t, pinned, "Cid should be pinned in IPFS node")
}

// RequireFilUnstored checks that a cid is not stored in the Filecoin network.
func RequireFilUnstored(ctx context.Context, t require.TestingT, client *apistruct.FullNodeStruct, c cid.Cid) {
	offers, err := client.ClientFindData(ctx, c, nil)
	require.NoError(t, err)
	require.Empty(t, offers)
}

// RequireFilStored cehcks that a cid is stored in the Filecoin network.
func RequireFilStored(ctx context.Context, t require.TestingT, client *apistruct.FullNodeStruct, c cid.Cid) {
	offers, err := client.ClientFindData(ctx, c, nil)
	require.NoError(t, err)
	require.NotEmpty(t, offers)
}

// NewAPI returns a new set of components for FFS.
func NewAPI(t tests.TestingTWithCleanup, numMiners int) (*httpapi.HttpApi, *apistruct.FullNodeStruct, *api.API, func()) {
	ds := tests.NewTxMapDatastore()
	ipfs, ipfsMAddr := CreateIPFS(t)
	addr, client, ms := NewDevnet(t, numMiners, ipfsMAddr)
	manager, closeManager := NewFFSManager(t, ds, client, addr, ms, ipfs)
	_, auth, err := manager.Create(context.Background())
	require.NoError(t, err)
	time.Sleep(time.Second * 3) // Wait for funding txn to finish.
	fapi, err := manager.GetByAuthToken(auth)
	require.NoError(t, err)
	return ipfs, client, fapi, func() {
		err := fapi.Close()
		require.NoError(t, err)
		closeManager()
	}
}

// CreateIPFS creates a docker container running IPFS.
func CreateIPFS(t tests.TestingTWithCleanup) (*httpapi.HttpApi, string) {
	ipfsDocker, cls := tests.LaunchIPFSDocker(t)
	t.Cleanup(cls)
	ipfsAddr := util.MustParseAddr("/ip4/127.0.0.1/tcp/" + ipfsDocker.GetPort("5001/tcp"))
	ipfs, err := httpapi.NewApi(ipfsAddr)
	require.NoError(t, err)
	bridgeIP := ipfsDocker.Container.NetworkSettings.Networks["bridge"].IPAddress
	ipfsDockerMAddr := fmt.Sprintf("/ip4/%s/tcp/5001", bridgeIP)

	return ipfs, ipfsDockerMAddr
}

// NewDevnet creates a localnet.
func NewDevnet(t tests.TestingTWithCleanup, numMiners int, ipfsAddr string) (address.Address, *apistruct.FullNodeStruct, ffs.MinerSelector) {
	client, addr, _ := tests.CreateLocalDevnetWithIPFS(t, numMiners, ipfsAddr, false)
	addrs := make([]string, numMiners)
	for i := 0; i < numMiners; i++ {
		addrs[i] = fmt.Sprintf("t0%d", 1000+i)
	}

	fixedMiners := make([]fixed.Miner, len(addrs))
	for i, a := range addrs {
		fixedMiners[i] = fixed.Miner{Addr: a, Country: "China", EpochPrice: 500000000}
	}
	ms := fixed.New(fixedMiners)
	return addr, client, ms
}

// NewFFSManager returns a new FFS manager.
func NewFFSManager(t require.TestingT, ds datastore.TxnDatastore, lotusClient *apistruct.FullNodeStruct, masterAddr address.Address, ms ffs.MinerSelector, ipfsClient *httpapi.HttpApi) (*manager.Manager, func()) {
	dm, err := dealsModule.New(txndstr.Wrap(ds, "deals"), lotusClient)
	require.NoError(t, err)

	fchain := filchain.New(lotusClient)
	l := cidlogger.New(txndstr.Wrap(ds, "ffs/cidlogger"))
	cl := filcold.New(ms, dm, ipfsClient, fchain, l)
	hl, err := coreipfs.New(ipfsClient, l)
	require.NoError(t, err)
	sched, err := scheduler.New(txndstr.Wrap(ds, "ffs/scheduler"), l, hl, cl)
	require.NoError(t, err)

	wm, err := walletModule.New(lotusClient, masterAddr, *big.NewInt(iWalletBal), false, "")
	require.NoError(t, err)

	pm := paych.New(lotusClient)

	manager, err := manager.New(ds, wm, pm, dm, sched, false)
	require.NoError(t, err)
	err = manager.SetDefaultStorageConfig(ffs.StorageConfig{
		Hot: ffs.HotConfig{
			Enabled:       true,
			AllowUnfreeze: false,
			Ipfs: ffs.IpfsConfig{
				AddTimeout: 10,
			},
		},
		Cold: ffs.ColdConfig{
			Enabled: true,
			Filecoin: ffs.FilConfig{
				ExcludedMiners:  nil,
				DealMinDuration: util.MinDealDuration,
				RepFactor:       1,
			},
		},
	})
	require.NoError(t, err)

	return manager, func() {
		if err := manager.Close(); err != nil {
			t.Errorf("closing api: %s", err)
			t.FailNow()
		}
		if err := sched.Close(); err != nil {
			t.Errorf("closing scheduler: %s", err)
			t.FailNow()
		}
		if err := l.Close(); err != nil {
			t.Errorf("closing cidlogger: %s", err)
			t.FailNow()
		}
	}
}

// RequireJobState watches a Job for a desired status.
func RequireJobState(t require.TestingT, fapi *api.API, jid ffs.JobID, status ffs.JobStatus) ffs.Job {
	ch := make(chan ffs.Job)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	go func() {
		err = fapi.WatchJobs(ctx, ch, jid)
		close(ch)
	}()
	stop := false
	var res ffs.Job
	for !stop {
		select {
		case <-time.After(120 * time.Second):
			t.Errorf("waiting for job update timeout")
			t.FailNow()
		case job, ok := <-ch:
			require.True(t, ok)
			require.Equal(t, jid, job.ID)
			if job.Status == ffs.Queued || job.Status == ffs.Executing {
				if job.Status == status {
					stop = true
					res = job
				}
				continue
			}
			require.Equal(t, status, job.Status, job.ErrCause)
			stop = true
			res = job
		}
	}
	require.NoError(t, err)
	return res
}

// RequireStorageConfig compares a cid storage config against a target.
func RequireStorageConfig(t require.TestingT, fapi *api.API, c cid.Cid, config *ffs.StorageConfig) {
	if config == nil {
		defConfig := fapi.DefaultStorageConfig()
		config = &defConfig
	}
	currentConfig, err := fapi.GetStorageConfig(c)
	require.NoError(t, err)
	require.Equal(t, *config, currentConfig)
}

// RequireStorageDealRecord checks that a storage deal record exist for a cid.
func RequireStorageDealRecord(t require.TestingT, fapi *api.API, c cid.Cid) {
	time.Sleep(time.Second)
	recs, err := fapi.ListStorageDealRecords(deals.WithIncludeFinal(true))
	require.NoError(t, err)
	require.Len(t, recs, 1)
	require.Equal(t, c, recs[0].RootCid)
}

// RequireRetrievalDealRecord checks that a retrieval deal record exits for a cid.
func RequireRetrievalDealRecord(t require.TestingT, fapi *api.API, c cid.Cid) {
	recs, err := fapi.ListRetrievalDealRecords()
	require.NoError(t, err)
	require.Len(t, recs, 1)
	require.Equal(t, c, recs[0].DealInfo.RootCid)
}

// RandomBytes returns a slice of random bytes of a desired size.
func RandomBytes(r *rand.Rand, size int) []byte {
	buf := make([]byte, size)
	_, _ = r.Read(buf)
	return buf
}

// AddRandomFile adds a random file to the IPFS node.
func AddRandomFile(t require.TestingT, r *rand.Rand, ipfs *httpapi.HttpApi) (cid.Cid, []byte) {
	return AddRandomFileSize(t, r, ipfs, 1600)
}

// AddRandomFileSize adds a random file with a specified size to the IPFS node.
func AddRandomFileSize(t require.TestingT, r *rand.Rand, ipfs *httpapi.HttpApi, size int) (cid.Cid, []byte) {
	data := RandomBytes(r, size)
	node, err := ipfs.Unixfs().Add(context.Background(), ipfsfiles.NewReaderFile(bytes.NewReader(data)), options.Unixfs.Pin(false))
	if err != nil {
		t.Errorf("error adding random file: %s", err)
		t.FailNow()
	}
	return node.Cid(), data
}
