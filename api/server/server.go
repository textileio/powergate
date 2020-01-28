package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger"
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
	txndstr "github.com/textileio/filecoin/txndstransform"
	"github.com/textileio/filecoin/wallet"
	walletPb "github.com/textileio/filecoin/wallet/pb"
	"google.golang.org/grpc"
)

const (
	datastoreFolderName = "datastore"
)

var (
	log = logging.Logger("server")
)

// Server represents the configured lotus client and filecoin grpc server
type Server struct {
	ds   datastore.TxnDatastore
	dm   *deals.Module
	ai   *ask.AskIndex
	wm   *wallet.Module
	si   *slashing.SlashingIndex
	mi   *miner.MinerIndex
	ip2l *ip2location.IP2Location

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
	RepoPath        string
}

// NewServer starts and returns a new server with the given configuration.
func NewServer(conf Config) (*Server, error) {
	c, cls, err := lotus.New(conf.LotusAddress, conf.LotusAuthToken)
	if err != nil {
		return nil, err
	}

	fchost, err := fchost.New()
	if err != nil {
		return nil, fmt.Errorf("error when creating filecoin host: %s", err)
	}
	if err := fchost.Bootstrap(); err != nil {
		return nil, fmt.Errorf("error when bootstrapping filecoin host: %s", err)
	}

	path := filepath.Join(conf.RepoPath, datastoreFolderName)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("error when creating repo folder: %s", err)
	}
	ds, err := badger.NewDatastore(path, &badger.DefaultOptions)
	if err != nil {
		return nil, fmt.Errorf("error when opening datastore on repo: %s", err)
	}
	dm := deals.New(txndstr.Wrap(ds, "dealmodule"), c)

	ip2l := ip2location.New([]string{"./ip2location-ip4.bin"})
	mi, err := miner.New(txndstr.Wrap(ds, "index/miner"), c, fchost, ip2l)
	if err != nil {
		return nil, fmt.Errorf("error when creating miner index: %s", err)
	}

	si, err := slashing.New(txndstr.Wrap(ds, "index/slashing"), c)
	if err != nil {
		return nil, fmt.Errorf("error when creating slashing index: %s", err)
	}

	ai, err := ask.New(txndstr.Wrap(ds, "index/ask"), c)
	if err != nil {
		return nil, fmt.Errorf("error when creating ask index: %s", err)
	}
	dealsService := deals.NewService(dm, ai)

	wm := wallet.New(c)
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
		ip2l:          ip2l,
		dealsService:  dealsService,
		walletService: walletService,
		closeLotus:    cls,
	}

	listener, err := net.Listen(conf.GrpcHostNetwork, conf.GrpcHostAddress)
	if err != nil {
		return nil, fmt.Errorf("error when listening to grpc: %s", err)
	}
	go func() {
		dealsPb.RegisterAPIServer(s.rpc, s.dealsService)
		walletPb.RegisterAPIServer(s.rpc, s.walletService)
		s.rpc.Serve(listener)
	}()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/index/ask", func(w http.ResponseWriter, r *http.Request) {
			index := ai.Get()
			buf, err := json.MarshalIndent(index, "", "  ")
			if err != nil {
				http.Error(w, "Error", http.StatusInternalServerError)
				return
			}
			w.Write(buf)
		})
		mux.HandleFunc("/index/miners", func(w http.ResponseWriter, r *http.Request) {
			index := mi.Get()
			buf, err := json.MarshalIndent(index, "", "  ")
			if err != nil {
				http.Error(w, "Error", http.StatusInternalServerError)
				return
			}
			w.Write(buf)
		})
		mux.HandleFunc("/index/slashing", func(w http.ResponseWriter, r *http.Request) {
			index := si.Get()
			buf, err := json.MarshalIndent(index, "", "  ")
			if err != nil {
				http.Error(w, "Error", http.StatusInternalServerError)
				return
			}
			w.Write(buf)
		})
		if err := http.ListenAndServe(":8889", mux); err != nil {
			log.Fatalf("Failed to run Prometheus scrape endpoint: %v", err)
		}
	}()

	return s, nil
}

// Close shuts down the server
func (s *Server) Close() {
	s.rpc.GracefulStop()
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
	s.closeLotus()
	s.ip2l.Close()
}
