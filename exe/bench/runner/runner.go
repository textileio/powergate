package runner

import (
	"context"
	"fmt"
	"io"
	"math/rand"

	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/api/client"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/health"
	"google.golang.org/grpc"
)

var (
	log = logging.Logger("runner")
)

type TestSetup struct {
	LotusAddr    multiaddr.Multiaddr
	SampleSize   int64
	MaxParallel  int
	TotalSamples int
	RandSeed     int
}

func Run(ctx context.Context, ts TestSetup) error {
	c, err := client.NewClient(ts.LotusAddr, grpc.WithInsecure())
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
	chLimit := make(chan struct{}, ts.MaxParallel)
	chErr := make(chan error, ts.TotalSamples)
	for i := 0; i < ts.TotalSamples; i++ {
		chLimit <- struct{}{}
		go func(i int) {
			defer func() { <-chLimit }()
			if err := run(ctx, c, i, ts.RandSeed+i, ts.SampleSize); err != nil {
				chErr <- fmt.Errorf("failed run %d: %s", i, err)
			}

		}(i)
	}

	for i := 0; i < ts.MaxParallel; i++ {
		chLimit <- struct{}{}
	}
	return nil
}

func run(ctx context.Context, c *client.Client, id int, seed int, size int64) error {
	ra := rand.New(rand.NewSource(int64(seed)))
	lr := io.LimitReader(ra, size)

	log.Infof("[%d] Adding to hot layer...", id)
	ci, err := c.Ffs.AddToHot(ctx, lr)
	if err != nil {
		return fmt.Errorf("importing data to hot storage (ipfs node): %s", err)
	}

	log.Infof("[%d] Pushing %s to FFS...", *ci, id)
	cidConfig := ffs.CidConfig{
		Cid:        *ci,
		Repairable: false,
		Hot: ffs.HotConfig{
			Enabled:       true,
			AllowUnfreeze: false,
			Ipfs: ffs.IpfsConfig{
				AddTimeout: 30,
			},
		},
		Cold: ffs.ColdConfig{
			Enabled: true,
			Filecoin: ffs.FilConfig{
				RepFactor:      1,
				DealDuration:   1000,
				Addr:           "TODO",
				CountryCodes:   nil,
				ExcludedMiners: nil,
				Renew:          ffs.FilRenew{},
			},
		},
	}

	jid, err := c.Ffs.PushConfig(ctx, *ci, client.WithCidConfig(cidConfig))
	if err != nil {
		return fmt.Errorf("pushing to FFS: %s", err)
	}
	log.Infof("[%d] Pushed successfully, queued job %s", id, jid)

	log.Infof("[%d] Waiting for Job to be executed...", jid)

	chJob := make(chan client.JobEvent, 1)
	ctxWatch, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		err = c.Ffs.WatchJobs(ctxWatch, chJob, jid)
		close(chJob)
	}()
	for s := range chJob {
		if s.Job.Status == ffs.Failed {
			return fmt.Errorf("failed job")
		}
		if s.Job.Status == ffs.Success {
			cancel()
		}
	}
	if err != nil {
		return fmt.Errorf("waiting for job termination: %s", err)
	}

	log.Infof("[%d] Done")
	return nil
}
