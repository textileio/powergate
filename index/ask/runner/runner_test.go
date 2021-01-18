package runner

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/v2/index/ask"
	"github.com/textileio/powergate/v2/tests"
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestFreshBuild(t *testing.T) {
	// Skipped since current interop devnet returns
	// "must resolve ID addresses before using them to verify a signature"
	// when querying Asks.
	t.SkipNow()
	ctx := context.Background()
	clientBuilder, _, miners := tests.CreateLocalDevnet(t, 1)
	c, cls, err := clientBuilder(ctx)
	require.NoError(t, err)
	defer cls()

	index, _, err := generateIndex(ctx, c, 1, time.Second*5)
	require.NoError(t, err)

	// We should have storage info about every miner in devnet
	for _, m := range miners {
		info, ok := index.Storage[m.String()]
		if !ok {
			t.Fatalf("missing storage ask info for miner %s", m.String())
		}
		if info.Miner != m.String() || info.Price == 0 ||
			info.MinPieceSize == 0 || info.Expiry == 0 {
			t.Fatalf("invalid storage state for miner %s: %v", m.String(), info)
		}
	}
	if index.StorageMedianPrice == 0 {
		t.Fatalf("median storage price should be greater than zero")
	}
}

func TestQueryAsk(t *testing.T) {
	t.Parallel()
	dm := Runner{}
	dm.orderedAsks = []*ask.StorageAsk{
		{Price: uint64(20), MinPieceSize: 128, Miner: "t01"},
		{Price: uint64(30), MinPieceSize: 64, MaxPieceSize: 128, Miner: "t02"},
		{Price: uint64(40), MinPieceSize: 256, Miner: "t03"},
		{Price: uint64(50), MinPieceSize: 16, Miner: "t04"},
	}

	facr := []ask.StorageAsk{
		{Price: 20, MinPieceSize: 128, Miner: "t01"},
		{Price: 30, MinPieceSize: 64, MaxPieceSize: 128, Miner: "t02"},
		{Price: 40, MinPieceSize: 256, Miner: "t03"},
		{Price: 50, MinPieceSize: 16, Miner: "t04"},
	}

	tests := []struct {
		name   string
		q      ask.Query
		expect []ask.StorageAsk
	}{
		{name: "All", q: ask.Query{}, expect: facr},
		{name: "LeqPrice35", q: ask.Query{MaxPrice: 35}, expect: []ask.StorageAsk{facr[0], facr[1]}},
		{name: "LeqPrice50", q: ask.Query{MaxPrice: 50}, expect: facr},
		{name: "LeqPrice40Piece96", q: ask.Query{MaxPrice: 35, PieceSize: 96}, expect: []ask.StorageAsk{facr[1]}},
		{name: "AllLimit2Offset1", q: ask.Query{Limit: 2, Offset: 1}, expect: []ask.StorageAsk{facr[1], facr[2]}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := dm.Query(tt.q)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tt.expect, got) {
				t.Fatalf("expected %v, got %v", tt.expect, got)
			}
		})
	}
}
