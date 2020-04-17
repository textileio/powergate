package client

import (
	dealsPb "github.com/textileio/powergate/deals/pb"
	ffsRpc "github.com/textileio/powergate/ffs/rpc"
	asksPb "github.com/textileio/powergate/index/ask/pb"
	minerPb "github.com/textileio/powergate/index/miner/pb"
	slashingPb "github.com/textileio/powergate/index/slashing/pb"
	reputationPb "github.com/textileio/powergate/reputation/pb"
	walletPb "github.com/textileio/powergate/wallet/pb"
	"google.golang.org/grpc"
)

// Client provides the client api
type Client struct {
	Asks       *Asks
	Miners     *Miners
	Slashing   *Slashing
	Deals      *Deals
	Wallet     *Wallet
	Reputation *Reputation
	Ffs        *ffs
	conn       *grpc.ClientConn
}

// NewClient starts the client
func NewClient(target string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Asks:       &Asks{client: asksPb.NewAPIClient(conn)},
		Miners:     &Miners{client: minerPb.NewAPIClient(conn)},
		Slashing:   &Slashing{client: slashingPb.NewAPIClient(conn)},
		Deals:      &Deals{client: dealsPb.NewAPIClient(conn)},
		Wallet:     &Wallet{client: walletPb.NewAPIClient(conn)},
		Reputation: &Reputation{client: reputationPb.NewAPIClient(conn)},
		Ffs:        &ffs{client: ffsRpc.NewFFSAPIClient(conn)},
		conn:       conn,
	}
	return client, nil
}

// Close closes the client's grpc connection and cancels any active requests
func (c *Client) Close() error {
	return c.conn.Close()
}
