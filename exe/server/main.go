package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "net/http/pprof"

	"contrib.go.opencensus.io/exporter/prometheus"
	logging "github.com/ipfs/go-log"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/textileio/fil-tools/api/server"
)

var (
	log    = logging.Logger("fil-toolsd")
	config = viper.New()
)

func main() {
	pflag.Bool("debug", false, "enable debug log levels")
	pflag.Bool("pprof", false, "enable pprof endpoint")
	pflag.String("grpchostaddr", "0.0.0.0:5002", "grpc host listening address")
	pflag.String("grpcwebproxyaddr", "0.0.0.0:6002", "grpc webproxy listening address")
	pflag.String("lotushost", "127.0.0.1:1234", "lotus full-node address")
	pflag.String("lotustoken", "", "lotus full-node auth token")
	pflag.String("repopath", "${HOME}/.texfc", "repo-path")
	pflag.Parse()

	config.SetEnvPrefix("TEXFILTOOLS")
	config.AutomaticEnv()
	config.BindPFlags(pflag.CommandLine)

	setupLogging()
	setupInstrumentation()
	setupPprof()

	maddr, err := getLotusMaddr()
	if err != nil {
		log.Fatal(err)
	}
	lotusToken, err := getLotusToken()
	if err != nil {
		log.Fatal(err)
	}

	repoPath := config.GetString("repopath")
	if repoPath == "" {
		repoPath, err = os.UserHomeDir()
		if err != nil {
			log.Fatal("error when setting default repo path to home dir")
		}
	}

	conf := server.Config{
		LotusAddress:   maddr,
		LotusAuthToken: lotusToken,
		// ToDo: Support secure gRPC connection
		GrpcHostNetwork:     "tcp",
		GrpcHostAddress:     config.GetString("grpchostaddr"),
		GrpcWebProxyAddress: config.GetString("grpcwebproxyaddr"),
		RepoPath:            repoPath,
	}
	confJson, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		log.Fatalf("can't show current config: %s", err)
	}
	log.Infof("Current configuration: \n%s", confJson)
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

func setupLogging() {
	logging.SetDebugLogging()
	if !config.GetBool("debug") {
		logging.SetLogLevel("*", "info")
		logging.SetLogLevel("rpc", "error")
		logging.SetLogLevel("dht", "error")
		logging.SetLogLevel("swarm2", "error")
	}
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
	addrsplt := strings.Split(config.GetString("lotushost"), ":")
	if len(addrsplt) != 2 {
		return nil, fmt.Errorf("lotus addr is invalid")
	}
	addr := fmt.Sprintf("/ip4/%v/tcp/%v", addrsplt[0], addrsplt[1])
	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return nil, err
	}
	return maddr, nil
}

func getLotusToken() (string, error) {
	token := config.GetString("lotustoken")
	if token == "" {
		return "", fmt.Errorf("lotus auth token can't be empty")
	}
	return token, nil
}
