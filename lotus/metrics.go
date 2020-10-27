package lotus

import (
	"context"
	"time"

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

	metricHeightInterval = time.Second * 10
)

// MonitorLotusSync fires a goroutine that will generate
// metrics with Lotus node height.
func MonitorLotusSync(clientBuilder ClientBuilder) {
	if err := view.Register(vHeight); err != nil {
		log.Fatalf("register metrics views: %v", err)
	}
	go func() {
		for {
			refreshHeightMetric(clientBuilder)
			time.Sleep(metricHeightInterval)
		}
	}()
}

func refreshHeightMetric(clientBuilder ClientBuilder) {
	c, cls, err := clientBuilder()
	if err != nil {
		log.Errorf("creating lotus client for monitoring: %s", err)
		return
	}
	defer cls()
	heaviest, err := c.ChainHead(context.Background())
	if err != nil {
		log.Errorf("get lotus sync status: %s", err)
		return
	}
	stats.Record(context.Background(), mLotusHeight.M(int64(heaviest.Height())))
}
