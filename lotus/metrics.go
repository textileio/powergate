package lotus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/lotus/api"
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

type LotusSyncMonitor struct {
	cb ClientBuilder

	lock       sync.Mutex
	heightDiff int64
	remaining  int64
}

// MonitorLotusSync fires a goroutine that will generate
// metrics with Lotus node height.
func NewLotusSyncMonitor(cb ClientBuilder) (*LotusSyncMonitor, error) {
	if err := view.Register(vHeight); err != nil {
		log.Fatalf("register metrics views: %v", err)
	}
	lsm := &LotusSyncMonitor{cb: cb}

	if err := lsm.refreshSyncDiff(); err != nil {
		return nil, fmt.Errorf("getting initial sync height diff: %s", err)
	}

	go func() {
		for {
			lsm.evaluate()
			time.Sleep(metricHeightInterval)
		}
	}()
	return lsm, nil
}

func (lsm *LotusSyncMonitor) SyncHeightDiff() int64 {
	lsm.lock.Lock()
	defer lsm.lock.Unlock()
	return lsm.heightDiff
}

func (lsm *LotusSyncMonitor) evaluate() {
	if err := lsm.refreshHeightMetric(); err != nil {
		log.Errorf("refreshing height metric: %s", err)
	}
	if err := lsm.checkSyncStatus(); err != nil {
		log.Errorf("checking sync status: %s", err)
	}
}

func (lsm *LotusSyncMonitor) refreshHeightMetric() error {
	c, cls, err := lsm.cb(context.Background())
	if err != nil {
		return fmt.Errorf("creating lotus client for monitoring: %s", err)
	}
	defer cls()

	heaviest, err := c.ChainHead(context.Background())
	if err != nil {
		return fmt.Errorf("getting chain head: %s", err)
	}
	stats.Record(context.Background(), mLotusHeight.M(int64(heaviest.Height())))
	return nil
}

func (lsm *LotusSyncMonitor) checkSyncStatus() error {
	if err := lsm.refreshSyncDiff(); err != nil {
		return fmt.Errorf("refreshing height difference: %s", err)
	}

	lsm.lock.Lock()
	defer lsm.lock.Unlock()
	if lsm.heightDiff > 10 {
		log.Warnf("Louts behind in syncing with height diff %d, todo: %d", lsm.heightDiff, lsm.remaining)
	}
	return nil
}

func (lsm *LotusSyncMonitor) refreshSyncDiff() error {
	c, cls, err := lsm.cb(context.Background())
	if err != nil {
		return fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	ctx, cls := context.WithTimeout(context.Background(), time.Second*5)
	defer cls()
	ss, err := c.SyncState(ctx)
	if err != nil {
		return fmt.Errorf("calling sync state: %s", err)
	}

	var maxHeightDiff, remaining int64
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

			if heightDiff > maxHeightDiff {
				maxHeightDiff = heightDiff
				remaining = int64(as.Target.Height() - as.Height)
			}
		}
	}

	lsm.lock.Lock()
	lsm.heightDiff = maxHeightDiff
	lsm.remaining = remaining
	lsm.lock.Unlock()

	return nil
}
