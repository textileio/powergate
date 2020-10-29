package client

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/textileio/powergate/api/client/admin"
	adminProto "github.com/textileio/powergate/proto/admin/v1"
	proto "github.com/textileio/powergate/proto/powergate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client provides the client api.
type Client struct {
	StorageConfig *StorageConfig
	Data          *Data
	Wallet        *Wallet
	Deals         *Deals
	StorageJobs   *StorageJobs
	Admin         *admin.Admin
	conn          *grpc.ClientConn
	client        proto.PowergateServiceClient
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
	client := proto.NewPowergateServiceClient(conn)
	return &Client{
		StorageConfig: &StorageConfig{client: client},
		Data:          &Data{client: client},
		Wallet:        &Wallet{client: client},
		Deals:         &Deals{client: client},
		StorageJobs:   &StorageJobs{client: client},
		Admin:         admin.NewAdmin(adminProto.NewPowergateAdminServiceClient(conn)),
		conn:          conn,
		client:        client,
	}, nil
}

// Host returns the client host address.
func (c *Client) Host() string {
	return c.conn.Target()
}

// BuildInfo returns build info about the server.
func (c *Client) BuildInfo(ctx context.Context) (*proto.BuildInfoResponse, error) {
	return c.client.BuildInfo(ctx, &proto.BuildInfoRequest{})
}

// StorageProfileID returns the storage profile StorageProfileID.
func (c *Client) StorageProfileID(ctx context.Context) (*proto.StorageProfileIdentifierResponse, error) {
	return c.client.StorageProfileIdentifier(ctx, &proto.StorageProfileIdentifierRequest{})
}

// Close closes the client's grpc connection and cancels any active requests.
func (c *Client) Close() error {
	return c.conn.Close()
}
