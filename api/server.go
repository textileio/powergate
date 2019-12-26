package api

import (
	"context"
	"net/http"

	"github.com/ipfs/go-datastore"
	"github.com/textileio/filecoin/client"
	"github.com/textileio/filecoin/deals"
	"google.golang.org/grpc"
)

// Server represents the configured filecoin grpc server
type Server struct {
	rpc     *grpc.Server
	proxy   *http.Server
	service *service

	ctx    context.Context
	cancel context.CancelFunc
}

// Config specifies server settings.
type Config struct {
	LotusAddress   string
	LotusAuthToken string
	// RepoPath  string
	// Addr      ma.Multiaddr
	// ProxyAddr ma.Multiaddr
	// Debug     bool
}

// NewServer starts and returns a new server with the given threadservice.
// The threadservice is *not* managed by the server.
func NewServer(ctx context.Context, conf Config) (*Server, error) {
	c, cls, err := client.New(conf.LotusAddress, conf.LotusAuthToken)
	if err != nil {
		panic(err)
	}
	defer cls()
	dm := deals.New(c, datastore.NewMapDatastore())

	ctx, cancel := context.WithCancel(ctx)
	s := &Server{
		rpc:     grpc.NewServer(),
		service: &service{dealModule: dm},
		ctx:     ctx,
		cancel:  cancel,
	}

	return nil, nil
}
