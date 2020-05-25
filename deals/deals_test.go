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

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/tests"
)

const (
	tmpDir = "/tmp/powergate/dealstest"
)

func TestMain(m *testing.M) {
	if err := os.RemoveAll(tmpDir); err != nil {
		panic(err)
	}
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
			panic(err)
		}
	}
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestStore(t *testing.T) {
	numMiners := []int{1, 2}
	for _, nm := range numMiners {
		t.Run(fmt.Sprintf("CantMiners%d", nm), func(t *testing.T) {
			client, _, _ := tests.CreateLocalDevnet(t, nm)
			m, err := New(client, WithImportPath(filepath.Join(tmpDir, "imports")))
			checkErr(t, err)
			_, err = storeMultiMiner(m, client, nm, randomBytes(600))
			checkErr(t, err)
		})
	}
}
func TestRetrieve(t *testing.T) {
	numMiners := []int{1} // go-fil-markets: doesn't support remembering more than 1 miner
	data := randomBytes(600)
	for _, nm := range numMiners {
		t.Run(fmt.Sprintf("CantMiners%d", nm), func(t *testing.T) {
			client, addr, _ := tests.CreateLocalDevnet(t, nm)
			m, err := New(client, WithImportPath(filepath.Join(tmpDir, "imports")))
			checkErr(t, err)

			dcid, err := storeMultiMiner(m, client, nm, data)
			checkErr(t, err)
			ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()

			r, err := m.Retrieve(ctx, addr.String(), dcid, false)
			checkErr(t, err)
			defer func() {
				require.NoError(t, r.Close())
			}()
			rdata, err := ioutil.ReadAll(r)
			checkErr(t, err)
			if !bytes.Equal(data, rdata) {
				t.Fatal("retrieved data doesn't match with stored data")
			}
		})
	}
}

func storeMultiMiner(m *Module, client *apistruct.FullNodeStruct, numMiners int, data []byte) (cid.Cid, error) {
	ctx := context.Background()
	miners, err := client.StateListMiners(ctx, types.EmptyTSK)
	if err != nil {
		return cid.Undef, err
	}
	if len(miners) != numMiners {
		return cid.Undef, fmt.Errorf("unexpected number of miners in the network")
	}
	addr, err := client.WalletDefaultAddress(ctx)
	if err != nil {
		return cid.Undef, err
	}

	cfgs := make([]StorageDealConfig, numMiners)
	for i := 0; i < numMiners; i++ {
		cfgs[i] = StorageDealConfig{
			Miner:      miners[i].String(),
			EpochPrice: 500000000,
		}
	}
	dcid, srs, err := m.Store(ctx, addr.String(), bytes.NewReader(data), cfgs, 1000, false)
	if err != nil {
		return cid.Undef, fmt.Errorf("error when calling Store(): %s", err)
	}
	if !dcid.Defined() {
		return cid.Undef, fmt.Errorf("data cid is undefined")
	}
	var pcids []cid.Cid
	for _, r := range srs {
		if !r.Success {
			return cid.Undef, fmt.Errorf("%v deal configurations failed", r)
		}
		pcids = append(pcids, r.ProposalCid)
	}
	if len(srs) != len(cfgs) {
		return cid.Undef, fmt.Errorf("some deal cids are missing, got %d, expected %d", len(srs), len(cfgs))
	}
	if err := waitForDealComplete(client, pcids); err != nil {
		return cid.Undef, fmt.Errorf("error waiting for deal to complete: %s", err)
	}
	return dcid, nil
}

func waitForDealComplete(client *apistruct.FullNodeStruct, deals []cid.Cid) error {
	ctx := context.Background()
	finished := make(map[cid.Cid]struct{})
	for len(finished) != len(deals) {
		time.Sleep(time.Second)
		for _, d := range deals {
			if _, ok := finished[d]; ok {
				continue
			}

			di, err := client.ClientGetDealInfo(ctx, d)
			if err != nil {
				return err
			}
			if di.State == storagemarket.StorageDealActive {
				finished[d] = struct{}{}
				continue
			}
			if di.State != storagemarket.StorageDealUnknown &&
				di.State != storagemarket.StorageDealProposalAccepted &&
				di.State != storagemarket.StorageDealStaged &&
				di.State != storagemarket.StorageDealValidating &&
				di.State != storagemarket.StorageDealClientFunding &&
				di.State != storagemarket.StorageDealPublish &&
				di.State != storagemarket.StorageDealPublishing &&
				di.State != storagemarket.StorageDealSealing {
				return fmt.Errorf("unexpected deal state: %s", storagemarket.DealStates[di.State])
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
	r := rand.New(rand.NewSource(22))
	return randomBytesFromSource(size, r)
}

func randomBytesFromSource(size int, r *rand.Rand) []byte {
	buf := make([]byte, size)
	_, _ = r.Read(buf)
	return buf

}
