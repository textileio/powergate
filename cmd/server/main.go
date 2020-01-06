package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	"github.com/textileio/filecoin/deals"
	"github.com/textileio/filecoin/lotus"
	"github.com/textileio/filecoin/tests"
	"github.com/textileio/filecoin/wallet"
)

var (
	log = logging.Logger("main")
)

func main() {
	logging.SetLogLevel("deals", "warn")

	addr := "127.0.0.1:1234"

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	token, ok := os.LookupEnv("TEXTILE_LOTUS_TOKEN")
	if !ok {
		token, err = tests.GetLotusToken(home)
		if err != nil {
			panic(err)
		}
	}
	c, cls, err := lotus.New(addr, token)
	if err != nil {
		panic(err)
	}
	defer cls()
	dm := deals.New(c, datastore.NewMapDatastore())
	wm := wallet.New(c)

	instrumentationSetup()

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	<-s
	fmt.Println("Closing...")
	dm.Close()
	wm.Close()
	fmt.Println("Closed")
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
