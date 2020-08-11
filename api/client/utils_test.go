package client

import (
	"context"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/api/server"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
	"google.golang.org/grpc"
)

var (
	grpcHostNetwork     = "tcp"
	grpcHostAddress     = "/ip4/127.0.0.1/tcp/5002"
	grpcWebProxyAddress = "127.0.0.1:6002"
	gatewayHostAddr     = "0.0.0.0:7000"
	ctx                 = context.Background()
)

func setupServer(t *testing.T) func() {
	repoPath, err := ioutil.TempDir("/tmp/powergate", ".powergate-*")
	if err != nil {
		t.Fatal(err)
	}

	dipfs, cls := tests.LaunchIPFSDocker(t)
	t.Cleanup(func() { cls() })

	ipfsAddrStr := "/ip4/127.0.0.1/tcp/" + dipfs.GetPort("5001/tcp")
	ipfsAddr := util.MustParseAddr(ipfsAddrStr)

	devnet := tests.LaunchDevnetDocker(t, 1, ipfsAddrStr, false)
	devnetAddr := util.MustParseAddr("/ip4/127.0.0.1/tcp/" + devnet.GetPort("7777/tcp"))

	grpcMaddr := util.MustParseAddr(grpcHostAddress)
	conf := server.Config{
		WalletInitialFunds:  *big.NewInt(int64(4000000000)),
		IpfsAPIAddr:         ipfsAddr,
		LotusAddress:        devnetAddr,
		LotusAuthToken:      "",
		LotusMasterAddr:     "",
		Devnet:              true,
		GrpcHostNetwork:     grpcHostNetwork,
		GrpcHostAddress:     grpcMaddr,
		GrpcWebProxyAddress: grpcWebProxyAddress,
		RepoPath:            repoPath,
		GatewayHostAddr:     gatewayHostAddr,
		MaxMindDBFolder:     "../../iplocation/maxmind",
	}
	server, err := server.NewServer(conf)
	checkErr(t, err)

	return func() {
		server.Close()
	}
}

func setupConnection(t *testing.T) (*grpc.ClientConn, func()) {
	auth := TokenAuth{}
	ma, err := multiaddr.NewMultiaddr(grpcHostAddress)
	checkErr(t, err)
	addr, err := util.TCPAddrFromMultiAddr(ma)
	checkErr(t, err)
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithPerRPCCredentials(auth))
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
