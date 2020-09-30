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
	t.Parallel()
	numMiners := []int{1, 2}
	for _, nm := range numMiners {
		nm := nm
		t.Run(fmt.Sprintf("CantMiners%d", nm), func(t *testing.T) {
			t.Parallel()
			tests.RunFlaky(t, func(t *tests.FlakyT) {
				clientBuilder, addr, _ := tests.CreateLocalDevnet(t, nm)
				m, err := New(tests.NewTxMapDatastore(), clientBuilder, util.AvgBlockTime, deals.WithImportPath(filepath.Join(tmpDir, "imports")))
				require.NoError(t, err)
				c, cls, err := clientBuilder()
				require.NoError(t, err)
				defer cls()
				cid, pcids, err := storeMultiMiner(m, c, nm, randomBytes(600))
				require.NoError(t, err)
				pending, err := m.ListStorageDealRecords(
					deals.WithIncludePending(true),
					deals.WithDataCids(util.CidToString(cid)),
					deals.WithAscending(true),
					deals.WithFromAddrs(addr.String()),
				)
				require.NoError(t, err)
				require.Len(t, pending, nm)
				final, err := m.ListStorageDealRecords(deals.WithIncludeFinal(true))
				require.NoError(t, err)
				require.Empty(t, final)
				err = waitForDealComplete(c, pcids)
				require.NoError(t, err)
				time.Sleep(util.AvgBlockTime)
				pending, err = m.ListStorageDealRecords(deals.WithIncludePending(true))
				require.NoError(t, err)
				require.Empty(t, pending)
				final, err = m.ListStorageDealRecords(
					deals.WithIncludeFinal(true),
					deals.WithDataCids(util.CidToString(cid)),
					deals.WithAscending(true),
					deals.WithFromAddrs(addr.String()),
				)
				require.NoError(t, err)
				require.Len(t, final, nm)
			})
		})
	}
}

func TestRetrieve(t *testing.T) {
	t.Parallel()
	numMiners := []int{1} // go-fil-markets: doesn't support remembering more than 1 miner
	for _, nm := range numMiners {
		nm := nm
		t.Run(fmt.Sprintf("CantMiners%d", nm), func(t *testing.T) {
			t.Parallel()
			tests.RunFlaky(t, func(t *tests.FlakyT) {
				data := randomBytes(600)
				clientBuilder, addr, _ := tests.CreateLocalDevnet(t, nm)
				m, err := New(tests.NewTxMapDatastore(), clientBuilder, util.AvgBlockTime, deals.WithImportPath(filepath.Join(tmpDir, "imports")))
				require.NoError(t, err)
				c, cls, err := clientBuilder()
				require.NoError(t, err)
				defer cls()

				dcid, pcids, err := storeMultiMiner(m, c, nm, data)
				require.NoError(t, err)

				err = waitForDealComplete(c, pcids)
				require.NoError(t, err)
				ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
				defer cancel()

				miner, r, err := m.Retrieve(ctx, addr.String(), dcid, nil, []string{"t01000"}, false)
				require.NoError(t, err)
				require.NotEmpty(t, miner)
				defer func() {
					require.NoError(t, r.Close())
				}()
				rdata, err := ioutil.ReadAll(r)
				require.NoError(t, err)
				require.True(t, bytes.Equal(data, rdata), "retrieved data doesn't match with stored data")
				retrievals, err := m.ListRetrievalDealRecords()
				require.NoError(t, err)
				require.Len(t, retrievals, 1)
			})
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
	srs, err := m.Store(ctx, addr.String(), dataCid, 2*uint64(size), cfgs, util.MinDealDuration)
	if err != nil {
		return cid.Undef, nil, fmt.Errorf("calling Store(): %s", err)
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
				storagemarket.StorageDealWaitingForData,
				storagemarket.StorageDealProposalAccepted,
				storagemarket.StorageDealStaged,
				storagemarket.StorageDealValidating,
				storagemarket.StorageDealTransferring,
				storagemarket.StorageDealFundsEnsured,
				storagemarket.StorageDealCheckForAcceptance,
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

func randomBytes(size int) []byte {
	r := rand.New(rand.NewSource(22))
	return randomBytesFromSource(size, r)
}

func randomBytesFromSource(size int, r *rand.Rand) []byte {
	buf := make([]byte, size)
	_, _ = r.Read(buf)
	return buf
}
