package api

import (
	"net"

	"github.com/ipfs/go-datastore"
	ma "github.com/multiformats/go-multiaddr"
	dealsPb "github.com/textileio/filecoin/deals/pb"
	walletPb "github.com/textileio/filecoin/wallet/pb"
	"github.com/textileio/filecoin/deals"
	"github.com/textileio/filecoin/lotus"
	"github.com/textileio/filecoin/util"
	"github.com/textileio/filecoin/wallet"
	"google.golang.org/grpc"
)

// Server represents the configured lotus client and filecoin grpc server
type Server struct {
	rpc          *grpc.Server
	dealsService *deals.Service
	walletService *wallet.Service
	closeLotus   func()
}

// Config specifies server settings.
type Config struct {
	LotusAddress    ma.Multiaddr
	LotusAuthToken  string
	GrpcHostAddress ma.Multiaddr
}

// NewServer starts and returns a new server with the given configuration.
func NewServer(conf Config) (*Server, error) {
	lotusAddr, err := util.TCPAddrFromMultiAddr(conf.LotusAddress)
	if err != nil {
		return nil, err
	}
	c, cls, err := lotus.New(lotusAddr, conf.LotusAuthToken)
	if err != nil {
		return nil, err
	}

	// ToDo: use some other persistent data store
	dm := deals.New(c, datastore.NewMapDatastore())
	wm := wallet.New(c)

	s := &Server{
		// ToDo: Support secure connection
		rpc: grpc.NewServer(),
		dealsService: &deals.Service{Module: dm},
		walletService: &wallet.Service{Module: wm},
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
		dealsPb.RegisterAPIServer(s.rpc, s.dealsService)
		walletPb.RegisterAPIServer(s.rpc, s.walletService)
		s.rpc.Serve(listener)
	}()

	return s, nil
}

// Close shuts down the server
func (s *Server) Close() {
	s.dealsService.Module.Close()
	s.walletService.Module.Close()
	s.closeLotus()
	s.rpc.Stop()
}
