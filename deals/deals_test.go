package deals

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/fil-tools/ldevnet"
	"github.com/textileio/fil-tools/tests"
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestStore(t *testing.T) {
	numMiners := []int{1}
	for _, nm := range numMiners {
		t.Run(fmt.Sprintf("CantMiners%d", nm), storeMultiMiner(t, nm))
	}
}

func TestWatchStore(t *testing.T) {
	dnet, addr, miners, close := tests.CreateLocalDevnet(t, 1)
	defer close()
	ctx := context.Background()
	m, err := New(dnet.Client, WithImportPath(filepath.Join(os.TempDir(), "imports")))
	checkErr(t, err)

	cfgs := []DealConfig{DealConfig{Miner: miners[0].String(), EpochPrice: types.NewInt(40000000)}}
	cids, failed, err := m.Store(ctx, addr.String(), bytes.NewReader(randomBytes(1000)), cfgs, 100)
	checkErr(t, err)
	if len(failed) > 0 {
		t.Fatalf("%d deal configurations failed", len(failed))
	}
	if len(cids) != len(cfgs) {
		t.Fatalf("some deal cids are missing, got %d, expected %d", len(cids), len(cfgs))
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	chDealInfo, err := m.Watch(ctx, cids)
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

func storeMultiMiner(t *testing.T, numMiners int) func(t *testing.T) {
	return func(t *testing.T) {
		dnet, addr, miners, close := tests.CreateLocalDevnet(t, numMiners)
		defer close()
		ctx := context.Background()
		m, err := New(dnet.Client, WithImportPath(filepath.Join(os.TempDir(), "imports")))
		checkErr(t, err)

		cfgs := make([]DealConfig, numMiners)
		for i := 0; i < numMiners; i++ {
			cfgs[i] = DealConfig{
				Miner:      miners[i].String(),
				EpochPrice: types.NewInt(40000000),
			}
		}
		cids, failed, err := m.Store(ctx, addr.String(), bytes.NewReader(randomBytes(1000)), cfgs, 100)
		checkErr(t, err)
		if len(failed) > 0 {
			t.Fatalf("%d deal configurations failed", len(failed))
		}
		if len(cids) != len(cfgs) {
			t.Fatalf("some deal cids are missing, got %d, expected %d", len(cids), len(cfgs))
		}
		if err := waitForDealComplete(dnet, cids); err != nil {
			t.Fatal(err)
		}
	}
}

func waitForDealComplete(dnet *ldevnet.LocalDevnet, deals []cid.Cid) error {
	ctx := context.Background()
	for {
		time.Sleep(time.Second)
		for _, d := range deals {
			di, err := dnet.Client.ClientGetDealInfo(ctx, d)
			if err != nil {
				return err
			}
			if di.State == api.DealComplete {
				return nil
			}
			if di.State != api.DealUnknown &&
				di.State != api.DealAccepted &&
				di.State != api.DealStaged &&
				di.State != api.DealSealing {
				return fmt.Errorf("unexpected deal state: %s", err)
			}
			fmt.Printf("Deal %s in state %d\n", d, di.State)
		}
	}
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
