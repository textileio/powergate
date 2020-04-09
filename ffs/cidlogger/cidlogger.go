package cidlogger

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
}

type logEntry struct {
	JID ffs.JobID
	Msg string
}

var _ ffs.CidLogger = (*CidLogger)(nil)

func New(ds datastore.Datastore) *CidLogger {
	return &CidLogger{
		ds: ds,
	}
}

func (cl *CidLogger) Log(ctx context.Context, c cid.Cid, format string, a ...interface{}) {
	log.Infof(format, a)
	if err := cl.save(ctx, c, format, a...); err != nil {
		log.Errorf("saving client-log to datastore: %s", err)
	}
}

func (cl *CidLogger) save(ctx context.Context, c cid.Cid, format string, a ...interface{}) error {
	jid := ffs.EmptyJobID
	if ctxjid, ok := ctx.Value(ffs.CtxKeyJid).(ffs.JobID); ok {
		jid = ctxjid
	}
	key := makeKey(c)
	le := logEntry{
		JID: jid,
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
