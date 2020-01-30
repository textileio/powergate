package client

import (
	dealsPb "github.com/textileio/filecoin/deals/pb"
	asksPb "github.com/textileio/filecoin/index/ask/pb"
	minerPb "github.com/textileio/filecoin/index/miner/pb"
	walletPb "github.com/textileio/filecoin/wallet/pb"
	"google.golang.org/grpc"
)

// Client provides the client api
type Client struct {
	Asks   *Asks
	Miners *Miners
	Deals  *Deals
	Wallet *Wallet
	conn   *grpc.ClientConn
}

// NewClient starts the client
func NewClient(target string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Asks:   &Asks{client: asksPb.NewAPIClient(conn)},
		Miners: &Miners{client: minerPb.NewAPIClient(conn)},
		Deals:  &Deals{client: dealsPb.NewAPIClient(conn)},
		Wallet: &Wallet{client: walletPb.NewAPIClient(conn)},
		conn:   conn,
	}
	return client, nil
}

// Close closes the client's grpc connection and cancels any active requests
func (c *Client) Close() error {
	return c.conn.Close()
}
