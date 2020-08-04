package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	logging "github.com/ipfs/go-log/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/buildinfo"
	"github.com/textileio/powergate/cmd/powbench/runner"
)

const (
	cmdPowergateAddr = "pgAddr"
	cmdSampleSize    = "sampleSize"
	cmdMaxParallel   = "maxParallel"
	cmdTotalSamples  = "totalSamples"
	cmdRandSeed      = "randSeed"
	cmdMinerAddr     = "minerAddr"
)

var (
	log    = logging.Logger("main")
	config = viper.New()
)

func main() {
	log.Infof("starting powbench:\n%s", buildinfo.Summary())

	pflag.String(cmdPowergateAddr, "/ip4/127.0.0.1/tcp/5002", "Powergate server multiaddress")
	pflag.Int(cmdSampleSize, 1024, "Size of randomly generated files in bytes")
	pflag.Int(cmdMaxParallel, 1, "Max parallel file storage")
	pflag.Int(cmdTotalSamples, 3, "Total samples to run")
	pflag.Int(cmdRandSeed, 42, "Random seed used to generate random samples data")
	pflag.String(cmdMinerAddr, "t01000", "Miner address to force Powergate to select for making deals")
	pflag.Parse()

	config.SetEnvPrefix("POWBENCH")
	config.AutomaticEnv()
	if err := config.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatalf("binding flags: %s", err)
	}

	pgAddr := config.GetString(cmdPowergateAddr)
	ts := runner.TestSetup{
		PowergateAddr: pgAddr,
		MinerAddr:     config.GetString(cmdMinerAddr),

		SampleSize:   config.GetInt64(cmdSampleSize),
		MaxParallel:  config.GetInt(cmdMaxParallel),
		TotalSamples: config.GetInt(cmdTotalSamples),
		RandSeed:     config.GetInt(cmdRandSeed),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		if err := runner.Run(ctx, ts); err != nil {
			log.Fatalf("running test setup: %s", err)
		}
	}(ctx)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	log.Info("Closing...")
	cancel()
	wg.Wait()
	log.Info("Done")
}
