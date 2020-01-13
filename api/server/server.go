package server

import (
	"net"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/filecoin/deals"
	dealsPb "github.com/textileio/filecoin/deals/pb"
	"github.com/textileio/filecoin/fchost"
	"github.com/textileio/filecoin/index/ask"
	"github.com/textileio/filecoin/index/miner"
	"github.com/textileio/filecoin/index/slashing"
	"github.com/textileio/filecoin/iplocation/ip2location"
	"github.com/textileio/filecoin/lotus"
	"github.com/textileio/filecoin/tests"
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
	si *slashing.SlashingIndex
	mi *miner.MinerIndex

	rpc           *grpc.Server
	dealsService  *deals.Service
	walletService *wallet.Service
	closeLotus    func()
}

// Config specifies server settings.
type Config struct {
	LotusAddress    ma.Multiaddr
	LotusAuthToken  string
	GrpcHostNetwork string
	GrpcHostAddress string
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
	dm := deals.New(ds, c)
	wm := wallet.New(c)
	fchost, err := fchost.New()
	if err != nil {
		return nil, err
	}
	if err := fchost.Bootstrap(); err != nil {
		return nil, err
	}
	// ToDo: Flags or embed
	ip2l := ip2location.New([]string{"ip2location-ip4.bin"})
	mi := miner.New(ds, c, fchost, ip2l)
	si := slashing.New(ds, c)
	dealsService := deals.NewService(dm, ai)
	walletService := wallet.NewService(wm)

	s := &Server{
		// ToDo: Support secure connection
		rpc:           grpc.NewServer(),
		ds:            ds,
		dm:            dm,
		ai:            ai,
		wm:            wm,
		mi:            mi,
		si:            si,
		dealsService:  dealsService,
		walletService: walletService,
		closeLotus:    cls,
	}

	listener, err := net.Listen(conf.GrpcHostNetwork, conf.GrpcHostAddress)
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
	if err := s.ai.Close(); err != nil {
		log.Errorf("error when closing ask index: %s", err)
	}
	if err := s.mi.Close(); err != nil {
		log.Errorf("error when closing miner index: %s", err)
	}
	if err := s.si.Close(); err != nil {
		log.Errorf("error when closing slashing index: %s", err)
	}
	if err := s.ds.Close(); err != nil {
		log.Errorf("error when closing datastore: %s", err)
	}
}
