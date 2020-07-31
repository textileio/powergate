package runner

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/api/client"
	"github.com/textileio/powergate/health"
)

var (
	powergateAddr = "127.0.0.1:5002"
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelInfo)
	os.Exit(m.Run())
}

func TestSimpleSetup(t *testing.T) {
	// Explicitly skip this test since its meant for benchmarking stuff.
	t.SkipNow()
	_ = spinup(t)
	ts := TestSetup{
		PowergateAddr: powergateAddr,
		MinerAddr:     "t01000",
		SampleSize:    700,
		MaxParallel:   1,
		TotalSamples:  1,
		RandSeed:      22,
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
		err := cmd.Run()
		require.NoError(t, err)
	}
	makeDown()

	cmd := exec.Command("docker-compose", "-f", "docker-compose-devnet.yaml", "build")
	cmd.Dir = dockerFolder
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("docker-compose build: %s", err)
	}

	cmd = exec.Command("make", "devnet")
	cmd.Dir = dockerFolder
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("running docker-compose: %s", err)
	}
	t.Cleanup(makeDown)

	var c *client.Client
	var err error
	limit := 30
	retries := 0
	for retries < limit {
		c, err = client.NewClient(powergateAddr)
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
