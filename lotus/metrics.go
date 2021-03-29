package lotus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/lotus/api"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

const (
	metricHeightInterval = time.Second * 120
)

// SyncMonitor provides information about the Lotus
// syncing status.
type SyncMonitor struct {
	cb ClientBuilder

	lock       sync.Mutex
	height     int64
	heightDiff int64
	remaining  int64
}

// NewSyncMonitor creates a new LotusSyncMonitor.
func NewSyncMonitor(cb ClientBuilder) (*SyncMonitor, error) {
	lsm := &SyncMonitor{cb: cb}
	lsm.initMetrics()

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

// SyncHeightDiff returns the height difference between the tip of the chain
// and the current synced height.
func (lsm *SyncMonitor) SyncHeightDiff() int64 {
	lsm.lock.Lock()
	defer lsm.lock.Unlock()
	return lsm.heightDiff
}

func (lsm *SyncMonitor) evaluate() {
	if err := lsm.refreshHeightMetric(); err != nil {
		log.Errorf("refreshing height metric: %s", err)
	}
	if err := lsm.checkSyncStatus(); err != nil {
		log.Errorf("checking sync status: %s", err)
	}
}

func (lsm *SyncMonitor) refreshHeightMetric() error {
	c, cls, err := lsm.cb(context.Background())
	if err != nil {
		return fmt.Errorf("creating lotus client for monitoring: %s", err)
	}
	defer cls()

	heaviest, err := c.ChainHead(context.Background())
	if err != nil {
		return fmt.Errorf("getting chain head: %s", err)
	}
	lsm.lock.Lock()
	lsm.height = int64(heaviest.Height())
	lsm.lock.Unlock()

	return nil
}

func (lsm *SyncMonitor) checkSyncStatus() error {
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

func (lsm *SyncMonitor) refreshSyncDiff() error {
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

func (lsm *SyncMonitor) initMetrics() {
	meter := global.Meter("powergate")

	_ = metric.Must(meter).NewInt64ValueObserver("powergate.lotus.height",
		func(ctx context.Context, result metric.Int64ObserverResult) {
			lsm.lock.Lock()
			defer lsm.lock.Unlock()
			result.Observe(lsm.height)
		}, metric.WithDescription("Lotus node height"))

	_ = metric.Must(meter).NewInt64ValueObserver("powergate.lotus.height.diff",
		func(ctx context.Context, result metric.Int64ObserverResult) {
			lsm.lock.Lock()
			defer lsm.lock.Unlock()
			result.Observe(lsm.heightDiff)
		}, metric.WithDescription("Lotus node height syncing diff"))
}
