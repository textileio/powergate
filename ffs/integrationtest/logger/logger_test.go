package logger

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	it "github.com/textileio/powergate/ffs/integrationtest"
	itmanager "github.com/textileio/powergate/ffs/integrationtest/manager"
	"github.com/textileio/powergate/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 500
	logging.SetAllLoggers(logging.LevelError)
	_ = logging.SetLogLevel("rpc", "FATAL")
	os.Exit(m.Run())
}

func TestLogHistory(t *testing.T) {
	t.Parallel()
	ipfs, _, fapi, cls := itmanager.NewAPI(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	c, _ := it.AddRandomFile(t, r, ipfs)
	jid, err := fapi.PushStorageConfig(c)
	require.NoError(t, err)
	job := it.RequireEventualJobState(t, fapi, jid, ffs.Success)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	ch := make(chan ffs.LogEntry, 100)
	go func() {
		err = fapi.WatchLogs(ctx, ch, c, api.WithHistory(true))
		close(ch)
	}()
	var lgs []ffs.LogEntry
	for le := range ch {
		require.Equal(t, c, le.Cid)
		require.Equal(t, job.ID, le.Jid)
		require.NotEmpty(t, le.Msg)
		if len(lgs) > 0 {
			require.True(t, le.Timestamp.After(lgs[len(lgs)-1].Timestamp))
		}

		lgs = append(lgs, le)
	}
	require.NoError(t, err)
	require.Greater(t, len(lgs), 3) // Ask to have more than 3 log messages.
}

func TestCidLogger(t *testing.T) {
	t.Parallel()
	t.Run("WithNoFilters", func(t *testing.T) {
		ipfs, _, fapi, cls := itmanager.NewAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, _ := it.AddRandomFile(t, r, ipfs)
		jid, err := fapi.PushStorageConfig(cid)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ch := make(chan ffs.LogEntry)
		go func() {
			err = fapi.WatchLogs(ctx, ch, cid)
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
				cancel()
				require.Equal(t, cid, le.Cid)
				require.Equal(t, jid, le.Jid)
				require.True(t, time.Since(le.Timestamp) < time.Second*5)
				require.NotEmpty(t, le.Msg)
			case <-time.After(time.Second):
				t.Fatal("no cid logs were received")
			}
		}

		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, nil)
	})
	t.Run("WithJidFilter", func(t *testing.T) {
		t.Run("CorrectJid", func(t *testing.T) {
			ipfs, _, fapi, cls := itmanager.NewAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := it.AddRandomFile(t, r, ipfs)
			jid, err := fapi.PushStorageConfig(cid)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ch := make(chan ffs.LogEntry)
			go func() {
				err = fapi.WatchLogs(ctx, ch, cid, api.WithJidFilter(jid))
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
					cancel()
					require.Equal(t, cid, le.Cid)
					require.Equal(t, jid, le.Jid)
					require.True(t, time.Since(le.Timestamp) < time.Second*5)
					require.NotEmpty(t, le.Msg)
				case <-time.After(time.Second):
					t.Fatal("no cid logs were received")
				}
			}

			it.RequireEventualJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, nil)
		})
		t.Run("IncorrectJid", func(t *testing.T) {
			ipfs, _, fapi, cls := itmanager.NewAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := it.AddRandomFile(t, r, ipfs)
			jid, err := fapi.PushStorageConfig(cid)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ch := make(chan ffs.LogEntry)
			go func() {
				fakeJid := ffs.NewJobID()
				err = fapi.WatchLogs(ctx, ch, cid, api.WithJidFilter(fakeJid))
				close(ch)
			}()
			select {
			case <-ch:
				t.Fatal("the channels shouldn't receive any log messages")
			case <-time.After(3 * time.Second):
			}
			require.NoError(t, err)

			it.RequireEventualJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, nil)
		})
	})
}
