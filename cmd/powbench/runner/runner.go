package runner

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/api/client"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/health"
	"github.com/textileio/powergate/util"
)

var (
	log = logging.Logger("runner")
)

// TestSetup describes a test configuration.
type TestSetup struct {
	PowergateAddr string
	MinerAddr     string
	SampleSize    int64
	MaxParallel   int
	TotalSamples  int
	RandSeed      int
}

// Run runs a test setup.
func Run(ctx context.Context, ts TestSetup) error {
	c, err := client.NewClient(ts.PowergateAddr)
	defer func() {
		if err := c.Close(); err != nil {
			log.Errorf("closing powergate client: %s", err)
		}
	}()
	if err != nil {
		return fmt.Errorf("creating client: %s", err)
	}

	if err := sanityCheck(ctx, c); err != nil {
		return fmt.Errorf("sanity check with client: %s", err)
	}

	if err := runSetup(ctx, c, ts); err != nil {
		return fmt.Errorf("running test setup: %s", err)
	}

	return nil
}

func sanityCheck(ctx context.Context, c *client.Client) error {
	s, _, err := c.Health.Check(ctx)
	if err != nil {
		return fmt.Errorf("health check call: %s", err)
	}
	if s != health.Ok {
		return fmt.Errorf("reported health check not Ok: %s", s)
	}
	return nil
}

func runSetup(ctx context.Context, c *client.Client, ts TestSetup) error {
	_, tok, err := c.FFS.Create(ctx)
	if err != nil {
		return fmt.Errorf("creating ffs instance: %s", err)
	}
	ctx = context.WithValue(ctx, client.AuthKey, tok)
	info, err := c.FFS.Info(ctx)
	if err != nil {
		return fmt.Errorf("getting instance info: %s", err)
	}
	addr := info.Balances[0].Addr
	time.Sleep(time.Second * 5)

	chLimit := make(chan struct{}, ts.MaxParallel)
	chErr := make(chan error, ts.TotalSamples)
	for i := 0; i < ts.TotalSamples; i++ {
		chLimit <- struct{}{}
		go func(i int) {
			defer func() { <-chLimit }()
			if err := run(ctx, c, i, ts.RandSeed+i, ts.SampleSize, addr, ts.MinerAddr); err != nil {
				chErr <- fmt.Errorf("failed run %d: %s", i, err)
			}
		}(i)
	}
	for i := 0; i < ts.MaxParallel; i++ {
		chLimit <- struct{}{}
	}
	close(chErr)
	for err := range chErr {
		return fmt.Errorf("sample run errored: %s", err)
	}
	return nil
}

func run(ctx context.Context, c *client.Client, id int, seed int, size int64, addr string, minerAddr string) error {
	log.Infof("[%d] Executing run...", id)
	defer log.Infof("[%d] Done", id)
	ra := rand.New(rand.NewSource(int64(seed)))
	lr := io.LimitReader(ra, size)

	log.Infof("[%d] Adding to hot layer...", id)
	ci, err := c.FFS.Stage(ctx, lr)
	if err != nil {
		return fmt.Errorf("importing data to hot storage (ipfs node): %s", err)
	}

	log.Infof("[%d] Pushing %s to FFS...", id, *ci)

	// For completeness, fields that could be relied on defaults
	// are explicitly kept here to have a better idea about their
	// existence.
	// This configuration will stop being static when we incorporate
	// other test cases.
	storageConfig := ffs.StorageConfig{
		Repairable: false,
		Hot: ffs.HotConfig{
			Enabled:          true,
			AllowUnfreeze:    false,
			UnfreezeMaxPrice: 0,
			Ipfs: ffs.IpfsConfig{
				AddTimeout: 30,
			},
		},
		Cold: ffs.ColdConfig{
			Enabled: true,
			Filecoin: ffs.FilConfig{
				RepFactor:       1,
				DealMinDuration: util.MinDealDuration,
				Addr:            addr,
				CountryCodes:    nil,
				ExcludedMiners:  nil,
				TrustedMiners:   []string{minerAddr},
				Renew:           ffs.FilRenew{},
			},
		},
	}

	jid, err := c.FFS.PushStorageConfig(ctx, *ci, client.WithStorageConfig(storageConfig))
	if err != nil {
		return fmt.Errorf("pushing to FFS: %s", err)
	}

	log.Infof("[%d] Pushed successfully, queued job %s. Waiting for termination...", id, jid)
	chJob := make(chan client.JobEvent, 1)
	ctxWatch, cancel := context.WithCancel(ctx)
	defer cancel()
	err = c.FFS.WatchJobs(ctxWatch, chJob, jid)
	if err != nil {
		return fmt.Errorf("opening listening job status: %s", err)
	}
	var s client.JobEvent
	for s = range chJob {
		if s.Err != nil {
			return fmt.Errorf("job watching: %s", s.Err)
		}
		log.Infof("[%d] Job changed to status %s", id, ffs.JobStatusStr[s.Job.Status])
		if s.Job.Status == ffs.Failed || s.Job.Status == ffs.Canceled {
			return fmt.Errorf("job execution failed or was canceled")
		}
		if s.Job.Status == ffs.Success {
			return nil
		}
	}
	return fmt.Errorf("unexpected Job status watcher")
}
