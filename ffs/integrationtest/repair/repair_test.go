package repair

import (
	"context"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	it "github.com/textileio/powergate/ffs/integrationtest"
	itmanager "github.com/textileio/powergate/ffs/integrationtest/manager"
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 500
	os.Exit(m.Run())
}

// This isn't very nice way to test for repair. The main problem is that now
// deal start is buffered for future start for 10000 blocks at the Lotus level.
// Se we can't wait that much on a devnet. That setup has some ToDo comments so
// most prob will change and we can do some nicier test here.
// Better than no test is some test, so this tests that the repair logic gets triggered
// and the related Job ran successfully.
func TestRepair(t *testing.T) {
	tests.RunFlaky(t, func(t *tests.FlakyT) {
		scheduler.RepairEvalFrequency = time.Second * 30
		ipfs, _, fapi, cls := itmanager.NewAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, _ := it.AddRandomFile(t, r, ipfs)
		config := fapi.DefaultStorageConfig().WithRepairable(true)
		jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, &config)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ch := make(chan ffs.LogEntry)
		go func() {
			err = fapi.WatchLogs(ctx, ch, cid, api.WithHistory(true))
			close(ch)
		}()
		stop := false
		for !stop {
			select {
			case le, ok := <-ch:
				if !ok {
					require.NoError(t, err)
					stop = true
					continue
				}
				// Expected message: "Job %s was queued for repair evaluation."
				if strings.Contains(le.Msg, "was queued for repair evaluation.") {
					parts := strings.SplitN(le.Msg, " ", 3)
					require.Equal(t, 3, len(parts), "Log message is malformed")
					jid := ffs.JobID(parts[1])
					it.RequireEventualJobState(t, fapi, jid, ffs.Success)
					it.RequireStorageConfig(t, fapi, cid, &config)
					cancel()
				}
			case <-time.After(time.Second * 10):
				t.Errorf("no cid logs related with repairing were received")
				t.FailNow()
			}
		}
	})
}
