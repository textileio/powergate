package integrationtest

import (
	"bytes"
	"context"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/api/store"
	"github.com/textileio/powergate/ffs/coreipfs"
	"github.com/textileio/powergate/ffs/filcold"
	"github.com/textileio/powergate/ffs/minerselector/fixed"
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/ffs/scheduler/jsonjobstore"
	"github.com/textileio/powergate/tests"
	txndstr "github.com/textileio/powergate/txndstransform"
	"github.com/textileio/powergate/util"
	"github.com/textileio/powergate/wallet"
)

var (
	ipfsDocker *dockertest.Resource
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelError)

	// logging.SetLogLevel("scheduler", "debug")
	// logging.SetLogLevel("api", "debug")
	// logging.SetLogLevel("jobstore", "debug")
	// logging.SetLogLevel("coreipfs", "debug")

	var cls func()
	ipfsDocker, cls = tests.LaunchDocker()
	code := m.Run()
	cls()
	os.Exit(code)
}

func TestDefaultConfig(t *testing.T) {
	_, fapi, cls := newApi(t)
	defer cls()

	new := ffs.CidConfig{
		Hot: ffs.HotConfig{
			Ipfs: ffs.IpfsConfig{
				Enabled: false,
			},
		},
		Cold: ffs.ColdConfig{
			Filecoin: ffs.FilecoinConfig{
				Enabled: false,
			},
		},
	}
	err := fapi.SetDefaultCidConfig(new)
	require.Nil(t, err)
	gNew := fapi.GetDefaultCidConfig()
	require.Equal(t, new, gNew)
}

func TestAdd(t *testing.T) {
	ipfsApi, fapi, cls := newApi(t)
	defer cls()

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfsApi)
	t.Run("AddCidSuccess", func(t *testing.T) {
		jid, err := fapi.AddCid(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Done)
	})
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	ipfs, fapi, cls := newApi(t)
	defer cls()

	r := rand.New(rand.NewSource(11))
	cid, data := addRandomFile(t, r, ipfs)
	jid, err := fapi.AddCid(cid)
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Done)

	t.Run("FromAPI", func(t *testing.T) {
		r, err := fapi.Get(ctx, cid)
		require.Nil(t, err)
		fetched, err := ioutil.ReadAll(r)
		require.Nil(t, err)
		require.True(t, bytes.Equal(data, fetched))
	})
}

func TestInfo(t *testing.T) {
	ctx := context.Background()
	ipfs, fapi, cls := newApi(t)
	defer cls()

	var err error
	var first api.InstanceInfo
	t.Run("Minimal", func(t *testing.T) {
		first, err = fapi.Info(ctx)
		require.Nil(t, err)
		require.NotEmpty(t, first.ID)
		require.NotEmpty(t, first.Wallet.Address)
		require.Greater(t, first.Wallet.Balance, uint64(0))
		require.Equal(t, len(first.Pins), 0)
	})

	r := rand.New(rand.NewSource(22))
	n := 3
	for i := 0; i < n; i++ {
		cid, _ := addRandomFile(t, r, ipfs)
		jid, err := fapi.AddCid(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Done)
	}

	t.Run("WithThreeAdds", func(t *testing.T) {
		second, err := fapi.Info(ctx)
		require.Nil(t, err)
		require.Equal(t, second.ID, first.ID)
		require.Equal(t, second.Wallet.Address, first.Wallet.Address)
		require.Less(t, second.Wallet.Balance, first.Wallet.Balance)
		require.Equal(t, n, len(second.Pins))
	})
}

func TestShow(t *testing.T) {
	ctx := context.Background()
	ipfs, fapi, cls := newApi(t)
	defer cls()

	t.Run("NotStored", func(t *testing.T) {
		c, _ := cid.Decode("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")
		_, err := fapi.Show(c)
		require.Equal(t, api.ErrNotStored, err)
	})

	t.Run("Success", func(t *testing.T) {
		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfs)
		jid, err := fapi.AddCid(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Done)

		inf, err := fapi.Info(ctx)
		require.Nil(t, err)
		require.Equal(t, 1, len(inf.Pins))

		c := inf.Pins[0]

		s, err := fapi.Show(c)
		require.Nil(t, err)

		require.True(t, s.Cid.Defined())
		require.True(t, time.Now().After(s.Created))
		require.Greater(t, s.Hot.Size, 0)
		require.NotNil(t, s.Hot.Ipfs)
		require.True(t, time.Now().After(s.Hot.Ipfs.Created))
		require.NotNil(t, s.Cold.Filecoin)
		require.True(t, s.Cold.Filecoin.PayloadCID.Defined())
		require.Greater(t, s.Cold.Filecoin.Duration, uint64(0))
		require.Equal(t, 1, len(s.Cold.Filecoin.Proposals))
		p := s.Cold.Filecoin.Proposals[0]
		require.True(t, p.ProposalCid.Defined())
		require.False(t, p.Failed)
	})
}

