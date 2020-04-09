package logger

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler"
)

var (
	log = logging.Logger("ffs-sch-logger")
)

type Logger struct {
	ds  datastore.Datastore
	pcs scheduler.PushConfigStore
}

var _ scheduler.Logger = (*Logger)(nil)

type logEntry struct {
	JID ffs.JobID
	Msg string
}

func New(ds datastore.Datastore, pcs scheduler.PushConfigStore) *Logger {
	return &Logger{
		ds:  ds,
		pcs: pcs,
	}
}

func (l *Logger) Logger(c cid.Cid, jid ffs.JobID) scheduler.CidLogger {
	return newCidLogger(l.ds, c, jid)
}

type CidLogger struct {
	ds  datastore.Datastore
	c   cid.Cid
	jid ffs.JobID
}

var _ scheduler.CidLogger = (*CidLogger)(nil)

func newCidLogger(ds datastore.Datastore, c cid.Cid, jid ffs.JobID) *CidLogger {
	return &CidLogger{
		ds:  ds,
		c:   c,
		jid: jid,
	}
}

func (cl *CidLogger) LogMsg(format string, a ...interface{}) {
	log.Infof(format, a)
	if err := cl.save(format, a...); err != nil {
		log.Errorf("saving client-log to datastore: %s", err)
	}
}

func (cl *CidLogger) save(format string, a ...interface{}) error {
	key := makeKey(cl.c)
	le := logEntry{
		JID: cl.jid,
		Msg: fmt.Sprintf(format, a...),
	}
	b, err := json.Marshal(le)
	if err != nil {
		return fmt.Errorf("marshaling to json: %s", err)
	}

	return cl.ds.Put(key, b)

}

func makeKey(c cid.Cid) datastore.Key {
	t := time.Now().UnixNano()
	strt := strconv.FormatInt(t, 10)
	return datastore.NewKey(c.String()).ChildString(strt)
}
