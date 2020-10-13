package client

import (
	"context"
	"crypto/tls"
	"strings"

	buildinfoRpc "github.com/textileio/powergate/buildinfo/rpc"
	ffsRpc "github.com/textileio/powergate/ffs/rpc"
	healthRpc "github.com/textileio/powergate/health/rpc"
	askRpc "github.com/textileio/powergate/index/ask/rpc"
	faultsRpc "github.com/textileio/powergate/index/faults/rpc"
	minerRpc "github.com/textileio/powergate/index/miner/rpc"
	netRpc "github.com/textileio/powergate/net/rpc"
	adminProto "github.com/textileio/powergate/proto/admin/v1"
	proto "github.com/textileio/powergate/proto/powergate/v1"
	reputationRpc "github.com/textileio/powergate/reputation/rpc"
	walletRpc "github.com/textileio/powergate/wallet/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client provides the client api.
type Client struct {
	Asks            *Asks
	Miners          *Miners
	Faults          *Faults
	Wallet          *Wallet
	Reputation      *Reputation
	FFS             *FFS
	Health          *Health
	Net             *Net
	Jobs            *Jobs
	Admin           *Admin
	conn            *grpc.ClientConn
	buildInfoClient buildinfoRpc.RPCServiceClient
}

type ctxKey string

// AuthKey is the key that should be used to set the auth token in a Context.
const AuthKey = ctxKey("ffstoken")

// AdminKey is the key that should be used to set the admin auth token in a Context.
const AdminKey = ctxKey("admintoken")

// TokenAuth provides token based auth.
type TokenAuth struct {
	Secure bool
}

// GetRequestMetadata returns request metadata that includes the auth token.
func (t TokenAuth) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	md := map[string]string{}

	token, ok := ctx.Value(AuthKey).(string)
	if ok && token != "" {
		md["X-ffs-Token"] = token
	}

	adminToken, ok := ctx.Value(AdminKey).(string)
	if ok && adminToken != "" {
		md["X-pow-admin-token"] = adminToken
	}

	return md, nil
}

// RequireTransportSecurity specifies if the connection should be secure.
func (t TokenAuth) RequireTransportSecurity() bool {
	return t.Secure
}

// CreateClientConn creates a gRPC connection with sensible defaults and the provided overrides.
func CreateClientConn(target string, optsOverrides ...grpc.DialOption) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if strings.Contains(target, "443") {
		creds = credentials.NewTLS(&tls.Config{})
	}

	auth := TokenAuth{}
	var opts []grpc.DialOption
	if creds != nil {
		opts = append(opts, grpc.WithTransportCredentials(creds))
		auth.Secure = true
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithPerRPCCredentials(auth))
	opts = append(opts, optsOverrides...)

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// NewClient creates a client.
func NewClient(host string, optsOverrides ...grpc.DialOption) (*Client, error) {
	conn, err := CreateClientConn(host, optsOverrides...)
	if err != nil {
		return nil, err
	}
	powClient := proto.NewPowergateServiceClient(conn)
	client := &Client{
		Asks:            &Asks{client: askRpc.NewRPCServiceClient(conn)},
		Miners:          &Miners{client: minerRpc.NewRPCServiceClient(conn)},
		Faults:          &Faults{client: faultsRpc.NewRPCServiceClient(conn)},
		Wallet:          &Wallet{walletClient: walletRpc.NewRPCServiceClient(conn), powergateClient: powClient},
		Reputation:      &Reputation{client: reputationRpc.NewRPCServiceClient(conn)},
		FFS:             &FFS{client: ffsRpc.NewRPCServiceClient(conn)},
		Health:          &Health{client: healthRpc.NewRPCServiceClient(conn)},
		Net:             &Net{client: netRpc.NewRPCServiceClient(conn)},
		Jobs:            &Jobs{client: powClient},
		Admin:           &Admin{client: adminProto.NewPowergateAdminServiceClient(conn)},
		conn:            conn,
		buildInfoClient: buildinfoRpc.NewRPCServiceClient(conn),
	}
	return client, nil
}

// Host returns the client host address.
func (c *Client) Host() string {
	return c.conn.Target()
}

// BuildInfo returns build info about the server.
func (c *Client) BuildInfo(ctx context.Context) (*buildinfoRpc.BuildInfoResponse, error) {
	return c.buildInfoClient.BuildInfo(ctx, &buildinfoRpc.BuildInfoRequest{})
}

// Close closes the client's grpc connection and cancels any active requests.
func (c *Client) Close() error {
	return c.conn.Close()
}
