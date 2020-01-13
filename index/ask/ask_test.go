package ask

import (
	"context"
	"reflect"
	"testing"

	"github.com/textileio/filecoin/lotus"
	"github.com/textileio/filecoin/tests"
)

func TestFreshBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test since we're on short mode")
	}
	addr, token := tests.ClientConfigMA()
	c, cls, err := lotus.New(addr, token)
	checkErr(t, err)
	defer cls()

	queryAskRateLim = 2000
	ai := &AskIndex{ctx: context.Background(), c: c}
	err = ai.updateIndex()
	checkErr(t, err)
	if len(ai.Get().Miners) == 0 {
		t.Fatalf("current asks can't be empty")
	}
}

func TestQueryAsk(t *testing.T) {
	t.Parallel()
	dm := AskIndex{}
	dm.queryCache = []*StorageAsk{
		{Price: uint64(20), MinPieceSize: 128, Miner: "t01"},
		{Price: uint64(30), MinPieceSize: 64, Miner: "t02"},
		{Price: uint64(40), MinPieceSize: 256, Miner: "t03"},
		{Price: uint64(50), MinPieceSize: 16, Miner: "t04"},
	}

	facr := []StorageAsk{
		{Price: 20, MinPieceSize: 128, Miner: "t01"},
		{Price: 30, MinPieceSize: 64, Miner: "t02"},
		{Price: 40, MinPieceSize: 256, Miner: "t03"},
		{Price: 50, MinPieceSize: 16, Miner: "t04"},
	}

	tests := []struct {
		name   string
		q      Query
		expect []StorageAsk
	}{
		{name: "All", q: Query{}, expect: facr},
		{name: "LeqPrice35", q: Query{MaxPrice: 35}, expect: []StorageAsk{facr[0], facr[1]}},
		{name: "LeqPrice50", q: Query{MaxPrice: 50}, expect: facr},
		{name: "LeqPrice40Piece96", q: Query{MaxPrice: 35, PieceSize: 96}, expect: []StorageAsk{facr[1]}},
		{name: "AllLimit2Offset1", q: Query{Limit: 2, Offset: 1}, expect: []StorageAsk{facr[1], facr[2]}},
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

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
