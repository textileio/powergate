package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	logging "github.com/ipfs/go-log/v2"
	homedir "github.com/mitchellh/go-homedir"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/server"
	"github.com/textileio/powergate/buildinfo"
	"github.com/textileio/powergate/util"
	"go.opencensus.io/plugin/runmetrics"
)

var (
	log    = logging.Logger("powd")
	config = viper.New()
)

func main() {
	// Configure flags.
	if err := setupFlags(); err != nil {
		log.Fatalf("configuring flags: %s", err)
	}

	// Create configuration from flags/envs.
	conf, err := configFromFlags()
	if err != nil {
		log.Fatalf("creating config from flags: %s", err)
	}

	// Configure logging.
	if err := setupLogging(conf.RepoPath); err != nil {
		log.Fatalf("configuring up logging: %s", err)
	}

	log.Infof("starting powd:\n%s", buildinfo.Summary())

	// Configuring Prometheus exporter.
	closeInstr, err := setupInstrumentation()
	if err != nil {
		log.Fatalf("starting instrumentation: %s", err)
	}
	confJSON, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		log.Fatalf("marshaling configuration: %s", err)
	}
	log.Infof("%s", confJSON)

	// Start server.
	log.Info("starting server...")
	powd, err := server.NewServer(conf)
	if err != nil {
		log.Fatalf("starting server: %s", err)
	}
	log.Info("server started.")

	// Wait for Ctrl+C and close.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	log.Info("Closing...")
	closeInstr()
	powd.Close()
	if conf.Devnet {
		if err := os.RemoveAll(conf.RepoPath); err != nil {
			log.Error(err)
		}
	}
	log.Info("Closed")
}

func configFromFlags() (server.Config, error) {
	devnet := config.GetBool("devnet")

	lotusToken, err := getLotusToken(devnet)
	if err != nil {
		return server.Config{}, fmt.Errorf("getting lotus auth token: %s", err)
	}

	repoPath, err := getRepoPath(devnet)
	if err != nil {
		return server.Config{}, fmt.Errorf("getting repo path: %s", err)
	}

	grpcHostMaddr, err := ma.NewMultiaddr(config.GetString("grpchostaddr"))
	if err != nil {
		return server.Config{}, fmt.Errorf("parsing grpchostaddr: %s", err)
	}

	maddr, err := ma.NewMultiaddr(config.GetString("lotushost"))
	if err != nil {
		return server.Config{}, fmt.Errorf("parsing lotus api multiaddr: %s", err)
	}

	walletInitialFunds := *big.NewInt(config.GetInt64("walletinitialfund"))
	ipfsAPIAddr := util.MustParseAddr(config.GetString("ipfsapiaddr"))
	lotusMasterAddr := config.GetString("lotusmasteraddr")
	autocreateMasterAddr := config.GetBool("autocreatemasteraddr")
	ffsUseMasterAddr := config.GetBool("ffsusemasteraddr")
	grpcWebProxyAddr := config.GetString("grpcwebproxyaddr")
	gatewayHostAddr := config.GetString("gatewayhostaddr")
	gatewayBasePath := config.GetString("gatewaybasepath")
	maxminddbfolder := config.GetString("maxminddbfolder")

	return server.Config{
		WalletInitialFunds:   walletInitialFunds,
		IpfsAPIAddr:          ipfsAPIAddr,
		LotusAddress:         maddr,
		LotusAuthToken:       lotusToken,
		LotusMasterAddr:      lotusMasterAddr,
		AutocreateMasterAddr: autocreateMasterAddr,
		FFSUseMasterAddr:     ffsUseMasterAddr,
		Devnet:               devnet,
		// ToDo: Support secure gRPC connection
		GrpcHostNetwork:     "tcp",
		GrpcHostAddress:     grpcHostMaddr,
		GrpcWebProxyAddress: grpcWebProxyAddr,
		RepoPath:            repoPath,
		GatewayHostAddr:     gatewayHostAddr,
		GatewayBasePath:     gatewayBasePath,
		MaxMindDBFolder:     maxminddbfolder,
	}, nil
}

func setupInstrumentation() (func(), error) {
	err := runmetrics.Enable(runmetrics.RunMetricOptions{
		EnableCPU:    true,
		EnableMemory: true,
	})
	if err != nil {
		return nil, fmt.Errorf("enabling runtime metrics: %s", err)
	}
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "textilefc",
	})
	if err != nil {
		return nil, fmt.Errorf("creating the prometheus stats exporter: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", pe)
	srv := &http.Server{Addr: ":8888", Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("running prometheus scrape endpoint: %v", err)
		}
	}()
	closeFunc := func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Error("shutting down prometheus server: %s", err)
		}
	}

	return closeFunc, nil
}

