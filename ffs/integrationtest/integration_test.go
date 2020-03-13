package integrationtest

import (
	"bytes"
	"context"
	"fmt"
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
	"github.com/textileio/powergate/ffs/api/istore"
	"github.com/textileio/powergate/ffs/coreipfs"
	"github.com/textileio/powergate/ffs/filcold"
	"github.com/textileio/powergate/ffs/minerselector/fixed"
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/ffs/scheduler/cistore"
	"github.com/textileio/powergate/ffs/scheduler/jstore"
	"github.com/textileio/powergate/ffs/scheduler/pcstore"
	"github.com/textileio/powergate/tests"
	txndstr "github.com/textileio/powergate/txndstransform"
	"github.com/textileio/powergate/util"
	"github.com/textileio/powergate/wallet"
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelError)

	// logging.SetLogLevel("scheduler", "debug")
	// logging.SetLogLevel("api", "debug")
	// logging.SetLogLevel("jobstore", "debug")
	// logging.SetLogLevel("coreipfs", "debug")

	os.Exit(m.Run())
}

func TestSetDefaultConfig(t *testing.T) {
	_, fapi, cls := newApi(t, 1)
	defer cls()

	config := ffs.DefaultCidConfig{
		Hot: ffs.HotConfig{
			Enabled: false,
			Ipfs: ffs.IpfsConfig{
				AddTimeout: 10,
			},
		},
		Cold: ffs.ColdConfig{
			Enabled: true,
			Filecoin: ffs.FilecoinConfig{
				DealDuration: 22333,
				RepFactor:    23,
			},
		},
	}
	err := fapi.SetDefaultCidConfig(config)
	require.Nil(t, err)
	newConfig := fapi.GetDefaultCidConfig(cid.Undef)
	require.Equal(t, newConfig.Hot, config.Hot)
	require.Equal(t, newConfig.Cold, config.Cold)
}

func TestAdd(t *testing.T) {
	ipfsApi, fapi, cls := newApi(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfsApi)
	t.Run("WithDefaultConfig", func(t *testing.T) {
		jid, err := fapi.PushConfig(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
	})

	cid, _ = addRandomFile(t, r, ipfsApi)
	t.Run("WithCustomConfig", func(t *testing.T) {
		config := fapi.GetDefaultCidConfig(cid).WithHotEnabled(false).WithColdFilDealDuration(int64(4321))
		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)

		cconfig, err := fapi.GetCidConfig(cid)
		require.Nil(t, err)
		require.Equal(t, config, cconfig)
	})
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	ipfs, fapi, cls := newApi(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	cid, data := addRandomFile(t, r, ipfs)
	jid, err := fapi.PushConfig(cid)
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)

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
	ipfs, fapi, cls := newApi(t, 1)
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
		jid, err := fapi.PushConfig(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
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
	ipfs, fapi, cls := newApi(t, 1)

	defer cls()

	t.Run("NotStored", func(t *testing.T) {
		c, _ := cid.Decode("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")
		_, err := fapi.Show(c)
		require.Equal(t, api.ErrNotStored, err)
	})

	t.Run("Success", func(t *testing.T) {
		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfs)
		jid, err := fapi.PushConfig(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)

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
		require.Equal(t, 1, len(s.Cold.Filecoin.Proposals))
		p := s.Cold.Filecoin.Proposals[0]
		require.True(t, p.ProposalCid.Defined())
		require.Greater(t, p.Duration, int64(0))
		require.False(t, p.Failed)
	})
}

