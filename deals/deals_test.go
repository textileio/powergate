package deals

import (
	"bytes"
	"context"
	"math/big"
	"math/rand"
	"os"
	"testing"

	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/filecoin/ldevnet"
	"github.com/textileio/filecoin/tests"
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestStoreSimple(t *testing.T) {
	t.Parallel()
	dnet, err := ldevnet.New(t, 1)
	checkErr(t, err)
	defer dnet.Close()

	m := New(tests.NewTxMapDatastore(), dnet.Client)

	ctx := context.Background()
	addr, err := dnet.Client.WalletDefaultAddress(ctx)
	if err != nil {
		t.Fatal(err)
	}
	b := randomBytes(1000)
	dcfgs := []DealConfig{DealConfig{Miner: dnet.Miner.String(), EpochPrice: bigExp(1, 8)}}
	cids, failed, err := m.Store(ctx, addr.String(), bytes.NewReader(b), dcfgs, 1000)
	checkErr(t, err)
	if len(failed) > 0 {
		t.Fatalf("%d deal configurations failed", len(failed))
	}
	if len(cids) != len(dcfgs) {
		t.Fatalf("some deal cids are missing, got %d, expected %d", len(cids), len(dcfgs))
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

func bigExp(base, exp int64) types.BigInt {
	return types.BigInt{Int: big.NewInt(0).Exp(big.NewInt(base), big.NewInt(exp), nil)}
}
