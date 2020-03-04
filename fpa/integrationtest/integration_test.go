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
	"github.com/ipfs/go-cid"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/require"
	"github.com/textileio/fil-tools/deals"
	"github.com/textileio/fil-tools/fpa"
	"github.com/textileio/fil-tools/fpa/coreipfs"
	"github.com/textileio/fil-tools/fpa/fastapi"
	"github.com/textileio/fil-tools/fpa/fastapi/store"
	"github.com/textileio/fil-tools/fpa/filcold"
	"github.com/textileio/fil-tools/fpa/minerselector/fixed"
	"github.com/textileio/fil-tools/fpa/scheduler"
	"github.com/textileio/fil-tools/fpa/scheduler/jsonjobstore"
	"github.com/textileio/fil-tools/tests"
	txndstr "github.com/textileio/fil-tools/txndstransform"
	"github.com/textileio/fil-tools/util"
	"github.com/textileio/fil-tools/wallet"
)

var (
	ipfsDocker *dockertest.Resource
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelError)

	// logging.SetLogLevel("scheduler", "debug")
	// logging.SetLogLevel("fastapi", "debug")
	// logging.SetLogLevel("jobstore", "debug")
	// logging.SetLogLevel("coreipfs", "debug")

	var cls func()
	ipfsDocker, cls = tests.LaunchDocker()
	code := m.Run()
	cls()
	os.Exit(code)
}

func TestAdd(t *testing.T) {
	ipfsApi, fapi, cls := newFastAPI(t)
	defer cls()

	t.Run("Unretrievable", func(t *testing.T) {
		c, _ := cid.Decode("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z") // ipfs hello-world not in the node
		jid, err := fapi.AddCid(c)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, fpa.Failed)
	})

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfsApi)
	t.Run("AddCidSuccess", func(t *testing.T) {
		jid, err := fapi.AddCid(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, fpa.Done)
	})
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	ipfs, fapi, cls := newFastAPI(t)
	defer cls()

	r := rand.New(rand.NewSource(11))
	cid, data := addRandomFile(t, r, ipfs)
	jid, err := fapi.AddCid(cid)
	require.Nil(t, err)
	requireJobState(t, fapi, jid, fpa.Done)

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
	ipfs, fapi, cls := newFastAPI(t)
	defer cls()

	var err error
	var first fastapi.InstanceInfo
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
		requireJobState(t, fapi, jid, fpa.Done)
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
	ipfs, fapi, cls := newFastAPI(t)
	defer cls()

	t.Run("NotStored", func(t *testing.T) {
		c, _ := cid.Decode("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")
		_, err := fapi.Show(c)
		require.Equal(t, fastapi.ErrNotStored, err)
	})

	t.Run("Success", func(t *testing.T) {
		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfs)
		jid, err := fapi.AddCid(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, fpa.Done)

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

func newFastAPI(t *testing.T) (*httpapi.HttpApi, *fastapi.Instance, func()) {
	ctx := context.Background()
	ds := tests.NewTxMapDatastore()

	ipfsAddr := util.MustParseAddr("/ip4/127.0.0.1/tcp/" + ipfsDocker.GetPort("5001/tcp"))
	ipfsClient, err := httpapi.NewApi(ipfsAddr)
	require.Nil(t, err)

	dnet, addr, _, close := tests.CreateLocalDevnet(t, 1)
	dm, err := deals.New(dnet.Client, deals.WithImportPath(filepath.Join(os.TempDir(), "imports")))
	require.Nil(t, err)

	ms := fixed.New("t0300", 4000000)

	jobstore := jsonjobstore.New(txndstr.Wrap(ds, "fpa/scheduler/jsonjobstore"))
	hl := coreipfs.New(ipfsClient)
	sch := scheduler.New(jobstore, hl, filcold.New(ms, dm, ipfsClient.Dag()))

	wm, err := wallet.New(dnet.Client, &addr, *big.NewInt(5000000000000))
	require.Nil(t, err)
	id := fpa.NewInstanceID()
	fapi, err := fastapi.New(ctx, id, store.New(id, txndstr.Wrap(ds, "fpa/fastapi/store")), sch, wm)
	require.Nil(t, err)
	time.Sleep(time.Second)

	return ipfsClient, fapi, func() {
		fapi.Close()
		sch.Close()
		close()
	}
}

func requireJobState(t *testing.T, fapi *fastapi.Instance, jid fpa.JobID, status fpa.JobStatus) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ch := fapi.Watch(jid)
	defer fapi.Unwatch(ch)
	stop := false
	for !stop {
		select {
		case <-ctx.Done():
			t.Fatalf("waiting for job update timeout")
		case job, ok := <-ch:
			require.True(t, ok)
			require.Equal(t, jid, job.ID)
			if job.Status == fpa.Queued {
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

type walletManagerMock struct {
	addr address.Address
}

func (wmm *walletManagerMock) NewWallet(ctx context.Context, typ string) (string, error) {
	return wmm.addr.String(), nil
}
