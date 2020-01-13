package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"contrib.go.opencensus.io/exporter/prometheus"
	logging "github.com/ipfs/go-log"
	"github.com/textileio/filecoin/api/server"
	"github.com/textileio/filecoin/tests"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	grpcHostAddr = "127.0.0.1:5002"

	log = logging.Logger("main")
)

func main() {
	logging.SetDebugLogging()
	logging.SetLogLevel("*", "error")
	instrumentationSetup()

	lotusAddr, token := tests.ClientConfigMA()
	if err != nil {
		panic(err)
	}
	conf := server.Config{
		LotusAddress:    lotusAddr,
		LotusAuthToken:  token,
		GrpcHostNetwork: "tcp",
		GrpcHostAddress: grpcHostAddr,
	}
	log.Info("starting server...")
	s, err := server.NewServer(conf)
	if err != nil {
		panic(err)
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
