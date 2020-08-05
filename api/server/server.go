package server

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger2"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	buildinfoRpc "github.com/textileio/powergate/buildinfo/rpc"
	"github.com/textileio/powergate/deals"
	dealsModule "github.com/textileio/powergate/deals/module"
	"github.com/textileio/powergate/fchost"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/cidlogger"
	"github.com/textileio/powergate/ffs/coreipfs"
	"github.com/textileio/powergate/ffs/filcold"
	"github.com/textileio/powergate/ffs/manager"
	"github.com/textileio/powergate/ffs/minerselector/reptop"
	ffsRpc "github.com/textileio/powergate/ffs/rpc"
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/filchain"
	"github.com/textileio/powergate/gateway"
	"github.com/textileio/powergate/health"
	healthRpc "github.com/textileio/powergate/health/rpc"
	askRpc "github.com/textileio/powergate/index/ask/rpc"
	ask "github.com/textileio/powergate/index/ask/runner"
	faultsModule "github.com/textileio/powergate/index/faults/module"
	faultsRpc "github.com/textileio/powergate/index/faults/rpc"
	minerModule "github.com/textileio/powergate/index/miner/module"
	minerRpc "github.com/textileio/powergate/index/miner/rpc"
	"github.com/textileio/powergate/iplocation/maxmind"
	"github.com/textileio/powergate/lotus"
	pgnet "github.com/textileio/powergate/net"
	pgnetlotus "github.com/textileio/powergate/net/lotus"
	pgnetRpc "github.com/textileio/powergate/net/rpc"
	paychLotus "github.com/textileio/powergate/paych/lotus"
	"github.com/textileio/powergate/reputation"
	reputationRpc "github.com/textileio/powergate/reputation/rpc"
	txndstr "github.com/textileio/powergate/txndstransform"
	"github.com/textileio/powergate/util"
	walletModule "github.com/textileio/powergate/wallet/module"
	walletRpc "github.com/textileio/powergate/wallet/rpc"
	"google.golang.org/grpc"
)

const (
	datastoreFolderName    = "datastore"
	lotusConnectionRetries = 10

	// WSPingInterval controls the WebSocket keepalive pinging interval. Must be >= 1s.
	WSPingInterval = time.Second * 5
)

var (
	log = logging.Logger("server")
)

// Server represents the configured lotus client and filecoin grpc server.
type Server struct {
	ds datastore.TxnDatastore

	mm *maxmind.MaxMind
	ai *ask.Runner
	mi *minerModule.Index
	fi *faultsModule.Index
	dm *dealsModule.Module
	wm *walletModule.Module
	rm *reputation.Module
	nm pgnet.Module
	hm *health.Module

	ffsManager *manager.Manager
	sched      *scheduler.Scheduler
	hs         ffs.HotStorage
	l          *cidlogger.CidLogger

	grpcServer *grpc.Server

	webProxy *http.Server

	gateway     *gateway.Gateway
	indexServer *http.Server

	closeLotus func()
}

// Config specifies server settings.
type Config struct {
	WalletInitialFunds   big.Int
	IpfsAPIAddr          ma.Multiaddr
	LotusAddress         ma.Multiaddr
	LotusAuthToken       string
	LotusMasterAddr      string
	AutocreateMasterAddr bool
	FFSUseMasterAddr     bool
	Devnet               bool
	GrpcHostNetwork      string
	GrpcHostAddress      ma.Multiaddr
	GrpcServerOpts       []grpc.ServerOption
	GrpcWebProxyAddress  string
	RepoPath             string
	GatewayHostAddr      string
	MaxMindDBFolder      string
}

