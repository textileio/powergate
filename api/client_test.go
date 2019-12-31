package api

import (
	"strings"
	"testing"

	"github.com/ipfs/go-cid"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/filecoin/deals"
	"github.com/textileio/filecoin/tests"
)

var (
	grpcHostAddr = "/ip4/127.0.0.1/tcp/50051"
)

func TestQueryAsk(t *testing.T) {
	client, done := setup(t)
	defer done()

	_, err := client.AvailableAsks(deals.Query{MaxPrice: 5})
	if err != nil {
		t.Fatalf("failed to call AvailableAsks: %v", err)
	}
}

func TestStore(t *testing.T) {
	client, done := setup(t)
	defer done()

	r := strings.NewReader("store me")
	_, _, err := client.Store("an address", r, make([]deals.DealConfig, 0), 1024)
	if err != nil {
		t.Fatalf("failed to call Store: %v", err)
	}
}

func TestWatch(t *testing.T) {
	client, done := setup(t)
	defer done()

	_, _, err := client.Watch(make([]cid.Cid, 0))
	if err != nil {
		t.Fatalf("failed to call Watch: %v", err)
	}
}

func setup(t *testing.T) (*Client, func()) {
	lotusAddr, token := tests.ClientConfigMA()
	grpcAddr, err := ma.NewMultiaddr(grpcHostAddr)
	checkErr(t, err)
	conf := Config{
		LotusAddress:    lotusAddr,
		LotusAuthToken:  token,
		GrpcHostAddress: grpcAddr,
	}
	server, err := NewServer(conf)
	checkErr(t, err)

	client, err := NewClient(grpcAddr)
	checkErr(t, err)

	return client, func() {
		client.Close()
		server.Close()
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
