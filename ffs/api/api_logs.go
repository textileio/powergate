package api

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

// WatchLogs pushes human-friendly messages about Cid executions. The method is blocking
// and will continue to send messages until the context is canceled.
func (i *API) WatchLogs(ctx context.Context, ch chan<- ffs.LogEntry, c cid.Cid, opts ...GetLogsOption) error {
	_, err := i.is.getStorageConfigs(c)
	if err == ErrNotFound {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("validating cid: %s", err)
	}

	config := &GetLogsConfig{}
	for _, o := range opts {
		o(config)
	}

	if config.history {
		lgs, err := i.sched.GetLogsByCid(ctx, i.cfg.ID, c)
		if err != nil {
			return fmt.Errorf("getting history logs of %s: %s", c, err)
		}
		for _, l := range lgs {
			if config.jid == ffs.EmptyJobID || config.jid == l.Jid {
				ch <- l
			}
		}
	}

	ichan := make(chan ffs.LogEntry, cap(ch))
	go func() {
		err = i.sched.WatchLogs(ctx, ichan)
		close(ichan)
	}()
	for le := range ichan {
		if c == le.Cid && le.APIID == i.cfg.ID && (config.jid == ffs.EmptyJobID || config.jid == le.Jid) {
			ch <- le
		}
	}
	if err != nil {
		return fmt.Errorf("listening to cid logs: %s", err)
	}

	return nil
}
