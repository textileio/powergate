package jsonjobstore

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/fil-tools/fpa"
	"github.com/textileio/fil-tools/fpa/scheduler"
)

var (
	log = logging.Logger("jobstore")
)

type JobStore struct {
	lock     sync.Mutex
	ds       datastore.Datastore
	watchers []watcher
}

type watcher struct {
	iid fpa.InstanceID
	ch  chan fpa.Job
}

var _ scheduler.JobStore = (*JobStore)(nil)

func New(ds datastore.Datastore) *JobStore {
	return &JobStore{
		ds: ds,
	}
}

func (js *JobStore) GetByStatus(status fpa.JobStatus) ([]fpa.Job, error) {
	js.lock.Lock()
	defer js.lock.Unlock()
	q := query.Query{Prefix: ""}
	res, err := js.ds.Query(q)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var ret []fpa.Job
	for r := range res.Next() {
		job := fpa.Job{}
		if err := json.Unmarshal(r.Value, &job); err != nil {
			return nil, err
		}
		if job.Status == status {
			ret = append(ret, job)
		}
	}
	return ret, nil
}

func (js *JobStore) Put(j fpa.Job) error {
	js.lock.Lock()
	defer js.lock.Unlock()
	log.Infof("saving new job %s state", j.ID)
	k := makeJobKey(j.ID)

	buf, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("json marshal job: %s", err)
	}
	if err := js.ds.Put(k, buf); err != nil {
		return fmt.Errorf("saving to store: %s", err)
	}
	js.notifyWatchers(j)
	return nil
}

func (js *JobStore) Watch(iid fpa.InstanceID) <-chan fpa.Job {
	ch := make(chan fpa.Job, 1)
	js.watchers = append(js.watchers, watcher{iid: iid, ch: ch})
	return ch
}

func (js *JobStore) Unwatch(ch <-chan fpa.Job) {
	for i := range js.watchers {
		if js.watchers[i].ch == ch {
			close(js.watchers[i].ch)
			js.watchers[i] = js.watchers[len(js.watchers)-1]
			js.watchers = js.watchers[:len(js.watchers)-1]
		}
	}
}

func (js *JobStore) notifyWatchers(j fpa.Job) {
	for _, w := range js.watchers {
		if w.iid != j.Config.InstanceID {
			continue
		}
		select {
		case w.ch <- j:
			log.Infof("notifying watcher")
		default:
			log.Warnf("slow watcher skipped job %s", j.ID)
		}
	}
}

func makeJobKey(jid fpa.JobID) datastore.Key {
	return datastore.NewKey(jid.String())
}
