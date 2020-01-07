package client

import (
	ma "github.com/multiformats/go-multiaddr"
	dealsPb "github.com/textileio/filecoin/deals/pb"
	"github.com/textileio/filecoin/util"
	walletPb "github.com/textileio/filecoin/wallet/pb"
	"google.golang.org/grpc"
)

// Client provides the client api
type Client struct {
	Deals  *Deals
	Wallet *Wallet
	conn   *grpc.ClientConn
}

// NewClient starts the client
func NewClient(maddr ma.Multiaddr) (*Client, error) {
	addr, err := util.TCPAddrFromMultiAddr(maddr)
	if err != nil {
		return nil, err
	}
	// ToDo: Support secure connection
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := &Client{
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
