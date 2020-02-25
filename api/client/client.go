package client

import (
	dealsPb "github.com/textileio/fil-tools/deals/pb"
	fpaPb "github.com/textileio/fil-tools/fpa/pb"
	asksPb "github.com/textileio/fil-tools/index/ask/pb"
	minerPb "github.com/textileio/fil-tools/index/miner/pb"
	slashingPb "github.com/textileio/fil-tools/index/slashing/pb"
	reputationPb "github.com/textileio/fil-tools/reputation/pb"
	walletPb "github.com/textileio/fil-tools/wallet/pb"
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
	Fpa        *fpa
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
		Fpa:        &fpa{client: fpaPb.NewAPIClient(conn)},
		conn:       conn,
	}
	return client, nil
}

// Close closes the client's grpc connection and cancels any active requests
func (c *Client) Close() error {
	return c.conn.Close()
}
