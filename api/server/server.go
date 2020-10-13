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
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/gogo/status"
	grpcm "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger2"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	mongods "github.com/textileio/go-ds-mongo"
	adminService "github.com/textileio/powergate/api/server/admin"
	powergateService "github.com/textileio/powergate/api/server/powergate"
	buildinfoRpc "github.com/textileio/powergate/buildinfo/rpc"
	"github.com/textileio/powergate/deals"
	dealsModule "github.com/textileio/powergate/deals/module"
	"github.com/textileio/powergate/fchost"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/coreipfs"
	"github.com/textileio/powergate/ffs/filcold"
	"github.com/textileio/powergate/ffs/joblogger"
	"github.com/textileio/powergate/ffs/manager"
	"github.com/textileio/powergate/ffs/minerselector/reptop"
	"github.com/textileio/powergate/ffs/minerselector/sr2"
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
	adminProto "github.com/textileio/powergate/proto/admin/v1"
	powergateProto "github.com/textileio/powergate/proto/powergate/v1"
	"github.com/textileio/powergate/reputation"
	reputationRpc "github.com/textileio/powergate/reputation/rpc"
	txndstr "github.com/textileio/powergate/txndstransform"
	"github.com/textileio/powergate/util"
	walletModule "github.com/textileio/powergate/wallet/module"
	walletRpc "github.com/textileio/powergate/wallet/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	datastoreFolderName    = "datastore"
	lotusConnectionRetries = 10
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
	l          *joblogger.Logger

	grpcServer *grpc.Server

	webProxy *http.Server

	gateway     *gateway.Gateway
	indexServer *http.Server
}

// Config specifies server settings.
type Config struct {
	RepoPath        string
	MaxMindDBFolder string
	Devnet          bool
	IpfsAPIAddr     ma.Multiaddr

	LotusAddress    ma.Multiaddr
	LotusAuthToken  string
	LotusMasterAddr string

	GrpcHostNetwork     string
	GrpcHostAddress     ma.Multiaddr
	GrpcServerOpts      []grpc.ServerOption
	GrpcWebProxyAddress string

	GatewayBasePath string
	GatewayHostAddr string

	MongoURI string
	MongoDB  string

	FFSUseMasterAddr       bool
	FFSDealFinalityTimeout time.Duration
	FFSMinimumPieceSize    uint64
	FFSAdminToken          string
	SchedMaxParallel       int
	MinerSelector          string
	MinerSelectorParams    string
	DealWatchPollDuration  time.Duration
	AutocreateMasterAddr   bool
	WalletInitialFunds     big.Int

	AskIndexQueryAskTimeout time.Duration
	AskindexMaxParallel     int
	AskIndexRefreshInterval time.Duration
	AskIndexRefreshOnStart  bool

	DisableIndices bool
}

