package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"contrib.go.opencensus.io/exporter/prometheus"
	logging "github.com/ipfs/go-log/v2"
	homedir "github.com/mitchellh/go-homedir"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/server"
	"github.com/textileio/powergate/util"
)

var (
	log    = logging.Logger("powd")
	config = viper.New()
)

func main() {
	pflag.Bool("debug", false, "enable debug log levels")
	pflag.Bool("pprof", false, "enable pprof endpoint")
	pflag.String("grpchostaddr", "0.0.0.0:5002", "grpc host listening address")
	pflag.String("grpcwebproxyaddr", "0.0.0.0:6002", "grpc webproxy listening address")
	pflag.String("lotushost", "/ip4/127.0.0.1/tcp/1234", "lotus full-node address")
	pflag.String("lotustoken", "", "lotus full-node auth token")
	pflag.String("lotustokenfile", "", "lotus full-node auth token file")
	pflag.String("lotusmasteraddr", "", "lotus addr to be considered master for ffs")
	pflag.String("repopath", "~/.powergate", "repo-path")
	pflag.Bool("embedded", false, "run in embedded ephemeral FIL network")
	pflag.String("ipfsapiaddr", "/ip4/127.0.0.1/tcp/5001", "ipfs api multiaddr")
	pflag.Int64("walletinitialfund", 4000000000, "created wallets initial fund in attoFIL")
	pflag.String("gatewayhostaddr", "0.0.0.0:7000", "gateway host listening address")
	pflag.Parse()

	config.SetEnvPrefix("TEXPOWERGATE")
	config.AutomaticEnv()
	if err := config.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatalf("binding pflags: %s", err)
	}

	if err := setupLogging(); err != nil {
		log.Fatalf("setting up logging: %s", err)
	}
	setupInstrumentation()
	setupPprof()

	embedded := config.GetBool("embedded")

	var lotusToken, repoPath string
	var err error
	maddr, err := getLotusMaddr()
	if err != nil {
		log.Fatal(err)
	}
	if !embedded {
		lotusToken, err = getLotusToken()
		if err != nil {
			log.Fatal(err)
		}

		repoPath = config.GetString("repopath")
		if repoPath == "~/.powergate" {
			expandedPath, err := homedir.Expand(repoPath)
			if err != nil {
				log.Fatalf("expanding homedir: %s", err)
			}
			repoPath = expandedPath
		}
	} else {
		repoPath, err = ioutil.TempDir("/tmp/powergate", ".powergate-*")
		if err != nil {
			log.Fatal(err)
		}
	}

	conf := server.Config{
		WalletInitialFunds: *big.NewInt(config.GetInt64("walletinitialfund")),
		IpfsAPIAddr:        util.MustParseAddr(config.GetString("ipfsapiaddr")),
		LotusAddress:       maddr,
		LotusAuthToken:     lotusToken,
		LotusMasterAddr:    config.GetString("lotusmasteraddr"),
		Embedded:           embedded,
		// ToDo: Support secure gRPC connection
		GrpcHostNetwork:     "tcp",
		GrpcHostAddress:     config.GetString("grpchostaddr"),
		GrpcWebProxyAddress: config.GetString("grpcwebproxyaddr"),
		RepoPath:            repoPath,
		GatewayHostAddr:     config.GetString("gatewayhostaddr"),
	}
	confJSON, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		log.Fatalf("can't show current config: %s", err)
	}
	log.Infof("Current configuration: \n%s", confJSON)
	log.Info("starting server...")
	s, err := server.NewServer(conf)
	if err != nil {
		log.Errorf("error starting server: %s", err)
		os.Exit(-1)
	}
	log.Info("server started.")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	log.Info("Closing...")
	s.Close()
	if embedded {
		if err := os.RemoveAll(repoPath); err != nil {
			log.Error(err)
		}
	}
	log.Info("Closed")
}

func setupInstrumentation() {
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "textilefc",
	})
	if err != nil {
		log.Fatalf("Failed to create the Prometheus stats exporter: %v", err)
	}
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		if err := http.ListenAndServe(":8888", mux); err != nil {
			log.Fatalf("Failed to run Prometheus scrape endpoint: %v", err)
		}
	}()
}

func setupLogging() error {
	if err := logging.SetLogLevel("*", "error"); err != nil {
		return err
	}
	loggers := []string{"index-miner", "index-ask", "index-slashing",
		"server", "deals", "powergated", "fchost", "ip2location", "reputation",
		"ffs-scheduler", "ffs-manager", "ffs-auth", "ffs-api", "ffs-coreipfs",
		"ffs-grpc-service", "ffs-filcold", "ffs-sched-cistore", "ffs-sched-jstore",
		"ffs-sched-pcstore", "ffs-cidlogger"}
	for _, l := range loggers {
		if err := logging.SetLogLevel(l, "info"); err != nil {
			return err
		}
	}
	if config.GetBool("debug") {
		for _, l := range loggers {
			if err := logging.SetLogLevel(l, "debug"); err != nil {
				return err
			}
		}
	}
	return nil
}

func setupPprof() {
	if !config.GetBool("pprof") {
		return
	}
	go func() {
		log.Error(http.ListenAndServe("localhost:6060", nil))
	}()
}

func getLotusMaddr() (ma.Multiaddr, error) {
	maddr, err := ma.NewMultiaddr(config.GetString("lotushost"))
	if err != nil {
		return nil, err
	}
	return maddr, nil
}

func getLotusToken() (string, error) {
	token := config.GetString("lotustoken")
	if token == "" {
		path := config.GetString("lotustokenfile")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return "", fmt.Errorf("lotus auth token can't be empty")
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("reading token file from lotus")
		}
		token = string(b)
	}
	return token, nil
}