func TestColdInstanceLoad(t *testing.T) {
	ctx := context.Background()
	ipfsDocker, cls := tests.LaunchDocker()
	t.Cleanup(func() { cls() })

	ds := tests.NewTxMapDatastore()
	addr, client, ms, closeDevnet := newDevnet(t, 1)
	defer closeDevnet()

	ipfsApi, fapi, cls := newApiFromDs(t, ds, ffs.EmptyInstanceID, client, addr, ms, ipfsDocker)
	ra := rand.New(rand.NewSource(22))
	cid, data := addRandomFile(t, ra, ipfsApi)
	jid, err := fapi.PushConfig(cid)
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	info, err := fapi.Info(ctx)
	require.Nil(t, err)
	shw, err := fapi.Show(cid)
	require.Nil(t, err)
	cls()

	_, fapi, cls = newApiFromDs(t, ds, fapi.ID(), client, addr, ms, ipfsDocker)
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

func TestRepFactor(t *testing.T) {
	ipfsApi, fapi, cls := newApi(t, 2)
	defer cls()

	rfs := []int{1, 2}
	r := rand.New(rand.NewSource(22))
	for _, rf := range rfs {
		t.Run(fmt.Sprintf("%d", rf), func(t *testing.T) {
			cid, _ := addRandomFile(t, r, ipfsApi)
			config := fapi.GetDefaultCidConfig(cid).WithColdFilRepFactor(rf)
			jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
			require.Nil(t, err)
			requireJobState(t, fapi, jid, ffs.Success)

			cinfo, err := fapi.GetCidInfo(cid)
			require.Nil(t, err)
			require.Equal(t, rf, len(cinfo.Cold.Filecoin.Proposals))
		})
	}

	t.Run("IncreaseBy1", func(t *testing.T) {
		t.SkipNow()
		cid, _ := addRandomFile(t, r, ipfsApi)
		jid, err := fapi.PushConfig(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)

		cinfo, err := fapi.GetCidInfo(cid)
		require.Nil(t, err)
		require.Equal(t, 1, len(cinfo.Cold.Filecoin.Proposals))
		firstProposal := cinfo.Cold.Filecoin.Proposals[0]

		config := fapi.GetDefaultCidConfig(cid).WithColdFilRepFactor(2)
		jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		cinfo, err = fapi.GetCidInfo(cid)
		require.Nil(t, err)
		require.Equal(t, 2, len(cinfo.Cold.Filecoin.Proposals))
		require.Contains(t, cinfo.Cold.Filecoin.Proposals, firstProposal)
	})
}

func TestHotTimeoutConfig(t *testing.T) {
	_, fapi, cls := newApi(t, 1)
	defer cls()

	t.Run("ShortTime", func(t *testing.T) {
		cid, _ := cid.Decode("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")
		config := fapi.GetDefaultCidConfig(cid).WithHotIpfsAddTimeout(1)
		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Failed)
	})
}

func TestDurationConfig(t *testing.T) {
	ipfsApi, fapi, cls := newApi(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfsApi)
	duration := int64(1234)
	config := fapi.GetDefaultCidConfig(cid).WithColdFilDealDuration(duration)
	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	cinfo, err := fapi.GetCidInfo(cid)
	require.Nil(t, err)
	p := cinfo.Cold.Filecoin.Proposals[0]
	require.Equal(t, duration, p.Duration)
	require.Greater(t, p.ActivationEpoch, uint64(0))
}

func TestFilecoinBlacklist(t *testing.T) {
	ipfsApi, fapi, cls := newApi(t, 2)
	defer cls()

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfsApi)
	excludedMiner := "t0300"
	config := fapi.GetDefaultCidConfig(cid).WithColdFilBlacklist([]string{excludedMiner})

	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	cinfo, err := fapi.GetCidInfo(cid)
	require.Nil(t, err)
	p := cinfo.Cold.Filecoin.Proposals[0]
	require.NotEqual(t, p.Miner, excludedMiner)
}

