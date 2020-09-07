package filchain

import (
	"context"
	"fmt"

	"github.com/textileio/powergate/lotus"
)

// FilChain is an abstraction of the Filecoin network.
type FilChain struct {
	clientBuilder lotus.ClientBuilder
}

// New creates a new deal module.
func New(clientBuilder lotus.ClientBuilder) *FilChain {
	return &FilChain{
		clientBuilder: clientBuilder,
	}
}

// GetHeight returns the current height of the chain for the targeted Lotus node.
func (lc *FilChain) GetHeight(ctx context.Context) (uint64, error) {
	client, cls, err := lc.clientBuilder()
	if err != nil {
		return 0, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	h, err := client.ChainHead(ctx)
	if err != nil {
		return 0, fmt.Errorf("get head from lotus node: %s", err)
	}
	return uint64(h.Height()), nil
}
