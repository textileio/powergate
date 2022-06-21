package notifications

import (
	"sync"

	"github.com/textileio/powergate/v2/ffs"
)

type configStore struct {
	sync.RWMutex

	configs map[ffs.JobID][]*ffs.NotificationConfig
}

func newConfigStore() *configStore {
	return &configStore{
		configs: make(map[ffs.JobID][]*ffs.NotificationConfig),
	}
}

func (s *configStore) put(jobId ffs.JobID, configs []*ffs.NotificationConfig) {
	s.Lock()
	defer s.Unlock()

	s.configs[jobId] = configs
}

func (s *configStore) get(jobId ffs.JobID) []*ffs.NotificationConfig {
	s.RLock()
	defer s.RUnlock()

	return s.configs[jobId]
}

func (s *configStore) delete(jobId ffs.JobID) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.configs[jobId]; ok {
		delete(s.configs, jobId)
	}
}