// NewServer starts and returns a new server with the given configuration.
func NewServer(conf Config) (*Server, error) {
	if conf.FFSUseMasterAddr && !conf.Devnet && !(len(conf.LotusMasterAddr) > 0 || conf.AutocreateMasterAddr) {
		return nil, fmt.Errorf("FFSUseMasterAddr requires LotusMasterAddr or AutocreateMasterAddr to be provided")
	}

	var err error
	clientBuilder, err := lotus.NewBuilder(conf.LotusAddress, conf.LotusAuthToken, lotusConnectionRetries)
	if err != nil {
		return nil, fmt.Errorf("creating lotus client builder: %s", err)
	}
	lotus.MonitorLotusSync(clientBuilder)

	c, cls, err := clientBuilder()
	if err != nil {
		return nil, fmt.Errorf("connecting to lotus node: %s", err)
	}

	masterAddr, err := evaluateMasterAddr(conf, c)
	if err != nil {
		return nil, fmt.Errorf("evaluating ffs master addr: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	network, err := c.StateNetworkName(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting Lotus network name: %s", err)
	}
	cls()

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

	ds, err := createDatastore(conf)
	if err != nil {
		return nil, fmt.Errorf("creating datastore: %s", err)
	}

	log.Info("Wiring internal components...")
	mm, err := maxmind.New(filepath.Join(conf.MaxMindDBFolder, "GeoLite2-City.mmdb"))
	if err != nil {
		return nil, fmt.Errorf("opening maxmind database: %s", err)
	}
	askConf := ask.Config{
		Disable:         conf.DisableIndices,
		QueryAskTimeout: conf.AskIndexQueryAskTimeout,
		MaxParallel:     conf.AskindexMaxParallel,
		RefreshInterval: conf.AskIndexRefreshInterval,
		RefreshOnStart:  conf.Devnet || conf.AskIndexRefreshOnStart,
	}
	ai, err := ask.New(txndstr.Wrap(ds, "index/ask"), clientBuilder, askConf)
	if err != nil {
		return nil, fmt.Errorf("creating ask index: %s", err)
	}
	mi, err := minerModule.New(txndstr.Wrap(ds, "index/miner"), clientBuilder, fchost, mm, conf.DisableIndices)
	if err != nil {
		return nil, fmt.Errorf("creating miner index: %s", err)
	}
	si, err := faultsModule.New(txndstr.Wrap(ds, "index/faults"), clientBuilder, conf.DisableIndices)
	if err != nil {
		return nil, fmt.Errorf("creating faults index: %s", err)
	}
	if conf.Devnet {
		conf.DealWatchPollDuration = time.Second
	}
	dm, err := dealsModule.New(txndstr.Wrap(ds, "deals"), clientBuilder, conf.DealWatchPollDuration, deals.WithImportPath(filepath.Join(conf.RepoPath, "imports")))
	if err != nil {
		return nil, fmt.Errorf("creating deal module: %s", err)
	}
	wm, err := walletModule.New(clientBuilder, masterAddr, conf.WalletInitialFunds, conf.AutocreateMasterAddr, networkName)
	if err != nil {
		return nil, fmt.Errorf("creating wallet module: %s", err)
	}
	pm := paychLotus.New(clientBuilder)
	rm := reputation.New(txndstr.Wrap(ds, "reputation"), mi, si, ai)
	nm := pgnetlotus.New(clientBuilder, mm)
	hm := health.New(nm)

	ipfs, err := httpapi.NewApi(conf.IpfsAPIAddr)
	if err != nil {
		return nil, fmt.Errorf("creating ipfs client: %s", err)
	}

	chain := filchain.New(clientBuilder)

	ms, err := getMinerSelector(conf, rm, ai, clientBuilder)
	if err != nil {
		return nil, fmt.Errorf("creating miner selector: %s", err)
	}

	l := joblogger.New(txndstr.Wrap(ds, "ffs/joblogger"))
	cs := filcold.New(ms, dm, ipfs, chain, l, conf.FFSMinimumPieceSize)
	hs, err := coreipfs.New(ipfs, l)
	if err != nil {
		return nil, fmt.Errorf("creating coreipfs: %s", err)
	}

	var sr2rf func() (int, error)
	if ms, ok := ms.(*sr2.MinerSelector); ok {
		sr2rf = ms.GetReplicationFactor
	}
	sched, err := scheduler.New(txndstr.Wrap(ds, "ffs/scheduler"), l, hs, cs, conf.SchedMaxParallel, conf.FFSDealFinalityTimeout, sr2rf)
	if err != nil {
		return nil, fmt.Errorf("creating scheduler: %s", err)
	}

	ffsManager, err := manager.New(txndstr.Wrap(ds, "ffs/manager"), wm, pm, dm, sched, conf.FFSUseMasterAddr, conf.Devnet)
	if err != nil {
		return nil, fmt.Errorf("creating ffs instance: %s", err)
	}

	log.Info("Starting gRPC, gateway and index HTTP servers...")
	unaryInterceptorChain := grpcm.WithUnaryServerChain(
		adminAuth(conf),
	)
	opts := append(conf.GrpcServerOpts, unaryInterceptorChain)
	grpcServer := grpc.NewServer(opts...)
	wrappedGRPCServer := wrapGRPCServer(grpcServer)
	httpFFSAuthInterceptor, err := newHTTPFFSAuthInterceptor(conf, ffsManager)
	if err != nil {
		return nil, fmt.Errorf("creating ffsHTTPAuth: %s", err)
	}
	webProxy := createProxyServer(wrappedGRPCServer, httpFFSAuthInterceptor, conf.GrpcWebProxyAddress)

	gateway := gateway.NewGateway(conf.GatewayHostAddr, ai, mi, si, rm)
	gateway.Start(conf.GatewayBasePath)

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
	rph.FlushInterval = -1
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
		} else {
			http.NotFound(w, r)
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
	ffsService := ffsRpc.New(s.ffsManager, s.wm, s.hs)
	powergateService := powergateService.New(s.sched, s.wm)
	adminService := adminService.New(s.ffsManager, s.sched)

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
		powergateProto.RegisterPowergateServiceServer(server, powergateService)
		adminProto.RegisterPowergateAdminServiceServer(server, adminService)
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
		log.Errorf("closing joblogger: %s", err)
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
	if err := s.mm.Close(); err != nil {
		log.Errorf("closing maxmind: %s", err)
	}
}

