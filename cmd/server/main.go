package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"contrib.go.opencensus.io/exporter/prometheus"
	logging "github.com/ipfs/go-log"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/filecoin/api"
	"github.com/textileio/filecoin/tests"
)

var (
	grpcHostAddr = "/ip4/127.0.0.1/tcp/50051"

	log = logging.Logger("main")
)

func main() {
	logging.SetDebugLogging()
	logging.SetLogLevel("deals", "warn")
	instrumentationSetup()

	lotusAddr, _ := tests.ClientConfigMA()
	grpcAddr, err := ma.NewMultiaddr(grpcHostAddr)
	token, ok := os.LookupEnv("TEXTILE_LOTUS_TOKEN")
	if !ok {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		token, err = tests.GetLotusToken(home)
		if err != nil {
			panic(err)
		}
	}
	conf := api.Config{
		LotusAddress:    lotusAddr,
		LotusAuthToken:  token,
		GrpcHostAddress: grpcAddr,
	}
	log.Info("starting server...")
	server, err := api.NewServer(conf)
	if err != nil {
		panic(err)
	}
	log.Info("server started.")

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	<-s
	log.Info("Closing...")
	server.Close()
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
