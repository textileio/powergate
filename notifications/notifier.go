package notifications

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/ffs"
)

type Notifier interface {
	RegisterStorageJob(job ffs.StorageJob, notificationConfig []*ffs.NotificationConfig)
	NotifyStorageJob(job ffs.StorageJob, dealInfo deals.StorageDealInfo)
}

type storageJobUpdate struct {
	job      ffs.StorageJob
	dealInfo deals.StorageDealInfo
}

type notifier struct {
	client                  *http.Client
	store                   *configStore
	storageJobNotifications chan *storageJobUpdate
}

func New() *notifier {
	nt := &notifier{
		client:                  http.DefaultClient,
		store:                   newConfigStore(),
		storageJobNotifications: make(chan *storageJobUpdate, 1000),
	}

	go nt.run()

	return nt
}

func (n *notifier) RegisterStorageJob(job ffs.StorageJob, notificationConfig []*ffs.NotificationConfig) {
	if notificationConfig == nil {
		return
	}

	n.store.put(job, notificationConfig)
}

func (n *notifier) NotifyStorageJob(job ffs.StorageJob, dealInfo deals.StorageDealInfo) {
	n.storageJobNotifications <- &storageJobUpdate{
		job:      job,
		dealInfo: dealInfo,
	}
}

func (n *notifier) run() {
	for updates := range n.storageJobNotifications {
		if updates == nil {
			continue
		}

		config := n.store.get(updates.job.ID)
		if config == nil {
			continue
		}

		n.notifyAll(config.notifications, updates)
	}
}

func (n *notifier) notifyAll(configs []*ffs.NotificationConfig, updates *storageJobUpdate) {
	for _, cfg := range configs {
		n.notify(cfg, updates)
	}
}

func (n *notifier) notify(config *ffs.NotificationConfig, updates *storageJobUpdate) {
	if matchNotificationConfig(config.Configuration, updates.dealInfo) {
		n.publishNotification(config.Webhook, updates)
	}
}

func matchNotificationConfig(config *ffs.WebhookConfiguration, updates deals.StorageDealInfo) bool {
	if config == nil {
		return false
	}

	return matchNotificationEvents(config.Events, updates) || matchNotificationAlerts(config.Alerts, updates)
}

func matchNotificationEvents(events []string, updates deals.StorageDealInfo) bool {
	// TODO
	return true
}

func matchNotificationAlerts(alerts []*ffs.WebhookAlert, updates deals.StorageDealInfo) bool {
	// TODO
	return false
}

func (n *notifier) publishNotification(webhook *ffs.Webhook, updates *storageJobUpdate) {
	if webhook == nil {
		return
	}

	data, err := json.Marshal(updates.dealInfo)
	if err != nil {
		// TODO: log error
		return
	}

	err = webhook.Publish(n.client, bytes.NewBuffer(data))
	if err != nil {
		// TODO: log error
		return
	}
}

type configStore struct {
	sync.RWMutex

	configs map[ffs.JobID]*jobConfig
}

func newConfigStore() *configStore {
	return &configStore{
		configs: make(map[ffs.JobID]*jobConfig),
	}
}

type jobConfig struct {
	job           ffs.StorageJob
	notifications []*ffs.NotificationConfig
}

func (s *configStore) put(job ffs.StorageJob, notifications []*ffs.NotificationConfig) {
	s.Lock()
	defer s.Unlock()

	s.configs[job.ID] = &jobConfig{
		job:           job,
		notifications: notifications,
	}
}

func (s *configStore) get(jobID ffs.JobID) *jobConfig {
	s.RLock()
	defer s.RUnlock()

	return s.configs[jobID]
}
