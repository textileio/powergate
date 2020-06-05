package stub

import (
	"google.golang.org/grpc"
)

// Client provides the client api.
type Client struct {
	Asks       *Asks
	Miners     *Miners
	Faults     *Faults
	Deals      *Deals
	Wallet     *Wallet
	Reputation *Reputation
}

// NewClient starts the client.
func NewClient(target string, opts ...grpc.DialOption) (*Client, error) {
	client := &Client{
		Asks:       &Asks{},
		Miners:     &Miners{},
		Faults:     &Faults{},
		Deals:      &Deals{},
		Wallet:     &Wallet{},
		Reputation: &Reputation{},
	}
	return client, nil
}

// Close closes the client's grpc connection and cancels any active requests.
func (c *Client) Close() error {
	return nil
}
