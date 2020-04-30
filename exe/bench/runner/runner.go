package runner

import (
	"context"
	"fmt"

	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/api/client"
	"github.com/textileio/powergate/health"
	"google.golang.org/grpc"
)

type TestSetup struct {
	LotusAddr    multiaddr.Multiaddr
	SampleSize   int
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
