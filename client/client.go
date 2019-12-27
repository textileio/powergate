package client

import (
	"net/http"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
)

// New creates a new client to Lotus API
func New(addr string, authToken string) (api.FullNode, func(), error) {
	headers := http.Header{
		"Authorization": []string{"Bearer " + authToken},
	}
	client, close, err := client.NewFullNodeRPC("ws://"+addr+"/rpc/v0", headers)
	if err != nil {
		return nil, nil, err
	}
	return client, close, nil
}
