package notifications

import (
	"context"
	"io"
	"net/http"

	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/v2/ffs"
)

var (
	log = logging.Logger("notifier")
)

type Notifier interface {
	RegisterJob(jobId ffs.JobID, configs []*ffs.NotificationConfig)
	NotifyJobUpdates(job JobUpdates)
}

type notifier struct {
	ctx           context.Context
	configs       *configStore
	updates       chan JobUpdates
	toDelete      chan ffs.JobID
	notifications chan *notification
}

func New(ctx context.Context) *notifier {
	nt := &notifier{
		ctx:           ctx,
		configs:       newConfigStore(),
		updates:       make(chan JobUpdates, 1000),
		toDelete:      make(chan ffs.JobID, 1000),
		notifications: make(chan *notification, 1000),
	}

	go nt.run()

	const workers = 10
	for i := 0; i < workers; i++ {
		go nt.worker()
	}

	return nt
}

func (n *notifier) RegisterJob(jobId ffs.JobID, configs []*ffs.NotificationConfig) {
	if configs == nil {
		return
	}

	n.configs.put(jobId, configs)
}

func (n *notifier) NotifyJobUpdates(jobUpdates JobUpdates) {
	n.updates <- jobUpdates
}

func (n *notifier) run() {
	for {
		select {
		case <-n.ctx.Done():
			return

		case updates, ok := <-n.updates:
			if !ok {
				return
			}

			config := n.configs.get(updates.JobID())
			if config == nil {
				continue
			}

			n.notifyAll(config, updates)

			if updates.FinalUpdates() {
				n.configs.delete(updates.JobID())
			}
		}
	}
}

func (n *notifier) notifyAll(configs []*ffs.NotificationConfig, updates JobUpdates) {
	for _, cfg := range configs {
		if cfg == nil {
			continue
		}

		n.notify(cfg, updates)
	}
}

func (n *notifier) notify(config *ffs.NotificationConfig, updates JobUpdates) {
	if matchNotificationConfig(config.Configuration, updates) {
		payload, err := updates.Payload()
		if err != nil {
			log.Errorf("failed to make notification payload: %s", err)
			return
		}

		n.publishNotification(config.Webhook, payload)
	}
}

func matchNotificationConfig(config *ffs.WebhookConfiguration, updates JobUpdates) bool {
	if config == nil {
		return false
	}

	return matchNotificationEvents(config.Events, updates) || matchNotificationAlerts(config.Alerts, updates)
}

func matchNotificationEvents(events []string, updates JobUpdates) bool {
	for _, event := range events {
		if updates.MatchNotificationEvent(event) {
			return true
		}
	}

	return false
}

func matchNotificationAlerts(alerts []*ffs.WebhookAlert, updates JobUpdates) bool {
	for _, alert := range alerts {
		if updates.MatchNotificationAlert(alert) {
			return true
		}
	}

	return false
}

func (n *notifier) publishNotification(webhook *ffs.Webhook, payload io.Reader) {
	if webhook == nil || payload == nil {
		return
	}

	n.notifications <- &notification{
		webhook: webhook,
		payload: payload,
	}
}

type notification struct {
	webhook *ffs.Webhook
	payload io.Reader
}

func (n *notifier) worker() {
	client := http.DefaultClient

	for {
		select {
		case <-n.ctx.Done():
			return

		case notification, ok := <-n.notifications:
			if !ok {
				return
			}

			err := notification.webhook.Publish(client, notification.payload)
			if err != nil {
				log.Errorf("failed to publish notification: %s", err)
			}
		}
	}
}
