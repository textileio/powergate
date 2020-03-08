package jsonjobstore

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/fil-tools/ffs"
	"github.com/textileio/fil-tools/ffs/scheduler"
)

var (
	log = logging.Logger("ffs-sched-jobstore")
)

// JobStore is an scheduler.JobStore implementation that saves
// state of scheduler Jobs in a Datastore.
type JobStore struct {
	lock     sync.Mutex
	ds       datastore.Datastore
	watchers []watcher
}

var _ scheduler.JobStore = (*JobStore)(nil)

type watcher struct {
	iid ffs.InstanceID
	ch  chan ffs.Job
}

// New returns a new JobStore backed by the Datastore.
func New(ds datastore.Datastore) *JobStore {
	return &JobStore{
		ds: ds,
	}
}

// GetByStatus returns all Jobs with the specified JobStatus.
func (js *JobStore) GetByStatus(status ffs.JobStatus) ([]ffs.Job, error) {
	js.lock.Lock()
	defer js.lock.Unlock()
	q := query.Query{Prefix: ""}
	res, err := js.ds.Query(q)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var ret []ffs.Job
	for r := range res.Next() {
		job := ffs.Job{}
		if err := json.Unmarshal(r.Value, &job); err != nil {
			return nil, err
		}
		if job.Status == status {
			ret = append(ret, job)
		}
	}
	return ret, nil
}

// Put saves Job's data in the Datastore.
func (js *JobStore) Put(j ffs.Job) error {
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

// Get returns the current state of Job. If doesn't exist, returns
// ErrNotFound.
func (js *JobStore) Get(jid ffs.JobID) (ffs.Job, error) {
	js.lock.Lock()
	defer js.lock.Unlock()

	k := makeJobKey(jid)
	buf, err := js.ds.Get(k)
	job := ffs.Job{}
	if err == datastore.ErrNotFound {
		return job, scheduler.ErrNotFound
	}
	if err != nil {
		return job, fmt.Errorf("getting job from datastore: %s", err)
	}
	if err := json.Unmarshal(buf, &job); err != nil {
		return job, fmt.Errorf("unmarshaling job: %s", err)
	}
	return job, nil
}

// Watch subscribes to Job changes from a specified Powergate instance.
func (js *JobStore) Watch(iid ffs.InstanceID) <-chan ffs.Job {
	js.lock.Lock()
	defer js.lock.Unlock()

	ch := make(chan ffs.Job, 1)
	js.watchers = append(js.watchers, watcher{iid: iid, ch: ch})
	return ch
}

// Unwatch unregisters a channel returned from Watch().
func (js *JobStore) Unwatch(ch <-chan ffs.Job) {
	js.lock.Lock()
	defer js.lock.Unlock()
	for i := range js.watchers {
		if js.watchers[i].ch == ch {
			close(js.watchers[i].ch)
			js.watchers[i] = js.watchers[len(js.watchers)-1]
			js.watchers = js.watchers[:len(js.watchers)-1]
		}
	}
}

// Close closes the JobStore, unregistering any subscribed watchers.
func (js *JobStore) Close() error {
	js.lock.Lock()
	defer js.lock.Unlock()
	for i := range js.watchers {
		close(js.watchers[i].ch)
	}
	return nil
}

func (js *JobStore) notifyWatchers(j ffs.Job) {
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

func makeJobKey(jid ffs.JobID) datastore.Key {
	return datastore.NewKey(jid.String())
}
