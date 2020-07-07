package module

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
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

const (
	tmpDir = "/tmp/powergate/dealstest"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Second
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
			m, err := New(tests.NewTxMapDatastore(), client, deals.WithImportPath(filepath.Join(tmpDir, "imports")))
			checkErr(t, err)
			_, pcids, err := storeMultiMiner(m, client, nm, randomBytes(600))
			checkErr(t, err)
			pending, err := m.ListStorageDealRecords(ffs.WithOnlyPending(true))
			require.Nil(t, err)
			require.Len(t, pending, nm)
			final, err := m.ListStorageDealRecords()
			require.Nil(t, err)
			require.Empty(t, final)
			err = waitForDealComplete(client, pcids)
			checkErr(t, err)
			time.Sleep(util.AvgBlockTime)
			pending, err = m.ListStorageDealRecords(ffs.WithOnlyPending(true))
			require.Nil(t, err)
			require.Empty(t, pending)
			final, err = m.ListStorageDealRecords()
			require.Nil(t, err)
			require.Len(t, final, nm)
		})
	}
}
func TestRetrieve(t *testing.T) {
	numMiners := []int{1} // go-fil-markets: doesn't support remembering more than 1 miner
	data := randomBytes(600)
	for _, nm := range numMiners {
		t.Run(fmt.Sprintf("CantMiners%d", nm), func(t *testing.T) {
			client, addr, _ := tests.CreateLocalDevnet(t, nm)
			m, err := New(tests.NewTxMapDatastore(), client, deals.WithImportPath(filepath.Join(tmpDir, "imports")))
			checkErr(t, err)

			dcid, pcids, err := storeMultiMiner(m, client, nm, data)
			checkErr(t, err)
			err = waitForDealComplete(client, pcids)
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
			retrievals, err := m.ListRetrievalDealRecords()
			require.Nil(t, err)
			require.Len(t, retrievals, 1)
		})
	}
}

func storeMultiMiner(m *Module, client *apistruct.FullNodeStruct, numMiners int, data []byte) (cid.Cid, []cid.Cid, error) {
	ctx := context.Background()
	miners, err := client.StateListMiners(ctx, types.EmptyTSK)
	if err != nil {
		return cid.Undef, nil, err
	}
	if len(miners) != numMiners {
		return cid.Undef, nil, fmt.Errorf("unexpected number of miners in the network")
	}
	addr, err := client.WalletDefaultAddress(ctx)
	if err != nil {
		return cid.Undef, nil, err
	}

	cfgs := make([]deals.StorageDealConfig, numMiners)
	for i := 0; i < numMiners; i++ {
		cfgs[i] = deals.StorageDealConfig{
			Miner:      miners[i].String(),
			EpochPrice: 500000000,
		}
	}
	dataCid, size, err := m.Import(ctx, bytes.NewReader(data), false)
	if err != nil {
		return cid.Undef, nil, err
	}
	if !dataCid.Defined() {
		return cid.Undef, nil, fmt.Errorf("data cid is undefined")
	}
	srs, err := m.Store(ctx, addr.String(), dataCid, 2*uint64(size), cfgs, 1000)
	if err != nil {
		return cid.Undef, nil, fmt.Errorf("error when calling Store(): %s", err)
	}

	var pcids []cid.Cid
	for _, r := range srs {
		if !r.Success {
			return cid.Undef, nil, fmt.Errorf("%v deal configurations failed", r)
		}
		pcids = append(pcids, r.ProposalCid)
	}
	if len(srs) != len(cfgs) {
		return cid.Undef, nil, fmt.Errorf("some deal cids are missing, got %d, expected %d", len(srs), len(cfgs))
	}
	return dataCid, pcids, nil
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
			switch di.State {
			case
				storagemarket.StorageDealUnknown,
				storagemarket.StorageDealWaitingForDataRequest,
				storagemarket.StorageDealProposalAccepted,
				storagemarket.StorageDealStaged,
				storagemarket.StorageDealValidating,
				storagemarket.StorageDealClientFunding,
				storagemarket.StorageDealPublish,
				storagemarket.StorageDealPublishing,
				storagemarket.StorageDealSealing:
			default:

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
