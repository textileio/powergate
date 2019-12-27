package api

import (
	"net"

	"github.com/ipfs/go-datastore"
	pb "github.com/textileio/filecoin/api/pb"
	"github.com/textileio/filecoin/deals"
	"github.com/textileio/filecoin/lotus"
	"google.golang.org/grpc"
)

// Server represents the configured lotus client and filecoin grpc server
type Server struct {
	rpc        *grpc.Server
	service    *service
	closeLotus func()
}

// Config specifies server settings.
type Config struct {
	LotusAddress    string
	LotusAuthToken  string
	GrpcHostAddress string
}

// NewServer starts and returns a new server with the given configuration.
func NewServer(conf Config) (*Server, error) {
	c, cls, err := lotus.New(conf.LotusAddress, conf.LotusAuthToken)
	if err != nil {
		panic(err)
	}

	// ToDo: use some other persistent data store
	dm := deals.New(c, datastore.NewMapDatastore())

	s := &Server{
		rpc:        grpc.NewServer(),
		service:    &service{dealModule: dm},
		closeLotus: cls,
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

// Close shuts down the server
func (s *Server) Close() {
	s.service.dealModule.Close()
	s.closeLotus()
	s.rpc.Stop()
}