// NewServer starts and returns a new server with the given configuration.
func NewServer(conf Config) (*Server, error) {
	if conf.FFSUseMasterAddr && !(len(conf.LotusMasterAddr) > 0 || conf.AutocreateMasterAddr) {
		return nil, fmt.Errorf("FFSUseMasterAddr requires LotusMasterAddr or AutocreateMasterAddr to be provided")
	}

	var err error
	var masterAddr address.Address
	c, cls, err := lotus.New(conf.LotusAddress, conf.LotusAuthToken, lotusConnectionRetries)
	if err != nil {
		return nil, fmt.Errorf("connecting to lotus node: %s", err)
	}

	if conf.Devnet {
		// Wait for the devnet to bootstrap completely and generate at least 1 block.
		time.Sleep(time.Second * 6)
		if masterAddr, err = c.WalletDefaultAddress(context.Background()); err != nil {
			return nil, fmt.Errorf("getting default wallet addr as masteraddr: %s", err)
		}
	} else {
		if masterAddr, err = address.NewFromString(conf.LotusMasterAddr); err != nil {
			return nil, fmt.Errorf("parsing masteraddr: %s", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	network, err := c.StateNetworkName(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting Lotus network name: %s", err)
	}
	networkName := string(network)
	log.Infof("Detected Lotus node connected to network: %s", networkName)

	fchost, err := fchost.New(networkName, !conf.Devnet)
	if err != nil {
		return nil, fmt.Errorf("creating filecoin host: %s", err)
	}
	if !conf.Devnet {
		if err := fchost.Bootstrap(); err != nil {
			return nil, fmt.Errorf("bootstrapping filecoin host: %s", err)
		}
	}

	path := filepath.Join(conf.RepoPath, datastoreFolderName)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating repo folder: %s", err)
	}

	log.Info("Opening badger database...")
	opts := &badger.DefaultOptions
	ds, err := badger.NewDatastore(path, opts)
	if err != nil {
		return nil, fmt.Errorf("opening datastore on repo: %s", err)
	}

	log.Info("Wiring internal components...")
	mm, err := maxmind.New(filepath.Join(conf.MaxMindDBFolder, "GeoLite2-City.mmdb"))
	if err != nil {
		return nil, fmt.Errorf("opening maxmind database: %s", err)
	}
	ai, err := ask.New(txndstr.Wrap(ds, "index/ask"), c)
	if err != nil {
		return nil, fmt.Errorf("creating ask index: %s", err)
	}
	mi, err := minerModule.New(txndstr.Wrap(ds, "index/miner"), c, fchost, mm)
	if err != nil {
		return nil, fmt.Errorf("creating miner index: %s", err)
	}
	si, err := faultsModule.New(txndstr.Wrap(ds, "index/faults"), c)
	if err != nil {
		return nil, fmt.Errorf("creating faults index: %s", err)
	}
	dm, err := dealsModule.New(txndstr.Wrap(ds, "deals"), c, deals.WithImportPath(filepath.Join(conf.RepoPath, "imports")))
	if err != nil {
		return nil, fmt.Errorf("creating deal module: %s", err)
	}
	wm, err := walletModule.New(c, masterAddr, conf.WalletInitialFunds, conf.AutocreateMasterAddr, networkName)
	if err != nil {
		return nil, fmt.Errorf("creating wallet module: %s", err)
	}
	pm := paychLotus.New(c)
	rm := reputation.New(txndstr.Wrap(ds, "reputation"), mi, si, ai)
	nm := pgnetlotus.New(c, mm)
	hm := health.New(nm)

	ipfs, err := httpapi.NewApi(conf.IpfsAPIAddr)
	if err != nil {
		return nil, fmt.Errorf("creating ipfs client: %s", err)
	}

	chain := filchain.New(c)
	ms := reptop.New(rm, ai)

	l := cidlogger.New(txndstr.Wrap(ds, "ffs/cidlogger"))
	cs := filcold.New(ms, dm, ipfs, chain, l)
	hs, err := coreipfs.New(ipfs, l)
	if err != nil {
		return nil, fmt.Errorf("creating coreipfs: %s", err)
	}
	sched, err := scheduler.New(txndstr.Wrap(ds, "ffs/scheduler"), l, hs, cs)
	if err != nil {
		return nil, fmt.Errorf("creating scheduler: %s", err)
	}

	ffsManager, err := manager.New(txndstr.Wrap(ds, "ffs/manager"), wm, pm, dm, sched, conf.FFSUseMasterAddr)
	if err != nil {
		return nil, fmt.Errorf("creating ffs instance: %s", err)
	}

	log.Info("Starting gRPC, gateway and index HTTP servers...")
	grpcServer := grpc.NewServer(conf.GrpcServerOpts...)
	wrappedGRPCServer := wrapGRPCServer(grpcServer)
	httpFFSAuthInterceptor, err := newHTTPFFSAuthInterceptor(conf, ffsManager)
	if err != nil {
		return nil, fmt.Errorf("creating ffsHTTPAuth: %s", err)
	}
	webProxy := createProxyServer(wrappedGRPCServer, httpFFSAuthInterceptor, conf.GrpcWebProxyAddress)

	gateway := gateway.NewGateway(conf.GatewayHostAddr, ai, mi, si, rm)
	gateway.Start()

	s := &Server{
		ds: ds,

		mm: mm,

		ai: ai,
		mi: mi,
		fi: si,
		dm: dm,
		wm: wm,
		rm: rm,
		nm: nm,
		hm: hm,

		ffsManager: ffsManager,
		sched:      sched,
		hs:         hs,
		l:          l,

		grpcServer: grpcServer,
		webProxy:   webProxy,
		gateway:    gateway,
		closeLotus: cls,
	}

	if err := startGRPCServices(grpcServer, webProxy, s, conf.GrpcHostNetwork, conf.GrpcHostAddress); err != nil {
		return nil, fmt.Errorf("starting GRPC services: %s", err)
	}

	s.indexServer = startIndexHTTPServer(s)

	log.Info("Starting finished, serving requests")

	return s, nil
}

type ffsHTTPAuth struct {
	cont       http.Handler
	ffsManager *manager.Manager
}

func newHTTPFFSAuthInterceptor(conf Config, m *manager.Manager) (*ffsHTTPAuth, error) {
	log.Info("Starting IPFS reverse proxy...")
	ipfsIP, err := util.TCPAddrFromMultiAddr(conf.IpfsAPIAddr)
	if err != nil {
		return nil, fmt.Errorf("converting IPFS multiaddr to tcp addr: %s", err)
	}

	urlIPFS, err := url.Parse("http://" + ipfsIP)
	if err != nil {
		return nil, fmt.Errorf("generating IPFS URL for reverse proxy: %s", err)
	}
	rph := httputil.NewSingleHostReverseProxy(urlIPFS)
	fha := &ffsHTTPAuth{
		cont:       rph,
		ffsManager: m,
	}
	return fha, nil
}

func (fha *ffsHTTPAuth) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	authFFS := r.Header.Get("x-ipfs-ffs-auth")
	_, err := fha.ffsManager.GetByAuthToken(authFFS)
	if authFFS == "" || err == manager.ErrAuthTokenNotFound {
		http.Error(rw, "FFS token required", http.StatusUnauthorized)
		return
	}
	fha.cont.ServeHTTP(rw, r)
}

func (fha *ffsHTTPAuth) IsIPFSRequest(r *http.Request) bool {
	return len(r.Header.Get("x-ipfs-ffs-auth")) > 0
}

func createProxyServer(wrappedGRPCServer *grpcweb.WrappedGrpcServer, fha *ffsHTTPAuth, webProxyAddr string) *http.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fha.IsIPFSRequest(r) {
			fha.ServeHTTP(w, r)
		} else if wrappedGRPCServer.IsGrpcWebRequest(r) ||
			wrappedGRPCServer.IsAcceptableGrpcCorsRequest(r) ||
			wrappedGRPCServer.IsGrpcWebSocketRequest(r) {
			wrappedGRPCServer.ServeHTTP(w, r)
		}
	})
	webProxy := &http.Server{
		Addr:    webProxyAddr,
		Handler: handler,
	}
	return webProxy
}

