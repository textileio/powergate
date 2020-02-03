package main

import (
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"contrib.go.opencensus.io/exporter/prometheus"
	logging "github.com/ipfs/go-log"
	"github.com/textileio/fil-tools/api/server"
	"github.com/textileio/fil-tools/tests"
	_ "net/http/pprof"
)

var (
	grpcHostAddr = "127.0.0.1:5002"

	log = logging.Logger("main")
)

func main() {
	logging.SetDebugLogging()
	logging.SetLogLevel("*", "info")
	logging.SetLogLevel("rpc", "error")
	logging.SetLogLevel("dht", "error")
	logging.SetLogLevel("swarm2", "error")

	instrumentationSetup()
	pprofSetup()

	// ToDo: Flags for configuration

	lotusAddr, token := tests.ClientConfigMA()
	repoPath, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("error getting home dir: %s", err)
		os.Exit(-1)
	}
	conf := server.Config{
		LotusAddress:    lotusAddr,
		LotusAuthToken:  token,
		GrpcHostNetwork: "tcp",
		GrpcHostAddress: grpcHostAddr,
		RepoPath:        filepath.Join(repoPath, ".texfc"),
	}
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

func instrumentationSetup() {
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

func pprofSetup() {
	go func() {
		log.Error(http.ListenAndServe("localhost:6060", nil))
	}()
}
