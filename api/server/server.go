package server

import (
	"net"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/filecoin/deals"
	dealsPb "github.com/textileio/filecoin/deals/pb"
	"github.com/textileio/filecoin/index/ask"
	"github.com/textileio/filecoin/lotus"
	"github.com/textileio/filecoin/tests"
	"github.com/textileio/filecoin/util"
	"github.com/textileio/filecoin/wallet"
	walletPb "github.com/textileio/filecoin/wallet/pb"
	"google.golang.org/grpc"
)

var (
	log = logging.Logger("server")
)

// Server represents the configured lotus client and filecoin grpc server
type Server struct {
	ds datastore.TxnDatastore
	dm *deals.Module
	ai *ask.AskIndex
	wm *wallet.Module

	rpc           *grpc.Server
	dealsService  *deals.Service
	walletService *wallet.Service
	closeLotus    func()
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
	wm := wallet.New(c)
	dealsService := deals.NewService(dm, ai)

	s := &Server{
		// ToDo: Support secure connection
		rpc:           grpc.NewServer(),
		ds:            ds,
		dm:            dm,
		ai:            ai,
		dealsService:  dealsService,
		walletService: &wallet.Service{Module: wm},
		closeLotus:    cls,
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
		dealsPb.RegisterAPIServer(s.rpc, s.dealsService)
		walletPb.RegisterAPIServer(s.rpc, s.walletService)
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
