package deals

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	dt "github.com/textileio/fil-tools/deals/types"
	"github.com/textileio/fil-tools/ldevnet"
	"github.com/textileio/fil-tools/tests"
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestStore(t *testing.T) {
	numMiners := []int{1, 3}
	for _, nm := range numMiners {
		t.Run(fmt.Sprintf("CantMiners%d", nm), func(t *testing.T) {
			dnet, _, _, close := tests.CreateLocalDevnet(t, nm)
			defer close()
			m, err := New(dnet.Client, dt.WithImportPath(filepath.Join(os.TempDir(), "imports")))
			checkErr(t, err)
			_, err = storeMultiMiner(m, dnet, nm, randomBytes(1000))
			checkErr(t, err)
		})
	}
}

func TestRetrieve(t *testing.T) {
	numMiners := []int{1} // go-fil-markets: doesn't support remembering more than 1 miner
	data := randomBytes(1000)
	for _, nm := range numMiners {
		t.Run(fmt.Sprintf("CantMiners%d", nm), func(t *testing.T) {
			dnet, addr, _, close := tests.CreateLocalDevnet(t, nm)
			defer close()
			m, err := New(dnet.Client, dt.WithImportPath(filepath.Join(os.TempDir(), "imports")))
			checkErr(t, err)

			dcid, err := storeMultiMiner(m, dnet, nm, data)
			checkErr(t, err)
			ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()

			r, err := m.Retrieve(ctx, addr.String(), dcid)
			checkErr(t, err)
			defer r.Close()
			rdata, err := ioutil.ReadAll(r)
			checkErr(t, err)
			if !bytes.Equal(data, rdata) {
				t.Fatal("retrieved data doesn't match with stored data")
			}
		})
	}
}

func TestWatchStore(t *testing.T) {
	dnet, addr, miners, close := tests.CreateLocalDevnet(t, 1)
	defer close()
	ctx := context.Background()
	m, err := New(dnet.Client, dt.WithImportPath(filepath.Join(os.TempDir(), "imports")))
	checkErr(t, err)

	cfgs := []dt.StorageDealConfig{dt.StorageDealConfig{Miner: miners[0].String(), EpochPrice: types.NewInt(40000000)}}
	_, pcids, failed, err := m.Store(ctx, addr.String(), bytes.NewReader(randomBytes(1000)), cfgs, 100)
	checkErr(t, err)
	if len(failed) > 0 {
		t.Fatalf("%d deal configurations failed", len(failed))
	}
	if len(pcids) != len(cfgs) {
		t.Fatalf("some deal cids are missing, got %d, expected %d", len(pcids), len(cfgs))
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	chDealInfo, err := m.Watch(ctx, pcids)
	checkErr(t, err)
	expectedStatePath := []api.DealState{
		api.DealUnknown,
		// api.DealAccepted, api.DealStaged, // Off-chain negotation isn't delayable at the moment, so too fast to detect
		api.DealSealing,
		api.DealComplete,
	}
	for i := 0; i < len(expectedStatePath); i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		select {
		case d := <-chDealInfo:
			if d.StateID != expectedStatePath[i] {
				t.Fatalf("proposal missed expected state %d, got %d", expectedStatePath[i], d.StateID)
			}
		case <-ctx.Done():
			t.Fatalf("waiting for next state timeout")
		}
		cancel()
	}
}

func storeMultiMiner(m *Module, dnet *ldevnet.LocalDevnet, numMiners int, data []byte) (cid.Cid, error) {
	ctx := context.Background()
	miners, err := dnet.Client.StateListMiners(ctx, nil)
	if err != nil {
		return cid.Undef, err
	}
	if len(miners) != numMiners {
		return cid.Undef, fmt.Errorf("unexpected number of miners in the network")
	}
	addr, err := dnet.Client.WalletDefaultAddress(ctx)
	if err != nil {
		return cid.Undef, err
	}

	cfgs := make([]dt.StorageDealConfig, numMiners)
	for i := 0; i < numMiners; i++ {
		cfgs[i] = dt.StorageDealConfig{
			Miner:      miners[i].String(),
			EpochPrice: types.NewInt(40000000),
		}
	}
	dcid, pcids, failed, err := m.Store(ctx, addr.String(), bytes.NewReader(data), cfgs, 100)
	if err != nil {
		return cid.Undef, fmt.Errorf("error when calling Store()")
	}
	if !dcid.Defined() {
		return cid.Undef, fmt.Errorf("data cid is undefined")
	}
	if len(failed) > 0 {
		return cid.Undef, fmt.Errorf("%d deal configurations failed", len(failed))
	}
	if len(pcids) != len(cfgs) {
		return cid.Undef, fmt.Errorf("some deal cids are missing, got %d, expected %d", len(pcids), len(cfgs))
	}
	if err := waitForDealComplete(dnet, pcids); err != nil {
		return cid.Undef, fmt.Errorf("error waiting for deal to complete: %s", err)
	}
	return dcid, nil
}

func waitForDealComplete(dnet *ldevnet.LocalDevnet, deals []cid.Cid) error {
	ctx := context.Background()
	finished := make(map[cid.Cid]struct{})
	for len(finished) != len(deals) {
		time.Sleep(time.Second)
		for _, d := range deals {
			if _, ok := finished[d]; ok {
				continue
			}

			di, err := dnet.Client.ClientGetDealInfo(ctx, d)
			if err != nil {
				return err
			}
			if di.State == api.DealComplete {
				finished[d] = struct{}{}
				continue
			}
			if di.State != api.DealUnknown &&
				di.State != api.DealAccepted &&
				di.State != api.DealStaged &&
				di.State != api.DealSealing {
				return fmt.Errorf("unexpected deal state: %s", api.DealStates[di.State])
			}
		}
	}
	return nil
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func randomBytes(size int) []byte {
	buf := make([]byte, size)
	r := rand.New(rand.NewSource(22))
	_, _ = r.Read(buf)
	return buf
}
