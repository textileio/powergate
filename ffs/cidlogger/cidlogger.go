package cidlogger

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
)

var (
	log = logging.Logger("ffs-cidlogger")
)

// CidLogger is a datastore backed implementation of ffs.CidLogger.
type CidLogger struct {
	ds datastore.Datastore

	lock     sync.Mutex
	watchers []chan<- ffs.LogEntry
	closed   bool
}

type logEntry struct {
	Cid         cid.Cid
	RetrievalID ffs.RetrievalID
	Timestamp   int64
	Jid         ffs.JobID
	Msg         string
}

var _ ffs.CidLogger = (*CidLogger)(nil)

// New returns a new CidLogger.
func New(ds datastore.Datastore) *CidLogger {
	return &CidLogger{
		ds: ds,
	}
}

// Log logs a log entry for a Cid. The ctx can contain an optional ffs.CtxKeyJid to add
// additional metadata about the log entry being part of a Job execution.
func (cl *CidLogger) Log(ctx context.Context, c cid.Cid, format string, a ...interface{}) {
	cl.log(ctx, c, ffs.EmptyRetrievalID, format, a...)
}

func (cl *CidLogger) LogRetrieval(ctx context.Context, rid ffs.RetrievalID, format string, a ...interface{}) {
	cl.log(ctx, cid.Undef, rid, format, a...)
}

func (cl *CidLogger) log(ctx context.Context, c cid.Cid, rid ffs.RetrievalID, format string, a ...interface{}) {
	log.Infof(format, a...)
	jid := ffs.EmptyJobID
	if ctxjid, ok := ctx.Value(ffs.CtxKeyJid).(ffs.JobID); ok {
		jid = ctxjid
	}
	now := time.Now()
	nowNano := now.UnixNano()
	key := makeKey(c, rid, nowNano)
	le := logEntry{
		Cid:         c,
		RetrievalID: rid,
		Jid:         jid,
		Msg:         fmt.Sprintf(format, a...),
		Timestamp:   nowNano,
	}
	b, err := json.Marshal(le)
	if err != nil {
		log.Errorf("marshaling to json: %s", err)
		return
	}
	if err := cl.ds.Put(key, b); err != nil {
		log.Error("saving to datastore: %s", err)
		return
	}

	entry := ffs.LogEntry{
		Cid:       le.Cid,
		Jid:       le.Jid,
		Timestamp: now,
		Msg:       fmt.Sprintf(format, a...),
	}
	cl.lock.Lock()
	defer cl.lock.Unlock()
	for _, c := range cl.watchers {
		select {
		case c <- entry:
		default:
			log.Warn("slow cid log receiver")
		}
	}
}

// Get returns history logs of a Cid.
// ToDo: rename
func (cl *CidLogger) Get(ctx context.Context, c cid.Cid) ([]ffs.LogEntry, error) {
	q := query.Query{Prefix: makeCidKey(c).String()}
	res, err := cl.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("running query: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing query result: %s", err)
		}
	}()
	var lgs []ffs.LogEntry
	for r := range res.Next() {
		var le logEntry
		if err := json.Unmarshal(r.Value, &le); err != nil {
			return nil, fmt.Errorf("unmarshaling log entry: %s", err)
		}
		lgs = append(lgs, ffs.LogEntry{
			Cid:       le.Cid,
			Jid:       le.Jid,
			Msg:       le.Msg,
			Timestamp: time.Unix(0, le.Timestamp),
		})
	}
	sort.Slice(lgs, func(a, b int) bool {
		return lgs[a].Timestamp.Before(lgs[b].Timestamp)
	})
	return lgs, nil
}

// Watch is a blocking function that writes to the channel all new created log entries.
// The client should cancel the ctx to signal stopping writing to the channel and free resources.
func (cl *CidLogger) Watch(ctx context.Context, c chan<- ffs.LogEntry) error {
	cl.lock.Lock()
	ic := make(chan ffs.LogEntry)
	cl.watchers = append(cl.watchers, ic)
	cl.lock.Unlock()

	stop := false
	for !stop {
		select {
		case <-ctx.Done():
			stop = true
		case l, ok := <-ic:
			if !ok {
				return fmt.Errorf("cidlogger was closed with a listening client")
			}
			c <- l
		}
	}
	cl.lock.Lock()
	defer cl.lock.Unlock()
	for i := range cl.watchers {
		if cl.watchers[i] == ic {
			cl.watchers = append(cl.watchers[:i], cl.watchers[i+1:]...)
			break
		}
	}
	return nil
}

// Close closes and cancels all watchers that might be active.
func (cl *CidLogger) Close() error {
	log.Info("closing...")
	defer log.Info("closed")
	cl.lock.Lock()
	defer cl.lock.Unlock()

	if cl.closed {
		return nil
	}
	cl.closed = true
	for _, w := range cl.watchers {
		close(w)
	}
	return nil
}

func makeKey(c cid.Cid, rid ffs.RetrievalID, timestamp int64) datastore.Key {
	strt := strconv.FormatInt(timestamp, 10)
	if c != cid.Undef {
		return makeCidKey(c).ChildString(strt)
	}
	if rid != ffs.EmptyRetrievalID {
		return makeRetrievalKey(c).ChildString(strt)
	}
	panic("log should be from stored cid or retrieval request")
}

func makeCidKey(c cid.Cid) datastore.Key {
	return datastore.NewKey(util.CidToString(c))
}

func makeRetrievalKey(rid cid.Cid) datastore.Key {
	return datastore.NewKey(rid.String())
}
