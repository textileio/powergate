package client

import (
	"context"

	"github.com/multiformats/go-multiaddr"
	dealsRpc "github.com/textileio/powergate/deals/rpc"
	ffsRpc "github.com/textileio/powergate/ffs/rpc"
	healthRpc "github.com/textileio/powergate/health/rpc"
	asksPb "github.com/textileio/powergate/index/ask/pb"
	minerPb "github.com/textileio/powergate/index/miner/pb"
	slashingPb "github.com/textileio/powergate/index/slashing/pb"
	netRpc "github.com/textileio/powergate/net/rpc"
	reputationRpc "github.com/textileio/powergate/reputation/rpc"
	"github.com/textileio/powergate/util"
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
	FFS        *FFS
	Health     *Health
	Net        *Net
	conn       *grpc.ClientConn
}

type ctxKey string

// AuthKey is the key that should be used to set the auth token in a Context
const AuthKey = ctxKey("ffstoken")

// TokenAuth provides token based auth
type TokenAuth struct {
	secure bool
}

// GetRequestMetadata returns request metadata that includes the auth token
func (t TokenAuth) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	md := map[string]string{}
	token, ok := ctx.Value(ctxKey(AuthKey)).(string)
	if ok && token != "" {
		md["X-ffs-Token"] = token
	}
	return md, nil
}

// RequireTransportSecurity specifies if the connection should be secure
func (t TokenAuth) RequireTransportSecurity() bool {
	return t.secure
}

// NewClient creates a client
func NewClient(ma multiaddr.Multiaddr, opts ...grpc.DialOption) (*Client, error) {
	addr, err := util.TCPAddrFromMultiAddr(ma)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Asks:       &Asks{client: asksPb.NewAPIClient(conn)},
		Miners:     &Miners{client: minerPb.NewAPIClient(conn)},
		Slashing:   &Slashing{client: slashingPb.NewAPIClient(conn)},
		Deals:      &Deals{client: dealsRpc.NewAPIClient(conn)},
		Wallet:     &Wallet{client: walletPb.NewAPIClient(conn)},
		Reputation: &Reputation{client: reputationRpc.NewAPIClient(conn)},
		FFS:        &FFS{client: ffsRpc.NewAPIClient(conn)},
		Health:     &Health{client: healthRpc.NewAPIClient(conn)},
		Net:        &Net{client: netRpc.NewAPIClient(conn)},
		conn:       conn,
	}
	return client, nil
}

// Close closes the client's grpc connection and cancels any active requests
func (c *Client) Close() error {
	return c.conn.Close()
}
