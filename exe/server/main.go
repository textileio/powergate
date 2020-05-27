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
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	logging "github.com/ipfs/go-log/v2"
	homedir "github.com/mitchellh/go-homedir"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/server"
	"github.com/textileio/powergate/util"
	"go.opencensus.io/plugin/runmetrics"
)

var (
	log    = logging.Logger("powd")
	config = viper.New()
)

func main() {
	pflag.Bool("debug", false, "Enable debug log level in all loggers.")
	pflag.String("grpchostaddr", "/ip4/0.0.0.0/tcp/5002", "gRPC host listening address.")
	pflag.String("grpcwebproxyaddr", "0.0.0.0:6002", "gRPC webproxy listening address.")
	pflag.String("lotushost", "/ip4/127.0.0.1/tcp/1234", "Lotus client API endpoint multiaddress.")
	pflag.String("lotustoken", "", "Lotus API authorization token. This flag or --lotustoken file are mandatory.")
	pflag.String("lotustokenfile", "", "Path of a file that contains the Lotus API authorization token.")
	pflag.String("lotusmasteraddr", "", "Existing wallet address in Lotus to be used as source of funding for new FFS instances. (Optional)")
	pflag.String("repopath", "~/.powergate", "Path of the repository where Powergate state will be saved.")
	pflag.Bool("devnet", false, "Indicate that will be running on an ephemeral devnet. --repopath will be autocleaned on exit.")
	pflag.String("ipfsapiaddr", "/ip4/127.0.0.1/tcp/5001", "IPFS API endpoint multiaddress. (Optional, only needed if FFS is used)")
	pflag.Int64("walletinitialfund", 4000000000000000, "FFS initial funding transaction amount in attoFIL received by --lotusmasteraddr. (if set)")
	pflag.String("gatewayhostaddr", "0.0.0.0:7000", "Gateway host listening address")
	pflag.Parse()

	config.SetEnvPrefix("POWD")
	config.AutomaticEnv()
	if err := config.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatalf("binding pflags: %s", err)
	}

	if err := setupLogging(); err != nil {
		log.Fatalf("setting up logging: %s", err)
	}
	prometheusServer := setupInstrumentation()

	devnet := config.GetBool("devnet")

	var lotusToken, repoPath string
	var err error
	maddr, err := getLotusMaddr()
	if err != nil {
		log.Fatal(err)
	}
	if !devnet {
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

	grpcHostMaddr, err := ma.NewMultiaddr(config.GetString("grpchostaddr"))
	if err != nil {
		log.Fatalf("parsing grpchostaddr: %s", err)
	}

	conf := server.Config{
		WalletInitialFunds: *big.NewInt(config.GetInt64("walletinitialfund")),
		IpfsAPIAddr:        util.MustParseAddr(config.GetString("ipfsapiaddr")),
		LotusAddress:       maddr,
		LotusAuthToken:     lotusToken,
		LotusMasterAddr:    config.GetString("lotusmasteraddr"),
		Devnet:             devnet,
		// ToDo: Support secure gRPC connection
		GrpcHostNetwork:     "tcp",
		GrpcHostAddress:     grpcHostMaddr,
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := prometheusServer.Shutdown(ctx); err != nil {
		log.Error("shutting down prometheus server: %s", err)
	}
	s.Close()
	if devnet {
		if err := os.RemoveAll(repoPath); err != nil {
			log.Error(err)
		}
	}
	log.Info("Closed")
}

func setupInstrumentation() *http.Server {
	err := runmetrics.Enable(runmetrics.RunMetricOptions{
		EnableCPU:    true,
		EnableMemory: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "textilefc",
	})
	if err != nil {
		log.Fatalf("Failed to create the Prometheus stats exporter: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", pe)
	srv := &http.Server{Addr: ":8888", Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to run Prometheus scrape endpoint: %v", err)
		}
	}()
	return srv
}

func setupLogging() error {
	if err := logging.SetLogLevel("*", "error"); err != nil {
		return err
	}
	loggers := []string{"index-miner", "index-ask", "index-slashing", "chainstore",
		"server", "deals", "powd", "fchost", "ip2location", "reputation",
		"reputation-source-store", "ffs-scheduler", "ffs-manager", "ffs-auth",
		"ffs-api", "ffs-api-istore", "ffs-coreipfs", "ffs-grpc-service", "ffs-filcold",
		"ffs-sched-jstore", "ffs-sched-astore", "ffs-cidlogger"}
	for _, l := range loggers {
		if err := logging.SetLogLevel(l, "info"); err != nil {
			return fmt.Errorf("setting up logger %s: %s", l, err)
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