func createDatastore(conf Config) (datastore.TxnDatastore, error) {
	if conf.MongoURI != "" {
		log.Info("Opening Mongo database...")
		mongoCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if conf.MongoDB == "" {
			return nil, fmt.Errorf("mongo database name is empty")
		}
		ds, err := mongods.New(mongoCtx, conf.MongoURI, conf.MongoDB)
		if err != nil {
			return nil, fmt.Errorf("opening mongo datastore: %s", err)
		}
		return ds, nil
	}

	log.Info("Opening badger database...")
	path := filepath.Join(conf.RepoPath, datastoreFolderName)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating repo folder: %s", err)
	}
	opts := &badger.DefaultOptions
	ds, err := badger.NewDatastore(path, opts)
	if err != nil {
		return nil, fmt.Errorf("opening badger datastore: %s", err)
	}
	return ds, nil
}

func getMinerSelector(conf Config, rm *reputation.Module, ai *ask.Runner, cb lotus.ClientBuilder) (ffs.MinerSelector, error) {
	if conf.Devnet {
		return reptop.New(rm, ai), nil
	}
	var ms ffs.MinerSelector
	var err error

	switch conf.MinerSelector {
	case "reputation":
		ms = reptop.New(rm, ai)
	case "sr2":
		ms, err = sr2.New(conf.MinerSelectorParams, cb)
		if err != nil {
			return nil, fmt.Errorf("creating sr2 miner selector: %s", err)
		}
	default:
		return nil, fmt.Errorf("unknown miner selector: %s", conf.MinerSelector)
	}

	return ms, nil
}

func evaluateMasterAddr(conf Config, c *apistruct.FullNodeStruct) (address.Address, error) {
	var res address.Address
	if conf.Devnet {
		// Wait for the devnet to bootstrap completely and generate at least 1 block.
		time.Sleep(time.Second * 6)
		res, err := c.WalletDefaultAddress(context.Background())
		if err != nil {
			return address.Address{}, fmt.Errorf("getting default wallet addr as masteraddr: %s", err)
		}
		return res, nil
	}
	res, err := address.NewFromString(conf.LotusMasterAddr)
	if err != nil {
		return address.Address{}, fmt.Errorf("parsing masteraddr: %s", err)
	}
	return res, nil
}
func adminAuth(conf Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if conf.FFSAdminToken == "" {
			return handler(ctx, req)
		}

		adminServicePrefix := "/proto.admin.v1.PowergateAdminService"

		method, _ := grpc.Method(ctx)

		if !strings.HasPrefix(method, adminServicePrefix) {
			return handler(ctx, req)
		}

		adminToken := metautils.ExtractIncoming(ctx).Get("X-pow-admin-token")
		if adminToken != conf.FFSAdminToken {
			return nil, status.Error(codes.PermissionDenied, "Method requires admin permission")
		}
		return handler(ctx, req)

	}
}
