package api

import (
	"context"
	"net"
	"net/http"

	"github.com/ipfs/go-datastore"
	pb "github.com/textileio/filecoin/api/pb"
	"github.com/textileio/filecoin/client"
	"github.com/textileio/filecoin/deals"
	"google.golang.org/grpc"
)

// Server represents the configured lotus client and filecoin grpc server
type Server struct {
	rpc     *grpc.Server
	proxy   *http.Server
	service *service

	ctx    context.Context
	cancel context.CancelFunc
}

// Config specifies server settings.
type Config struct {
	LotusAddress    string
	LotusAuthToken  string
	GrpcHostAddress string
}

// NewServer starts and returns a new server with the given configuration.
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

	listener, err := net.Listen("tcp", conf.GrpcHostAddress)
	if err != nil {
		return nil, err
	}
	go func() {
		pb.RegisterAPIServer(s.rpc, s.service)
		s.rpc.Serve(listener)
	}()

	return s, nil
}