func TestFilecoinCountryFilter(t *testing.T) {
	ipfsDocker, cls := tests.LaunchDocker()
	t.Cleanup(func() { cls() })

	countries := []string{"China", "Uruguay"}
	numMiners := len(countries)
	dnet, addr, _, close := tests.CreateLocalDevnet(t, numMiners)
	defer close()
	addrs := make([]string, numMiners)
	for i := 0; i < numMiners; i++ {
		addrs[i] = fmt.Sprintf("t0%d", 300+i)
	}
	fixedMiners := make([]fixed.Miner, len(addrs))
	for i, a := range addrs {
		fixedMiners[i] = fixed.Miner{Addr: a, Country: countries[i], EpochPrice: 4000000}
	}
	ms := fixed.New(fixedMiners)
	ds := tests.NewTxMapDatastore()
	ipfsApi, fapi, closeInternal := newApiFromDs(t, ds, ffs.EmptyInstanceID, dnet.Client, addr, ms, ipfsDocker)
	defer closeInternal()

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfsApi)
	countryFilter := []string{"Uruguay"}
	config := fapi.GetDefaultCidConfig(cid).WithColdFilCountryCodes(countryFilter)

	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	cinfo, err := fapi.GetCidInfo(cid)
	require.Nil(t, err)
	p := cinfo.Cold.Filecoin.Proposals[0]
	require.Equal(t, p.Miner, "t0301")
}

func TestFilecoinEnableConfig(t *testing.T) {
	tableTest := []struct {
		HotEnabled  bool
		ColdEnabled bool
	}{
		{HotEnabled: true, ColdEnabled: true},
		{HotEnabled: false, ColdEnabled: true},
		{HotEnabled: true, ColdEnabled: false},
		{HotEnabled: false, ColdEnabled: false},
	}

	for _, tt := range tableTest {
		name := fmt.Sprintf("Hot(%v)/Cold(%v)", tt.HotEnabled, tt.ColdEnabled)
		t.Run(name, func(t *testing.T) {
			ipfsApi, fapi, cls := newApi(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := addRandomFile(t, r, ipfsApi)
			config := fapi.GetDefaultCidConfig(cid).WithColdEnabled(tt.ColdEnabled).WithHotEnabled(tt.HotEnabled)

			var expectedErr error
			if !tt.HotEnabled && !tt.ColdEnabled {
				expectedErr = ffs.ErrBothStoragesDisabled
			}
			jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
			require.Equal(t, expectedErr, err)
			if expectedErr != nil {
				return
			}

			expectedJobState := ffs.Success
			var errCause string
			if !tt.ColdEnabled && !tt.HotEnabled {
				expectedJobState = ffs.Failed
				errCause = ffs.ErrBothStoragesDisabled.Error()
			}
			job := requireJobState(t, fapi, jid, expectedJobState)
			require.Equal(t, errCause, job.ErrCause)

			if expectedJobState == ffs.Success {
				cinfo, err := fapi.GetCidInfo(cid)
				require.Nil(t, err)
				require.Equal(t, tt.HotEnabled, cinfo.Hot.Enabled)
				require.Equal(t, tt.ColdEnabled, cinfo.Cold.Enabled)

				ctx := context.Background()
				_, err = fapi.Get(ctx, cid)
				var expectedErr error
				if !tt.HotEnabled {
					expectedErr = ffs.ErrHotStorageDisabled
				}
				require.Equal(t, expectedErr, err)
			}

		})
	}
}

func TestUnfreeze(t *testing.T) {
	ipfsApi, fapi, cls := newApi(t, 1)
	defer cls()

	ra := rand.New(rand.NewSource(22))
	cid, data := addRandomFile(t, ra, ipfsApi)
	config := fapi.GetDefaultCidConfig(cid).WithHotEnabled(false)

	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)

	ctx := context.Background()
	_, err = fapi.Get(ctx, cid)
	require.Equal(t, ffs.ErrHotStorageDisabled, err)

	config = config.WithHotEnabled(true)
	jid, err = fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)

	r, err := fapi.Get(ctx, cid)
	require.Nil(t, err)
	fetched, err := ioutil.ReadAll(r)
	require.Nil(t, err)
	require.True(t, bytes.Equal(data, fetched))
}

func TestUnfreezeFail(t *testing.T) {

}

func newApi(t *testing.T, numMiners int) (*httpapi.HttpApi, *api.API, func()) {
	ipfsDocker, cls := tests.LaunchDocker()
	t.Cleanup(func() { cls() })
	ds := tests.NewTxMapDatastore()
	addr, client, ms, close := newDevnet(t, numMiners)
	ipfsApi, fapi, closeInternal := newApiFromDs(t, ds, ffs.EmptyInstanceID, client, addr, ms, ipfsDocker)
	return ipfsApi, fapi, func() {
		closeInternal()
		close()
	}
}

