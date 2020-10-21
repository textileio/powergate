package runner

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/api/client"
	ffsRpc "github.com/textileio/powergate/ffs/rpc"
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

	if err := runSetup(ctx, c, ts); err != nil {
		return fmt.Errorf("running test setup: %s", err)
	}

	return nil
}

func runSetup(ctx context.Context, c *client.Client, ts TestSetup) error {
	res, err := c.Admin.CreateStorageProfile(ctx)
	if err != nil {
		return fmt.Errorf("creating ffs instance: %s", err)
	}
	ctx = context.WithValue(ctx, client.AuthKey, res.AuthEntry.Token)
	res2, err := c.FFS.Addrs(ctx)
	if err != nil {
		return fmt.Errorf("getting instance info: %s", err)
	}
	addr := res2.Addrs[0].Addr
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
	statgeRes, err := c.FFS.Stage(ctx, lr)
	if err != nil {
		return fmt.Errorf("importing data to hot storage (ipfs node): %s", err)
	}

	log.Infof("[%d] Pushing %s to FFS...", id, statgeRes.Cid)

	// For completeness, fields that could be relied on defaults
	// are explicitly kept here to have a better idea about their
	// existence.
	// This configuration will stop being static when we incorporate
	// other test cases.
	storageConfig := &ffsRpc.StorageConfig{
		Repairable: false,
		Hot: &ffsRpc.HotConfig{
			Enabled:          true,
			AllowUnfreeze:    false,
			UnfreezeMaxPrice: 0,
			Ipfs: &ffsRpc.IpfsConfig{
				AddTimeout: 30,
			},
		},
		Cold: &ffsRpc.ColdConfig{
			Enabled: true,
			Filecoin: &ffsRpc.FilConfig{
				RepFactor:       1,
				DealMinDuration: util.MinDealDuration,
				Addr:            addr,
				CountryCodes:    nil,
				ExcludedMiners:  nil,
				TrustedMiners:   []string{minerAddr},
				Renew:           &ffsRpc.FilRenew{},
			},
		},
	}

	pushRes, err := c.FFS.PushStorageConfig(ctx, statgeRes.Cid, client.WithStorageConfig(storageConfig))
	if err != nil {
		return fmt.Errorf("pushing to FFS: %s", err)
	}

	log.Infof("[%d] Pushed successfully, queued job %s. Waiting for termination...", id, pushRes.JobId)
	chJob := make(chan client.WatchJobsEvent, 1)
	ctxWatch, cancel := context.WithCancel(ctx)
	defer cancel()
	err = c.FFS.WatchJobs(ctxWatch, chJob, pushRes.JobId)
	if err != nil {
		return fmt.Errorf("opening listening job status: %s", err)
	}
	var s client.WatchJobsEvent
	for s = range chJob {
		if s.Err != nil {
			return fmt.Errorf("job watching: %s", s.Err)
		}
		log.Infof("[%d] Job changed to status %s", id, s.Res.Job.Status.String())
		if s.Res.Job.Status == ffsRpc.JobStatus_JOB_STATUS_FAILED || s.Res.Job.Status == ffsRpc.JobStatus_JOB_STATUS_CANCELED {
			return fmt.Errorf("job execution failed or was canceled")
		}
		if s.Res.Job.Status == ffsRpc.JobStatus_JOB_STATUS_SUCCESS {
			return nil
		}
	}
	return fmt.Errorf("unexpected Job status watcher")
}
