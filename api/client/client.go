package client

import (
	"context"

	dealsPb "github.com/textileio/powergate/deals/pb"
	ffsRpc "github.com/textileio/powergate/ffs/rpc"
	healthRpc "github.com/textileio/powergate/health/rpc"
	asksPb "github.com/textileio/powergate/index/ask/pb"
	minerPb "github.com/textileio/powergate/index/miner/pb"
	slashingPb "github.com/textileio/powergate/index/slashing/pb"
	netRpc "github.com/textileio/powergate/net/rpc"
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
	Health     *health
	Net        *net
	conn       *grpc.ClientConn
}

type AuthKey string

type TokenAuth struct {
	secure bool
}

func (t TokenAuth) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	md := map[string]string{}
	token, ok := ctx.Value(AuthKey("ffstoken")).(string)
	if ok && token != "" {
		md["X-ffs-Token"] = token
	}
	return md, nil
}

func (t TokenAuth) RequireTransportSecurity() bool {
	return t.secure
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
		Ffs:        &ffs{client: ffsRpc.NewFFSClient(conn)},
		Health:     &health{client: healthRpc.NewHealthClient(conn)},
		Net:        &net{client: netRpc.NewNetClient(conn)},
		conn:       conn,
	}
	return client, nil
}

// Close closes the client's grpc connection and cancels any active requests
func (c *Client) Close() error {
	return c.conn.Close()
}
