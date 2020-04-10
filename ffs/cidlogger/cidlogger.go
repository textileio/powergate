package cidlogger

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
)

var (
	log = logging.Logger("ffs-cidlogger")
)

type CidLogger struct {
	ds datastore.Datastore

	lock     sync.Mutex
	watchers []chan<- ffs.LogEntry
	closed   bool
}

type logEntry struct {
	Cid       cid.Cid
	Timestamp int64
	Jid       ffs.JobID
	Msg       string
}

var _ ffs.CidLogger = (*CidLogger)(nil)

func New(ds datastore.Datastore) *CidLogger {
	return &CidLogger{
		ds: ds,
	}
}

func (cl *CidLogger) Log(ctx context.Context, c cid.Cid, format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
	jid := ffs.EmptyJobID
	if ctxjid, ok := ctx.Value(ffs.CtxKeyJid).(ffs.JobID); ok {
		jid = ctxjid
	}
	now := time.Now().UnixNano()
	key := makeKey(c, now)
	le := logEntry{
		Cid:       c,
		Jid:       jid,
		Msg:       fmt.Sprintf(format, a...),
		Timestamp: now,
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
		Timestamp: time.Unix(0, le.Timestamp),
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

func (cl *CidLogger) Close() error {
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

func makeKey(c cid.Cid, t int64) datastore.Key {
	strt := strconv.FormatInt(t, 10)
	return datastore.NewKey(c.String()).ChildString(strt)
}
