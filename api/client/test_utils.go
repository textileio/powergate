package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/api/server"
	"github.com/textileio/powergate/tests"
	"google.golang.org/grpc"
)

var (
	grpcHostNetwork     = "tcp"
	grpcHostAddress     = "127.0.0.1:5002"
	grpcWebProxyAddress = "127.0.0.1:6002"
	ctx                 = context.Background()
)

func setupServer(t *testing.T) func() {
	lotusAddr, token := tests.ClientConfigMA()
	conf := server.Config{
		LotusAddress:   lotusAddr,
		LotusAuthToken: token,
		// ToDo: Support secure gRPC connection
		GrpcHostNetwork:     grpcHostNetwork,
		GrpcHostAddress:     grpcHostAddress,
		GrpcWebProxyAddress: grpcWebProxyAddress,
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
		require.NoError(t, conn.Close())
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping since is a short test run")
	}
}