func setupLogging(repoPath string) error {
	if err := os.MkdirAll(repoPath, os.ModePerm); err != nil {
		return fmt.Errorf("creating repo folder: %s", err)
	}
	cfg := logging.Config{
		Level:  logging.LevelError,
		Stdout: true,
		File:   filepath.Join(repoPath, "powd.log"),
	}
	logging.SetupLogging(cfg)
	loggers := []string{
		// Top-level
		"powd",
		"server",

		// Indexes & Reputation
		"index-miner",
		"index-ask",
		"index-faults",
		"reputation",
		"reputation-source-store",
		"chainstore",
		"fchost",
		"maxmind",

		// Lotus client
		"lotus-client",

		// Deals Module
		"deals",

		// Wallet Module
		"wallet",

		// FFS
		"ffs-scheduler",
		"ffs-manager",
		"ffs-auth",
		"ffs-api",
		"ffs-coreipfs",
		"ffs-grpc-service",
		"ffs-filcold",
		"ffs-sched-sjstore",
		"ffs-sched-rjstore",
		"ffs-sched-astore",
		"ffs-cidlogger",
	}

	// powd registered loggers get info level by default.
	for _, l := range loggers {
		if err := logging.SetLogLevel(l, "info"); err != nil {
			return fmt.Errorf("setting up logger %s: %s", l, err)
		}
	}
	debugLevel := config.GetBool("debug")
	if debugLevel {
		for _, l := range loggers {
			if err := logging.SetLogLevel(l, "debug"); err != nil {
				return err
			}
		}
	}
	return nil
}

func getRepoPath(devnet bool) (string, error) {
	if devnet {
		repoPath, err := ioutil.TempDir("", ".powergate-*")
		if err != nil {
			return "", fmt.Errorf("generating temp for repo folder: %s", err)
		}
		return repoPath, nil
	}
	repoPath := config.GetString("repopath")
	if repoPath == "~/.powergate" {
		expandedPath, err := homedir.Expand(repoPath)
		if err != nil {
			log.Fatalf("expanding homedir: %s", err)
		}
		repoPath = expandedPath
	}
	return repoPath, nil
}

func getLotusToken(devnet bool) (string, error) {
	// If running in devnet, there's no need for Lotus API auth token.
	if devnet {
		return "", nil
	}

	token := config.GetString("lotustoken")
	if token != "" {
		return token, nil
	}

	path := config.GetString("lotustokenfile")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("lotus auth token can't be empty")
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading token file from lotus")
	}
	return string(b), nil
}

func setupFlags() error {
	pflag.Bool("debug", false, "Enable debug log level in all loggers.")
	pflag.String("grpchostaddr", "/ip4/0.0.0.0/tcp/5002", "gRPC host listening address.")
	pflag.String("grpcwebproxyaddr", "0.0.0.0:6002", "gRPC webproxy listening address.")
	pflag.String("lotushost", "/ip4/127.0.0.1/tcp/1234", "Lotus client API endpoint multiaddress.")
	pflag.String("lotustoken", "", "Lotus API authorization token. This flag or --lotustoken file are mandatory.")
	pflag.String("lotustokenfile", "", "Path of a file that contains the Lotus API authorization token.")
	pflag.String("lotusmasteraddr", "", "Existing wallet address in Lotus to be used as source of funding for new FFS instances. (Optional)")
	pflag.Bool("autocreatemasteraddr", false, "Automatically creates & funds a master address if none is provided")
	pflag.Bool("ffsusemasteraddr", false, "Use the master address as the initial address for all new FFS instances instead of creating a new unique addess for each new FFS instance.")
	pflag.String("repopath", "~/.powergate", "Path of the repository where Powergate state will be saved.")
	pflag.Bool("devnet", false, "Indicate that will be running on an ephemeral devnet. --repopath will be autocleaned on exit.")
	pflag.String("ipfsapiaddr", "/ip4/127.0.0.1/tcp/5001", "IPFS API endpoint multiaddress. (Optional, only needed if FFS is used)")
	pflag.Int64("walletinitialfund", 4000000000000000, "FFS initial funding transaction amount in attoFIL received by --lotusmasteraddr. (if set)")
	pflag.String("gatewayhostaddr", "0.0.0.0:7000", "Gateway host listening address")
	pflag.String("gatewaybasepath", "/", "Gateway base path")
	pflag.String("maxminddbfolder", ".", "Path of the folder containing GeoLite2-City.mmdb")
	pflag.Parse()

	config.SetEnvPrefix("POWD")
	config.AutomaticEnv()
	if err := config.BindPFlags(pflag.CommandLine); err != nil {
		return fmt.Errorf("binding pflags: %s", err)
	}
	return nil
}
