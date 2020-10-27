package lotus

import (
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	mLotusHeight = stats.Int64("lotus/height", "Height of Lotus node", "By")
	vHeight      = &view.View{
		Name:        "lotus/height_count",
		Measure:     mLotusHeight,
		Description: "Current height of Lotus node",
		Aggregation: view.LastValue(),
	}

	metricHeightInterval = time.Second * 120
)

// MonitorLotusSync fires a goroutine that will generate
// metrics with Lotus node height.
func MonitorLotusSync(clientBuilder ClientBuilder) {
	if err := view.Register(vHeight); err != nil {
		log.Fatalf("register metrics views: %v", err)
	}
	go func() {
		for {
			evaluate(clientBuilder)
			time.Sleep(metricHeightInterval)
		}
	}()
}

func evaluate(cb ClientBuilder) {
	c, cls, err := cb()
	if err != nil {
		log.Errorf("creating lotus client for monitoring: %s", err)
		return
	}
	defer cls()
	if err := refreshHeightMetric(c); err != nil {
		log.Errorf("refreshing height metric: %s", err)
	}
	if err := checkSyncStatus(c); err != nil {
		log.Errorf("checking sync status: %s", err)
	}
}

func refreshHeightMetric(c *apistruct.FullNodeStruct) error {
	heaviest, err := c.ChainHead(context.Background())
	if err != nil {
		return fmt.Errorf("getting chain head: %s", err)
	}
	stats.Record(context.Background(), mLotusHeight.M(int64(heaviest.Height())))
	return nil
}

func checkSyncStatus(c *apistruct.FullNodeStruct) error {
	ctx, cls := context.WithTimeout(context.Background(), time.Second*5)
	defer cls()
	ss, err := c.SyncState(ctx)
	if err != nil {
		return fmt.Errorf("calling sync state: %s", err)
	}
	for _, as := range ss.ActiveSyncs {
		if as.Stage != api.StageSyncComplete {
			var heightDiff int64
			if as.Base != nil {
				heightDiff = int64(as.Base.Height())
			}
			if as.Target != nil {
				heightDiff = int64(as.Target.Height()) - heightDiff
			} else {
				heightDiff = 0
			}

			if heightDiff > 10 {
				log.Warnf("Louts node behind with height difference %d", heightDiff)
				return nil // Only report one worker falling behind
			}
		}

	}
	return nil
}
