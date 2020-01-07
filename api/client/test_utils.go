package client

import (
	"context"
	"testing"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/filecoin/api/server"
	"github.com/textileio/filecoin/tests"
	"github.com/textileio/filecoin/util"
	"google.golang.org/grpc"
)

var (
	grpcHostAddr = "/ip4/127.0.0.1/tcp/50051"
	ctx          = context.Background()
)

func getHostMultiaddress(t *testing.T) ma.Multiaddr {
	grpcAddr, err := ma.NewMultiaddr(grpcHostAddr)
	checkErr(t, err)
	return grpcAddr
}

func setupServer(t *testing.T) func() {
	lotusAddr, token := tests.ClientConfigMA()
	conf := server.Config{
		LotusAddress:    lotusAddr,
		LotusAuthToken:  token,
		GrpcHostAddress: getHostMultiaddress(t),
	}
	server, err := server.NewServer(conf)
	checkErr(t, err)

	return func() {
		server.Close()
	}
}

func setupConnection(t *testing.T) (*grpc.ClientConn, func()) {
	addr, err := util.TCPAddrFromMultiAddr(getHostMultiaddress(t))
	checkErr(t, err)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	checkErr(t, err)
	return conn, func() {
		conn.Close()
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