func TestColdInstanceLoad(t *testing.T) {
	ctx := context.Background()
	ds := tests.NewTxMapDatastore()
	dnet, addr, _, closeDevnet := tests.CreateLocalDevnet(t, 1)
	defer closeDevnet()

	ipfsApi, fapi, cls := newApiFromDs(t, ds, ffs.EmptyID, dnet.Client, addr)
	ra := rand.New(rand.NewSource(22))
	cid, data := addRandomFile(t, ra, ipfsApi)
	jid, err := fapi.AddCid(cid)
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Done)
	info, err := fapi.Info(ctx)
	require.Nil(t, err)
	shw, err := fapi.Show(cid)
	require.Nil(t, err)
	cls()

	_, fapi, cls = newApiFromDs(t, ds, fapi.ID(), dnet.Client, addr)
	defer cls()
	ninfo, err := fapi.Info(ctx)
	require.Nil(t, err)
	require.Equal(t, info, ninfo)

	nshw, err := fapi.Show(cid)
	require.Nil(t, err)
	require.Equal(t, shw, nshw)

	r, err := fapi.Get(ctx, cid)
	require.Nil(t, err)
	fetched, err := ioutil.ReadAll(r)
	require.Nil(t, err)
	require.True(t, bytes.Equal(data, fetched))
}

func newApi(t *testing.T) (*httpapi.HttpApi, *api.Instance, func()) {
	ds := tests.NewTxMapDatastore()
	dnet, addr, _, close := tests.CreateLocalDevnet(t, 1)
	ipfsApi, fapi, closeInternal := newApiFromDs(t, ds, ffs.EmptyID, dnet.Client, addr)
	return ipfsApi, fapi, func() {
		closeInternal()
		close()
	}
}

func newApiFromDs(t *testing.T, ds datastore.TxnDatastore, iid ffs.InstanceID, client *apistruct.FullNodeStruct, waddr address.Address) (*httpapi.HttpApi, *api.Instance, func()) {
	ctx := context.Background()
	ipfsAddr := util.MustParseAddr("/ip4/127.0.0.1/tcp/" + ipfsDocker.GetPort("5001/tcp"))
	ipfsClient, err := httpapi.NewApi(ipfsAddr)
	require.Nil(t, err)

	dm, err := deals.New(client, deals.WithImportPath(filepath.Join(os.TempDir(), "imports")))
	require.Nil(t, err)

	ms := fixed.New("t0300", 4000000)
	cl := filcold.New(ms, dm, ipfsClient.Dag())
	jobstore := jsonjobstore.New(txndstr.Wrap(ds, "ffs/scheduler/jsonjobstore"))
	hl := coreipfs.New(ipfsClient)
	sched := scheduler.New(jobstore, hl, cl)

	wm, err := wallet.New(client, &waddr, *big.NewInt(5000000000000))
	require.Nil(t, err)

	var fapi *api.Instance
	if iid == ffs.EmptyID {
		iid = ffs.NewInstanceID()
		confstore := store.New(iid, txndstr.Wrap(ds, "ffs/api/store"))
		defConfig := ffs.CidConfig{
			Hot: ffs.HotConfig{
				Ipfs: ffs.IpfsConfig{
					Enabled: true,
				},
			},
			Cold: ffs.ColdConfig{
				Filecoin: ffs.FilecoinConfig{
					Enabled: true,
				},
			},
		}
		fapi, err = api.New(ctx, iid, confstore, sched, wm, defConfig)
		require.Nil(t, err)
	} else {
		confstore := store.New(iid, txndstr.Wrap(ds, "ffs/api/store"))
		fapi, err = api.Load(iid, confstore, sched, wm)
		require.Nil(t, err)
	}
	time.Sleep(time.Second)

	return ipfsClient, fapi, func() {
		if err := fapi.Close(); err != nil {
			t.Fatalf("closing api: %s", err)
		}
		if err := sched.Close(); err != nil {
			t.Fatalf("closing scheduler: %s", err)
		}
		if err := jobstore.Close(); err != nil {
			t.Fatalf("closing jobstore: %s", err)
		}
	}
}

func requireJobState(t *testing.T, fapi *api.Instance, jid ffs.JobID, status ffs.JobStatus) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ch, err := fapi.Watch(jid)
	require.Nil(t, err)
	defer fapi.Unwatch(ch)
	stop := false
	for !stop {
		select {
		case <-ctx.Done():
			t.Fatalf("waiting for job update timeout")
		case job, ok := <-ch:
			require.True(t, ok)
			require.Equal(t, jid, job.ID)
			if job.Status == ffs.Queued || job.Status == ffs.InProgress {
				continue
			}
			require.Equal(t, status, job.Status)
			stop = true
		}
	}
}

func randomBytes(r *rand.Rand, size int) []byte {
	buf := make([]byte, size)
	_, _ = r.Read(buf)
	return buf
}

func addRandomFile(t *testing.T, r *rand.Rand, ipfs *httpapi.HttpApi) (cid.Cid, []byte) {
	t.Helper()
	data := randomBytes(r, 500)
	node, err := ipfs.Unixfs().Add(context.Background(), ipfsfiles.NewReaderFile(bytes.NewReader(data)), options.Unixfs.Pin(false))
	require.Nil(t, err)

	return node.Cid(), data
}
