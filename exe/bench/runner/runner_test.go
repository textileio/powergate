package runner

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/api/client"
	"github.com/textileio/powergate/health"
	"google.golang.org/grpc"
)

var (
	lotusAddr = multiaddr.StringCast("/ip4/127.0.0.1/tcp/5002")
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelInfo)
	os.Exit(m.Run())
}

func TestSimpleSetup(t *testing.T) {
	// Explicitly skip this test since its meant for benchmarking stuff.
	//t.SkipNow()
	_ = spinup(t)
	ts := TestSetup{
		LotusAddr:    lotusAddr,
		SampleSize:   700,
		MaxParallel:  1,
		TotalSamples: 1,
		RandSeed:     22,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	err := Run(ctx, ts)
	require.NoError(t, err)
}

func spinup(t *testing.T) *client.Client {
	dockerFolder := "../../../docker"

	makeDown := func() {
		cmd := exec.Command("make", "down")
		cmd.Dir = dockerFolder
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}
	makeDown()

	cmd := exec.Command("docker-compose", "-f", "docker-compose-embedded.yaml", "build")
	cmd.Dir = dockerFolder
	if err := cmd.Run(); err != nil {
		t.Fatalf("docker-compose build: %s", err)
	}

	cmd = exec.Command("make", "embed")
	cmd.Dir = dockerFolder
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("running docker-compose: %s", err)
	}
	t.Cleanup(makeDown)

	var c *client.Client
	var err error
	limit := 30
	retries := 0
	for retries < limit {
		c, err = client.NewClient(lotusAddr, grpc.WithInsecure())
		require.NoError(t, err)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		s, _, err := c.Health.Check(ctx)
		if err == nil {
			require.Equal(t, health.Ok, s)
			break
		}
		time.Sleep(time.Second)
		retries++
	}
	if retries == limit {
		t.Fatalf("failed to connect to powergate")
	}
	// After PG is up, wait a bit more to ensure IPFS does to.
	time.Sleep(time.Second * 5)
	return c
}