func wrapGRPCServer(grpcServer *grpc.Server) *grpcweb.WrappedGrpcServer {
	wrappedServer := grpcweb.WrapServer(
		grpcServer,
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true
		}),
		grpcweb.WithAllowedRequestHeaders([]string{"*"}),
		grpcweb.WithWebsockets(true),
		grpcweb.WithWebsocketPingInterval(WSPingInterval),
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool {
			return true
		}),
	)

	return wrappedServer
}

func startGRPCServices(server *grpc.Server, webProxy *http.Server, s *Server, hostNetwork string, hostAddress ma.Multiaddr) error {
	buildinfoService := buildinfoRpc.New()
	netService := pgnetRpc.New(s.nm)
	healthService := healthRpc.New(s.hm)
	walletService := walletRpc.New(s.wm)
	reputationService := reputationRpc.New(s.rm)
	askService := askRpc.New(s.ai)
	minerService := minerRpc.New(s.mi)
	faultsService := faultsRpc.New(s.fi)
	ffsService := ffsRpc.New(s.ffsManager, s.hs)

	hostAddr, err := util.TCPAddrFromMultiAddr(hostAddress)
	if err != nil {
		return fmt.Errorf("parsing host multiaddr: %s", err)
	}
	listener, err := net.Listen(hostNetwork, hostAddr)
	if err != nil {
		return fmt.Errorf("listening to grpc: %s", err)
	}
	go func() {
		buildinfoRpc.RegisterRPCServiceServer(server, buildinfoService)
		pgnetRpc.RegisterRPCServiceServer(server, netService)
		healthRpc.RegisterRPCServiceServer(server, healthService)
		walletRpc.RegisterRPCServiceServer(server, walletService)
		reputationRpc.RegisterRPCServiceServer(server, reputationService)
		askRpc.RegisterRPCServiceServer(server, askService)
		minerRpc.RegisterRPCServiceServer(server, minerService)
		faultsRpc.RegisterRPCServiceServer(server, faultsService)
		ffsRpc.RegisterRPCServiceServer(server, ffsService)
		if err := server.Serve(listener); err != nil {
			log.Errorf("serving grpc endpoint: %s", err)
		}
	}()

	go func() {
		if err := webProxy.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("error starting proxy: %v", err)
		}
	}()
	return nil
}

