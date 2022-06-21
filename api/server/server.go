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
	"github.com/filecoin-project/lotus/api"
	grpcm "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/ipfs/go-datastore"
	kt "github.com/ipfs/go-datastore/keytransform"
	badger "github.com/ipfs/go-ds-badger2"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	measure "github.com/textileio/go-ds-measure"
	mongods "github.com/textileio/go-ds-mongo"
	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
	"github.com/textileio/powergate/v2/api/server/admin"
	"github.com/textileio/powergate/v2/api/server/user"
	"github.com/textileio/powergate/v2/deals"
	dealsModule "github.com/textileio/powergate/v2/deals/module"
	"github.com/textileio/powergate/v2/fchost"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/ffs/coreipfs"
	"github.com/textileio/powergate/v2/ffs/filcold"
	"github.com/textileio/powergate/v2/ffs/joblogger"
	"github.com/textileio/powergate/v2/ffs/manager"
	"github.com/textileio/powergate/v2/ffs/minerselector/reptop"
	"github.com/textileio/powergate/v2/ffs/minerselector/sr2"
	"github.com/textileio/powergate/v2/ffs/scheduler"
	"github.com/textileio/powergate/v2/filchain"
	"github.com/textileio/powergate/v2/gateway"
	ask "github.com/textileio/powergate/v2/index/ask/runner"
	faultsModule "github.com/textileio/powergate/v2/index/faults/module"
	minerIndex "github.com/textileio/powergate/v2/index/miner/lotusidx"
	"github.com/textileio/powergate/v2/iplocation/maxmind"
	"github.com/textileio/powergate/v2/lotus"
	"github.com/textileio/powergate/v2/migration"
	"github.com/textileio/powergate/v2/reputation"
	txndstr "github.com/textileio/powergate/v2/txndstransform"
	"github.com/textileio/powergate/v2/util"
	lotusWallet "github.com/textileio/powergate/v2/wallet/lotuswallet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	datastoreFolderName = "datastore"
)

var (
	log = logging.Logger("server")

	nonCompliantAPIs = []string{
		"/ffs.rpc.RPCService/SendFil",
	}

	// Migrations contains the list of supported migrations.
	Migrations = map[int]migration.Migration{
		1: migration.V1MultitenancyMigration,
		2: migration.V2StorageInfoDealIDs,
		3: migration.V3StorageJobsIndexMigration,
		4: migration.V4RecordsMigration,
		5: migration.V5DeleteOldMinerIndex,
	}
)

// Server represents the configured lotus client and filecoin grpc server.
type Server struct {
	ds datastore.TxnDatastore

	mm *maxmind.MaxMind
	ai *ask.Runner
	mi *minerIndex.Index
	fi *faultsModule.Index
	dm *dealsModule.Module
	wm *lotusWallet.Module
	rm *reputation.Module

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

	LotusAddress           ma.Multiaddr
	LotusAuthToken         string
	LotusMasterAddr        string
	LotusConnectionRetries int

	GrpcHostNetwork     string
	GrpcHostAddress     ma.Multiaddr
	GrpcServerOpts      []grpc.ServerOption
	GrpcWebProxyAddress string

	GatewayBasePath      string
	GatewayHostAddr      string
	IndexRawJSONHostAddr string

	MongoURI string
	MongoDB  string

	FFSAdminToken                string
	FFSUseMasterAddr             bool
	FFSDealFinalityTimeout       time.Duration
	FFSMinimumPieceSize          uint64
	FFSRetrievalNextEventTimeout time.Duration
	FFSMaxParallelDealPreparing  int
	FFSGCAutomaticGCInterval     time.Duration
	FFSGCStageGracePeriod        time.Duration
	SchedMaxParallel             int
	MinerSelector                string
	MinerSelectorParams          string
	DealWatchPollDuration        time.Duration
	AutocreateMasterAddr         bool
	WalletInitialFunds           big.Int

	AskIndexQueryAskTimeout time.Duration
	AskindexMaxParallel     int
	AskIndexRefreshInterval time.Duration
	AskIndexRefreshOnStart  bool

	IndexMinersRefreshOnStart     bool
	IndexMinersOnChainMaxParallel int
	IndexMinersOnChainFrequency   time.Duration

	DisableIndices bool

	DisableNonCompliantAPIs bool
}

