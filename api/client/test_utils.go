package client

import (
	"context"
	"testing"

	"github.com/textileio/filecoin/api/server"
	"github.com/textileio/filecoin/tests"
	"google.golang.org/grpc"
)

var (
	grpcHostNetwork = "tcp"
	grpcHostAddress = "127.0.0.1:50051"
	ctx             = context.Background()
)

func setupServer(t *testing.T) func() {
	lotusAddr, token := tests.ClientConfigMA()
	conf := server.Config{
		LotusAddress:    lotusAddr,
		LotusAuthToken:  token,
		GrpcHostNetwork: grpcHostNetwork,
		GrpcHostAddress: grpcHostAddress,
	}
	server, err := server.NewServer(conf)
	checkErr(t, err)

	return func() {
		server.Close()
	}
}

func setupConnection(t *testing.T) (*grpc.ClientConn, func()) {
	conn, err := grpc.Dial(grpcHostAddress, grpc.WithInsecure())
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