func newDevnet(t *testing.T, numMiners int) (address.Address, *apistruct.FullNodeStruct, ffs.MinerSelector, func()) {
	dnet, addr, _, close := tests.CreateLocalDevnet(t, numMiners)
	addrs := make([]string, numMiners)
	for i := 0; i < numMiners; i++ {
		addrs[i] = fmt.Sprintf("t0%d", 300+i)
	}

	fixedMiners := make([]fixed.Miner, len(addrs))
	for i, a := range addrs {
		fixedMiners[i] = fixed.Miner{Addr: a, Country: "China", EpochPrice: 4000000}
	}
	ms := fixed.New(fixedMiners)
	return addr, dnet.Client, ms, close
}

func newApiFromDs(t *testing.T, ds datastore.TxnDatastore, iid ffs.InstanceID, client *apistruct.FullNodeStruct, waddr address.Address, ms ffs.MinerSelector, ipfsDocker *dockertest.Resource) (*httpapi.HttpApi, *api.API, func()) {
	ctx := context.Background()
	ipfsAddr := util.MustParseAddr("/ip4/127.0.0.1/tcp/" + ipfsDocker.GetPort("5001/tcp"))
	ipfsClient, err := httpapi.NewApi(ipfsAddr)
	require.Nil(t, err)

	dm, err := deals.New(client, deals.WithImportPath(filepath.Join(os.TempDir(), "imports")))
	require.Nil(t, err)

	cl := filcold.New(ms, dm, ipfsClient.Dag())
	cis := cistore.New(txndstr.Wrap(ds, "ffs/scheduler/cistore"))
	pcs := pcstore.New(txndstr.Wrap(ds, "ffs/scheduler/pcstore"))
	js := jstore.New(txndstr.Wrap(ds, "ffs/scheduler/jstore"))
	hl := coreipfs.New(ipfsClient)
	sched := scheduler.New(js, pcs, cis, hl, cl)

	wm, err := wallet.New(client, &waddr, *big.NewInt(5000000000000))
	require.Nil(t, err)

	var fapi *api.API
	if iid == ffs.EmptyInstanceID {
		iid = ffs.NewInstanceID()
		is := istore.New(iid, txndstr.Wrap(ds, "ffs/api/istore"))
		defConfig := ffs.DefaultCidConfig{
			Hot: ffs.HotConfig{
				Enabled: true,
				Ipfs: ffs.IpfsConfig{
					AddTimeout: 30,
				},
			},
			Cold: ffs.ColdConfig{
				Enabled: true,
				Filecoin: ffs.FilecoinConfig{
					Blacklist:    nil,
					DealDuration: 1000,
					RepFactor:    1,
				},
			},
		}
		fapi, err = api.New(ctx, iid, is, sched, wm, defConfig)
		require.Nil(t, err)
	} else {
		is := istore.New(iid, txndstr.Wrap(ds, "ffs/api/istore"))
		fapi, err = api.Load(iid, is, sched, wm)
		require.Nil(t, err)
	}
	time.Sleep(time.Second * 2)

	return ipfsClient, fapi, func() {
		if err := fapi.Close(); err != nil {
			t.Fatalf("closing api: %s", err)
		}
		if err := sched.Close(); err != nil {
			t.Fatalf("closing scheduler: %s", err)
		}
		if err := js.Close(); err != nil {
			t.Fatalf("closing jobstore: %s", err)
		}
	}
}

func requireJobState(t *testing.T, fapi *api.API, jid ffs.JobID, status ffs.JobStatus) ffs.Job {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ch, err := fapi.Watch(jid)
	require.Nil(t, err)
	defer fapi.Unwatch(ch)
	stop := false
	var res ffs.Job
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
			require.Equal(t, status, job.Status, job.ErrCause)
			stop = true
			res = job
		}
	}
	fapi.Unwatch(ch)
	return res
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
	if err != nil {
		t.Fatalf("error adding random file: %s", err)
	}
	return node.Cid(), data
}
