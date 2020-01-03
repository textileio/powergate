package api

import (
	"net"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	ma "github.com/multiformats/go-multiaddr"
	pb "github.com/textileio/filecoin/api/pb"
	"github.com/textileio/filecoin/deals"
	"github.com/textileio/filecoin/index/ask"
	"github.com/textileio/filecoin/lotus"
	"github.com/textileio/filecoin/tests"
	"github.com/textileio/filecoin/util"
	"google.golang.org/grpc"
)

var (
	log = logging.Logger("server")
)

// Server represents the configured lotus client and filecoin grpc server
type Server struct {
	ds datastore.TxnDatastore
	dm *deals.DealModule
	ai *ask.AskIndex

	rpc        *grpc.Server
	service    *service
	closeLotus func()
}

// Config specifies server settings.
type Config struct {
	LotusAddress    ma.Multiaddr
	LotusAuthToken  string
	GrpcHostAddress ma.Multiaddr
}

// NewServer starts and returns a new server with the given configuration.
func NewServer(conf Config) (*Server, error) {
	c, cls, err := lotus.New(conf.LotusAddress, conf.LotusAuthToken)
	if err != nil {
		return nil, err
	}

	// ToDo: use some other persistent data store
	ds := tests.NewTxMapDatastore()

	ai := ask.New(ds, c)
	dm := deals.New(c, ds)
	service := newService(dm, ai)

	s := &Server{
		// ToDo: Support secure connection
		rpc:        grpc.NewServer(),
		ds:         ds,
		dm:         dm,
		ai:         ai,
		service:    service,
		closeLotus: cls,
	}

	grpcAddr, err := util.TCPAddrFromMultiAddr(conf.GrpcHostAddress)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", grpcAddr)
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
	s.rpc.GracefulStop()
	s.closeLotus()
	s.ai.Close()
	if err := s.ds.Close(); err != nil {
		log.Errorf("error when closing datastore: %s", err)
	}
}
