package notifications

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/textileio/powergate/v2/deals"
)

type Notifier interface {
	NotifyDeal(dr deals.StorageDealRecord)
}

type notifier struct {
	client *http.Client
}

func New() *notifier {
	return &notifier{
		client: http.DefaultClient,
	}
}

func (n *notifier) NotifyDeal(dr deals.StorageDealRecord) {
	endpoint := "https://vmanilo.free.beeceptor.com/webhook"
	contentType := "application/json"
	payload, _ := json.Marshal(dr)

	n.client.Post(endpoint, contentType, bytes.NewBuffer(payload))
}