// NewServer starts and returns a new server with the given configuration.
func NewServer(conf Config) (*Server, error) {
	if conf.FFSUseMasterAddr && !conf.Devnet && !(len(conf.LotusMasterAddr) > 0 || conf.AutocreateMasterAddr) {
		return nil, fmt.Errorf("FFSUseMasterAddr requires LotusMasterAddr or AutocreateMasterAddr to be provided")
	}

	var err error
	clientBuilder, err := lotus.NewBuilder(conf.LotusAddress, conf.LotusAuthToken, conf.LotusConnectionRetries)
	if err != nil {
		return nil, fmt.Errorf("creating lotus client builder: %s", err)
	}
	lsm, err := lotus.NewSyncMonitor(clientBuilder)
	if err != nil {
		return nil, fmt.Errorf("creating lotus sync monitor: %s", err)
	}

	c, cls, err := clientBuilder(context.Background())
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

	if err := runMigrations(conf); err != nil {
		return nil, fmt.Errorf("running migrations: %s", err)
	}

	ds, err := createDatastore(conf, false)
	if err != nil {
		return nil, fmt.Errorf("creating datastore: %s", err)
	}

	log.Info("Wiring internal components...")
	mm, err := maxmind.New(filepath.Join(conf.MaxMindDBFolder, "GeoLite2-City.mmdb"))
	if err != nil {
		return nil, fmt.Errorf("opening maxmind database: %s", err)
	}
	askIdxConf := ask.Config{
		Disable:         conf.DisableIndices,
		QueryAskTimeout: conf.AskIndexQueryAskTimeout,
		MaxParallel:     conf.AskindexMaxParallel,
		RefreshInterval: conf.AskIndexRefreshInterval,
		RefreshOnStart:  conf.Devnet || conf.AskIndexRefreshOnStart,
	}
	log.Info("Starting ask index...")
	ai, err := ask.New(txndstr.Wrap(ds, "index/ask"), clientBuilder, askIdxConf)
	if err != nil {
		return nil, fmt.Errorf("creating ask index: %s", err)
	}

	log.Info("Starting miner index...")
	minerIdxConf := minerIndex.Config{
		RefreshOnStart:     conf.IndexMinersRefreshOnStart,
		Disable:            conf.DisableIndices,
		OnChainMaxParallel: conf.IndexMinersOnChainMaxParallel,
		OnChainFrequency:   conf.IndexMinersOnChainFrequency,
	}
	mi, err := minerIndex.New(kt.Wrap(ds, kt.PrefixTransform{Prefix: datastore.NewKey("index/miner")}), clientBuilder, fchost, mm, minerIdxConf)
	if err != nil {
		return nil, fmt.Errorf("creating miner index: %s", err)
	}

	log.Info("Starting faults index...")
	si, err := faultsModule.New(txndstr.Wrap(ds, "index/faults"), clientBuilder, conf.DisableIndices)
	if err != nil {
		return nil, fmt.Errorf("creating faults index: %s", err)
	}
	if conf.Devnet {
		conf.DealWatchPollDuration = time.Second
	}

	log.Info("Starting deals module...")
	dm, err := dealsModule.New(txndstr.Wrap(ds, "deals"), clientBuilder, conf.DealWatchPollDuration, conf.FFSDealFinalityTimeout, deals.WithImportPath(filepath.Join(conf.RepoPath, "imports")))
	if err != nil {
		return nil, fmt.Errorf("creating deal module: %s", err)
	}

	log.Info("Starting wallet module...")
	wm, err := lotusWallet.New(clientBuilder, masterAddr, conf.WalletInitialFunds, conf.AutocreateMasterAddr, networkName)
	if err != nil {
		return nil, fmt.Errorf("creating wallet module: %s", err)
	}
	rm := reputation.New(txndstr.Wrap(ds, "reputation"), mi, si, ai)

	ipfs, err := httpapi.NewApi(conf.IpfsAPIAddr)
	if err != nil {
		return nil, fmt.Errorf("creating ipfs client: %s", err)
	}

	chain := filchain.New(clientBuilder)

	ms, err := getMinerSelector(conf, rm, ai, clientBuilder)
	if err != nil {
		return nil, fmt.Errorf("creating miner selector: %s", err)
	}

	l := joblogger.New(txndstr.Wrap(ds, "ffs/joblogger_v2"))
	if conf.Devnet {
		conf.FFSMinimumPieceSize = 0
	}
	cs := filcold.New(ms, dm, wm, ipfs, chain, l, lsm, conf.FFSMinimumPieceSize, conf.FFSMaxParallelDealPreparing, conf.FFSRetrievalNextEventTimeout)
	hs, err := coreipfs.New(txndstr.Wrap(ds, "ffs/coreipfs"), ipfs, l)
	if err != nil {
		return nil, fmt.Errorf("creating coreipfs: %s", err)
	}

	log.Info("Starting FFS scheduler...")
	var sr2rf func() (int, error)
	if ms, ok := ms.(*sr2.MinerSelector); ok {
		sr2rf = ms.GetReplicationFactor
	}
	gcConfig := scheduler.GCConfig{StageGracePeriod: conf.FFSGCStageGracePeriod, AutoGCInterval: conf.FFSGCAutomaticGCInterval}
	sched, err := scheduler.New(txndstr.Wrap(ds, "ffs/scheduler"), l, hs, cs, conf.SchedMaxParallel, conf.FFSDealFinalityTimeout, sr2rf, gcConfig)
	if err != nil {
		return nil, fmt.Errorf("creating scheduler: %s", err)
	}

	ffsManager, err := manager.New(txndstr.Wrap(ds, "ffs/manager"), wm, dm, sched, conf.FFSUseMasterAddr, conf.Devnet)
	if err != nil {
		return nil, fmt.Errorf("creating ffs instance: %s", err)
	}

	log.Info("Starting gRPC, gateway and index HTTP servers...")

	unaryInterceptors := []grpc.UnaryServerInterceptor{adminAuth(conf)}
	if conf.DisableNonCompliantAPIs {
		unaryInterceptors = append(unaryInterceptors, nonCompliantAPIsInterceptor(nonCompliantAPIs))
	}
	unaryInterceptorChain := grpcm.WithUnaryServerChain(unaryInterceptors...)

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

	s.indexServer = startIndexHTTPServer(s, conf.IndexRawJSONHostAddr)

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
	userService := user.New(s.ffsManager, s.wm, s.hs)
	adminService := admin.New(s.ffsManager, s.sched, s.wm, s.dm, s.mi, s.ai)

	hostAddr, err := util.TCPAddrFromMultiAddr(hostAddress)
	if err != nil {
		return fmt.Errorf("parsing host multiaddr: %s", err)
	}
	listener, err := net.Listen(hostNetwork, hostAddr)
	if err != nil {
		return fmt.Errorf("listening to grpc: %s", err)
	}
	go func() {
		userPb.RegisterUserServiceServer(server, userService)
		adminPb.RegisterAdminServiceServer(server, adminService)
		if err := server.Serve(listener); err != nil {
			log.Errorf("serving grpc endpoint: %s", err)
		}
	}()

	go func() {
		if err := webProxy.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("starting proxy: %v", err)
		}
	}()
	return nil
}

