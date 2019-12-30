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
	server       *Server
	client       *Client
	grpcHostAddr = "/ip4/127.0.0.1/tcp/50051"
)

// func TestMain(m *testing.M) {
// 	server = makeServer()
// 	client = makeClient()
// 	exitVal := m.Run()
// 	shutdown()
// 	os.Exit(exitVal)
// }

func TestQueryAsk(t *testing.T) {
	_, err := client.AvailableAsks(deals.Query{MaxPrice: 5})
	if err != nil {
		t.Fatalf("failed to call AvailableAsks: %v", err)
	}
}

func TestStore(t *testing.T) {
	r := strings.NewReader("store me")
	_, _, err := client.Store("an address", r, make([]deals.DealConfig, 0), 1024)
	if err != nil {
		t.Fatalf("failed to call Store: %v", err)
	}
}

func TestWatch(t *testing.T) {
	_, _, err := client.Watch(make([]cid.Cid, 0))
	if err != nil {
		t.Fatalf("failed to call Watch: %v", err)
	}
}

func makeServer() *Server {
	addr, token := tests.ClientConfigMA()
	addr, err := ma.NewMultiaddr(grpcHostAddr)
	if err != nil {
		panic(err)
	}
	conf := Config{
		LotusAddress:    addr,
		LotusAuthToken:  token,
		GrpcHostAddress: addr,
	}
	server, err := NewServer(conf)
	if err != nil {
		panic(err)
	}
	return server
}

func makeClient() *Client {
	addr, err := ma.NewMultiaddr(grpcHostAddr)
	if err != nil {
		panic(err)
	}
	client, err := NewClient(addr)
	if err != nil {
		panic(err)
	}
	return client
}

func shutdown() {
	client.Close()
	server.Close()
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
