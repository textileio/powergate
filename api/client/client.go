package client

import (
	"context"
	"crypto/tls"
	"strings"

	ffsRpc "github.com/textileio/powergate/ffs/rpc"
	healthRpc "github.com/textileio/powergate/health/rpc"
	askRpc "github.com/textileio/powergate/index/ask/rpc"
	faultsRpc "github.com/textileio/powergate/index/faults/rpc"
	minerRpc "github.com/textileio/powergate/index/miner/rpc"
	netRpc "github.com/textileio/powergate/net/rpc"
	reputationRpc "github.com/textileio/powergate/reputation/rpc"
	walletRpc "github.com/textileio/powergate/wallet/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client provides the client api.
type Client struct {
	Asks       *Asks
	Miners     *Miners
	Faults     *Faults
	Wallet     *Wallet
	Reputation *Reputation
	FFS        *FFS
	Health     *Health
	Net        *Net
	conn       *grpc.ClientConn
}

type ctxKey string

// AuthKey is the key that should be used to set the auth token in a Context.
const AuthKey = ctxKey("ffstoken")

// TokenAuth provides token based auth.
type TokenAuth struct {
	secure bool
}

// GetRequestMetadata returns request metadata that includes the auth token.
func (t TokenAuth) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	md := map[string]string{}
	token, ok := ctx.Value(AuthKey).(string)
	if ok && token != "" {
		md["X-ffs-Token"] = token
	}
	return md, nil
}

// RequireTransportSecurity specifies if the connection should be secure.
func (t TokenAuth) RequireTransportSecurity() bool {
	return t.secure
}

// NewClient creates a client.
func NewClient(target string, optsOverrides ...grpc.DialOption) (*Client, error) {
	var creds credentials.TransportCredentials
	if strings.Contains(target, "443") {
		creds = credentials.NewTLS(&tls.Config{})
	}

	auth := TokenAuth{}
	var opts []grpc.DialOption
	if creds != nil {
		opts = append(opts, grpc.WithTransportCredentials(creds))
		auth.secure = true
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithPerRPCCredentials(auth))
	opts = append(opts, optsOverrides...)

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Asks:       &Asks{client: askRpc.NewRPCServiceClient(conn)},
		Miners:     &Miners{client: minerRpc.NewRPCServiceClient(conn)},
		Faults:     &Faults{client: faultsRpc.NewRPCServiceClient(conn)},
		Wallet:     &Wallet{client: walletRpc.NewRPCServiceClient(conn)},
		Reputation: &Reputation{client: reputationRpc.NewRPCServiceClient(conn)},
		FFS:        &FFS{client: ffsRpc.NewRPCServiceClient(conn)},
		Health:     &Health{client: healthRpc.NewRPCServiceClient(conn)},
		Net:        &Net{client: netRpc.NewRPCServiceClient(conn)},
		conn:       conn,
	}
	return client, nil
}

// Close closes the client's grpc connection and cancels any active requests.
func (c *Client) Close() error {
	return c.conn.Close()
}
