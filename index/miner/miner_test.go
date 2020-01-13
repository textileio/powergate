package miner

import (
	"testing"
	"time"

	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/filecoin/iplocation"
	"github.com/textileio/filecoin/lotus"
	"github.com/textileio/filecoin/tests"
)

func TestFullRefresh(t *testing.T) {
	t.Parallel()
	addr, token := tests.ClientConfigMA()
	c, cls, err := lotus.New(addr, token)
	checkErr(t, err)
	defer cls()

	mi := New(tests.NewTxMapDatastore(), c, nil, &LRMock{})

	select {
	case <-time.After(time.Second * 30):
		t.Fatal("timeout waiting for miner index full refresh")
	case <-mi.Listen():
	}
	info := mi.Get()
	if info.LastUpdated == 0 || len(info.Miners) == 0 {
		t.Fatalf("miner info state is invalid")
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

type LRMock struct {
}

func (lr *LRMock) Resolve(mas []multiaddr.Multiaddr) (iplocation.Location, error) {
	return iplocation.Location{
		Country:   "USA",
		Latitude:  0.1,
		Longitude: 0.1,
	}, nil
}