func startIndexHTTPServer(s *Server) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/index/ask", func(w http.ResponseWriter, r *http.Request) {
		index := s.ai.Get()
		buf, err := json.MarshalIndent(index, "", "  ")
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(buf); err != nil {
			log.Errorf("writing response body: %s", err)
		}
	})
	mux.HandleFunc("/index/miners", func(w http.ResponseWriter, r *http.Request) {
		index := s.mi.Get()
		buf, err := json.MarshalIndent(index, "", "  ")
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(buf); err != nil {
			log.Errorf("writing response body: %s", err)
		}
	})
	mux.HandleFunc("/index/faults", func(w http.ResponseWriter, r *http.Request) {
		index := s.fi.Get()
		buf, err := json.MarshalIndent(index, "", "  ")
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(buf); err != nil {
			log.Errorf("writing response body: %s", err)
		}
	})

	srv := &http.Server{Addr: ":8889", Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serving index http: %v", err)
		}
	}()
	return srv
}

// Close shuts down the server.
func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := s.indexServer.Shutdown(ctx); err != nil {
		log.Errorf("closing down index server: %s", err)
	}

	log.Info("closing gRPC endpoints...")
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.webProxy.Shutdown(ctx); err != nil {
		log.Errorf("closing down proxy: %s", err)
	}
	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()
	t := time.NewTimer(10 * time.Second)
	select {
	case <-t.C:
		s.grpcServer.Stop()
	case <-stopped:
		t.Stop()
	}
	log.Info("gRPC endpoints closed")

	if err := s.ffsManager.Close(); err != nil {
		log.Errorf("closing ffs manager: %s", err)
	}
	if err := s.sched.Close(); err != nil {
		log.Errorf("closing ffs scheduler: %s", err)
	}
	if err := s.l.Close(); err != nil {
		log.Errorf("closing cid logger: %s", err)
	}
	if err := s.rm.Close(); err != nil {
		log.Errorf("closing reputation module: %s", err)
	}
	if err := s.ai.Close(); err != nil {
		log.Errorf("closing ask index: %s", err)
	}
	if err := s.mi.Close(); err != nil {
		log.Errorf("closing miner index: %s", err)
	}
	if err := s.fi.Close(); err != nil {
		log.Errorf("closing faults index: %s", err)
	}

	log.Info("closing datastore...")
	if err := s.ds.Close(); err != nil {
		log.Errorf("closing datastore: %s", err)
	}
	log.Info("datastore closed")

	if err := s.gateway.Stop(); err != nil {
		log.Errorf("closing gateway: %s", err)
	}
	s.closeLotus()
	if err := s.mm.Close(); err != nil {
		log.Errorf("closing maxmind: %s", err)
	}
}