func startIndexHTTPServer(s *Server, addr string) *http.Server {
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

	srv := &http.Server{Addr: addr, Handler: mux}
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
	if err := s.dm.Close(); err != nil {
		log.Errorf("closing deal module: %s", err)
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

func createDatastore(conf Config, longTimeout bool) (datastore.TxnDatastore, error) {
	var ds datastore.TxnDatastore
	var err error

	if conf.MongoURI != "" {
		log.Info("Opening Mongo database...")
		mongoCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if conf.MongoDB == "" {
			return nil, fmt.Errorf("mongo database name is empty")
		}
		var opts []mongods.Option
		if longTimeout {
			opts = []mongods.Option{mongods.WithOpTimeout(time.Hour), mongods.WithTxnTimeout(time.Hour)}
		}
		ds, err = mongods.New(mongoCtx, conf.MongoURI, conf.MongoDB, opts...)
		if err != nil {
			return nil, fmt.Errorf("opening mongo datastore: %s", err)
		}
	} else {
		log.Info("Opening badger database...")
		path := filepath.Join(conf.RepoPath, datastoreFolderName)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return nil, fmt.Errorf("creating repo folder: %s", err)
		}
		opts := &badger.DefaultOptions
		ds, err = badger.NewDatastore(path, opts)
		if err != nil {
			return nil, fmt.Errorf("opening badger datastore: %s", err)
		}
	}

	return measure.New("powergate.datastore", ds), nil
}

func getMinerSelector(conf Config, rm *reputation.Module, ai *ask.Runner, cb lotus.ClientBuilder) (ffs.MinerSelector, error) {
	if conf.Devnet {
		return reptop.New(cb, rm, ai), nil
	}
	var ms ffs.MinerSelector
	var err error

	switch conf.MinerSelector {
	case "reputation":
		ms = reptop.New(cb, rm, ai)
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

func evaluateMasterAddr(conf Config, c *api.FullNodeStruct) (address.Address, error) {
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

		adminServicePrefix := "/powergate.admin.v1.AdminService"

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

func nonCompliantAPIsInterceptor(nonCompliantAPIs []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method, _ := grpc.Method(ctx)
		for _, nonCompliantAPI := range nonCompliantAPIs {
			if method == nonCompliantAPI {
				return nil, status.Error(codes.PermissionDenied, "method disabled by powergate administrators")
			}
		}
		return handler(ctx, req)
	}
}

func runMigrations(conf Config) error {
	log.Infof("Ensuring migrations...")
	ds, err := createDatastore(conf, true)
	if err != nil {
		return fmt.Errorf("creating migration datastore: %s", err)
	}
	defer func() {
		if err := ds.Close(); err != nil {
			log.Errorf("closing migration datastore: %s", err)
		}
	}()

	m := migration.New(ds, Migrations)
	if err := m.Ensure(); err != nil {
		return fmt.Errorf("running migrations: %s", err)
	}
	log.Infof("Migrations ensured")

	return nil
}
