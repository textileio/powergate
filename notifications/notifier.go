package notifications

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
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
	if config == nil {
		return
	}

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

const (
	all            = "*"
	created        = "created"
	completed      = "completed"
	retried        = "retried"
	failed         = "failed"
	expired        = "expired"
	slashed        = "slashed"
	separator      = "-"
	storageDeal    = "storage-deal"
	storageAuction = "storage-auction"
	dataRetrieval  = "data-retrieval"

	AllEvents          = all
	AllCreatedEvents   = all + separator + created
	AllCompletedEvents = all + separator + completed
	AllRetriedEvents   = all + separator + retried
	AllFailedEvents    = all + separator + failed

	AllStorageDealEvents      = storageDeal + separator + all
	StorageDealCreatedEvent   = storageDeal + separator + created
	StorageDealCompletedEvent = storageDeal + separator + completed
	StorageDealRetriedEvent   = storageDeal + separator + retried
	StorageDealFailedEvent    = storageDeal + separator + failed
	StorageDealExpiredEvent   = storageDeal + separator + expired
	StorageDealSlashedEvent   = storageDeal + separator + slashed

	AllStorageAuctionEvents      = storageAuction + separator + all
	StorageAuctionCreatedEvent   = storageAuction + separator + created
	StorageAuctionCompletedEvent = storageAuction + separator + completed
	StorageAuctionFailedEvent    = storageAuction + separator + failed

	AllDataRetrievalEvents      = dataRetrieval + separator + all
	DataRetrievalCompletedEvent = dataRetrieval + separator + completed
	DataRetrievalRetriedEvent   = dataRetrieval + separator + retried
	DataRetrievalFailedEvent    = dataRetrieval + separator + failed
)

func matchNotificationEvents(events []string, updates deals.StorageDealInfo) bool {
	for _, event := range events {
		if matchNotificationEvent(event, updates) {
			return true
		}
	}

	return false
}

func matchNotificationEvent(event string, updates deals.StorageDealInfo) bool {
	switch event {
	case AllEvents:
		return true
	case AllCreatedEvents:
		// TODO: add created events
		return updates.DealID != 0 // ||
	case AllCompletedEvents:
		// TODO: add other completed events
		return updates.StateID == storagemarket.StorageDealActive // ||

	// TODO:
	// case AllRetriedEvents:

	case AllStorageDealEvents:
		return true

	case StorageDealCreatedEvent:
		return updates.DealID != 0

	case StorageDealCompletedEvent:
		return updates.StateID == storagemarket.StorageDealActive

	// TODO:
	// case StorageDealRetriedEvent:

	case StorageDealFailedEvent:
		return updates.StateID == storagemarket.StorageDealFailing || updates.StateID == storagemarket.StorageDealError || updates.Message != ""

	case StorageDealExpiredEvent:
		return updates.StateID == storagemarket.StorageDealExpired

	case StorageDealSlashedEvent:
		return updates.StateID == storagemarket.StorageDealSlashed

	default:
		return false
	}
}

func matchNotificationAlerts(alerts []*ffs.WebhookAlert, updates deals.StorageDealInfo) bool {
	// TODO
	return false
}

type notification struct {
	Cid         string    `json:"cid"`
	JobID       ffs.JobID `json:"jobId"`
	JobStatus   string    `json:"jobStatus"`
	Miner       string    `json:"miner"`
	Price       uint64    `json:"price"`
	ProposalCid string    `json:"proposalCid"`
	DealID      uint64    `json:"dealId,omitempty"`
	DealStatus  string    `json:"dealStatus"`
	ErrCause    string    `json:"error,omitempty"`
	Message     string    `json:"message,omitempty"`
}

func (n *notifier) publishNotification(webhook *ffs.Webhook, updates *storageJobUpdate) {
	if webhook == nil {
		return
	}

	obj := &notification{
		Cid:         updates.job.Cid.String(),
		JobID:       updates.job.ID,
		JobStatus:   ffs.JobStatusStr[updates.job.Status],
		Miner:       updates.dealInfo.Miner,
		Price:       updates.dealInfo.PricePerEpoch,
		ProposalCid: updates.dealInfo.ProposalCid.String(),
		DealID:      updates.dealInfo.DealID,
		DealStatus:  updates.dealInfo.StateName,
		ErrCause:    updates.job.ErrCause,
		Message:     updates.dealInfo.Message,
	}

	data, err := json.Marshal(obj)
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
