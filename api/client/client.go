package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	files "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	"github.com/ipfs/interface-go-ipfs-core/options"
	ipfspath "github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/multiformats/go-multiaddr"
	adminProto "github.com/textileio/powergate/proto/admin/v1"
	proto "github.com/textileio/powergate/proto/powergate/v1"
	"github.com/textileio/powergate/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// Client provides the client api.
type Client struct {
	Wallet      *Wallet
	Deals       *Deals
	StorageJobs *StorageJobs
	Admin       *Admin
	conn        *grpc.ClientConn
	powClient   proto.PowergateServiceClient
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
		Wallet:      &Wallet{client: powClient},
		Deals:       &Deals{client: powClient},
		StorageJobs: &StorageJobs{client: powClient},
		Admin:       &Admin{client: adminProto.NewPowergateAdminServiceClient(conn)},
		conn:        conn,
		powClient:   proto.NewPowergateServiceClient(conn),
	}
	return client, nil
}

// ApplyStorageConfigOption mutates a push request.
type ApplyStorageConfigOption func(r *proto.ApplyStorageConfigRequest)

// WithStorageConfig overrides the Api default Cid configuration.
func WithStorageConfig(c *proto.StorageConfig) ApplyStorageConfigOption {
	return func(r *proto.ApplyStorageConfigRequest) {
		r.HasConfig = true
		r.Config = c
	}
}

// WithOverride allows a new push configuration to override an existing one.
// It's used as an extra security measure to avoid unwanted configuration changes.
func WithOverride(override bool) ApplyStorageConfigOption {
	return func(r *proto.ApplyStorageConfigRequest) {
		r.HasOverrideConfig = true
		r.OverrideConfig = override
	}
}

// WatchLogsOption is a function that changes GetLogsConfig.
type WatchLogsOption func(r *proto.WatchLogsRequest)

// WithJidFilter filters only log messages of a Cid related to
// the Job with id jid.
func WithJidFilter(jid string) WatchLogsOption {
	return func(r *proto.WatchLogsRequest) {
		r.Jid = jid
	}
}

// WithHistory indicates that prior history logs should
// be sent in the channel before getting real time logs.
func WithHistory(enabled bool) WatchLogsOption {
	return func(r *proto.WatchLogsRequest) {
		r.History = enabled
	}
}

// WatchLogsEvent represents an event for watching cid logs.
type WatchLogsEvent struct {
	Res *proto.WatchLogsResponse
	Err error
}

// Host returns the client host address.
func (c *Client) Host() string {
	return c.conn.Target()
}

// BuildInfo returns build info about the server.
func (c *Client) BuildInfo(ctx context.Context) (*proto.BuildInfoResponse, error) {
	return c.powClient.BuildInfo(ctx, &proto.BuildInfoRequest{})
}

// ID returns the FFS instance ID.
func (c *Client) ID(ctx context.Context) (*proto.IDResponse, error) {
	return c.powClient.ID(ctx, &proto.IDRequest{})
}

// DefaultStorageConfig returns the default storage config.
func (c *Client) DefaultStorageConfig(ctx context.Context) (*proto.DefaultStorageConfigResponse, error) {
	return c.powClient.DefaultStorageConfig(ctx, &proto.DefaultStorageConfigRequest{})
}

// SetDefaultStorageConfig sets the default storage config.
func (c *Client) SetDefaultStorageConfig(ctx context.Context, config *proto.StorageConfig) (*proto.SetDefaultStorageConfigResponse, error) {
	req := &proto.SetDefaultStorageConfigRequest{
		Config: config,
	}
	return c.powClient.SetDefaultStorageConfig(ctx, req)
}

// Stage allows to temporarily stage data in the Hot Storage in preparation for pushing a cid storage config.
func (c *Client) Stage(ctx context.Context, data io.Reader) (*proto.StageResponse, error) {
	stream, err := c.powClient.Stage(ctx)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024*32) // 32KB
	for {
		bytesRead, err := data.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		sendErr := stream.Send(&proto.StageRequest{Chunk: buffer[:bytesRead]})
		if sendErr != nil {
			if sendErr == io.EOF {
				var noOp interface{}
				return nil, stream.RecvMsg(noOp)
			}
			return nil, sendErr
		}
		if err == io.EOF {
			break
		}
	}
	return stream.CloseAndRecv()
}

// StageFolder allows to temporarily stage a folder in the Hot Storage in preparation for pushing a cid storage config.
func (c *Client) StageFolder(ctx context.Context, ipfsRevProxyAddr string, folderPath string) (string, error) {
	ffsToken := ctx.Value(AuthKey).(string)

	ipfs, err := newDecoratedIPFSAPI(ipfsRevProxyAddr, ffsToken)
	if err != nil {
		return "", fmt.Errorf("creating IPFS HTTP client: %s", err)
	}

	stat, err := os.Lstat(folderPath)
	if err != nil {
		return "", err
	}
	ff, err := files.NewSerialFile(folderPath, false, stat)
	if err != nil {
		return "", err
	}
	defer func() { _ = ff.Close() }()
	opts := []options.UnixfsAddOption{
		options.Unixfs.CidVersion(1),
		options.Unixfs.Pin(false),
	}
	pth, err := ipfs.Unixfs().Add(context.Background(), files.ToDir(ff), opts...)
	if err != nil {
		return "", err
	}

	return pth.Cid().String(), nil
}

// ApplyStorageConfig push a new configuration for the Cid in the Hot and Cold layers.
func (c *Client) ApplyStorageConfig(ctx context.Context, cid string, opts ...ApplyStorageConfigOption) (*proto.ApplyStorageConfigResponse, error) {
	req := &proto.ApplyStorageConfigRequest{Cid: cid}
	for _, opt := range opts {
		opt(req)
	}
	return c.powClient.ApplyStorageConfig(ctx, req)
}

// ReplaceData pushes a StorageConfig for c2 equal to that of c1, and removes c1. This operation
// is more efficient than manually removing and adding in two separate operations.
func (c *Client) ReplaceData(ctx context.Context, cid1, cid2 string) (*proto.ReplaceDataResponse, error) {
	return c.powClient.ReplaceData(ctx, &proto.ReplaceDataRequest{Cid1: cid1, Cid2: cid2})
}

// Get returns an io.Reader for reading a stored Cid from the Hot Storage.
func (c *Client) Get(ctx context.Context, cid string) (io.Reader, error) {
	stream, err := c.powClient.Get(ctx, &proto.GetRequest{
		Cid: cid,
	})
	if err != nil {
		return nil, err
	}
	reader, writer := io.Pipe()
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				_ = writer.Close()
				break
			} else if err != nil {
				_ = writer.CloseWithError(err)
				break
			}
			_, err = writer.Write(res.GetChunk())
			if err != nil {
				_ = writer.CloseWithError(err)
				break
			}
		}
	}()

	return reader, nil
}

// GetFolder retrieves to outputDir a Cid which corresponds to a folder.
func (c *Client) GetFolder(ctx context.Context, ipfsRevProxyAddr, cid, outputDir string) error {
	token := ctx.Value(AuthKey).(string)
	ipfs, err := newDecoratedIPFSAPI(ipfsRevProxyAddr, token)
	if err != nil {
		return fmt.Errorf("creating decorated IPFS client: %s", err)
	}
	ci, err := util.CidFromString(cid)
	if err != nil {
		return fmt.Errorf("decoding cid: %s", err)
	}
	n, err := ipfs.Unixfs().Get(ctx, ipfspath.IpfsPath(ci))
	if err != nil {
		return fmt.Errorf("getting folder DAG from IPFS: %s", err)
	}
	err = files.WriteTo(n, outputDir)
	if err != nil {
		return fmt.Errorf("saving folder DAG to output folder: %s", err)
	}
	return nil
}

// Remove removes a Cid from being tracked as an active storage. The Cid should have
// both Hot and Cold storage disabled, if that isn't the case it will return ErrActiveInStorage.
func (c *Client) Remove(ctx context.Context, cid string) (*proto.RemoveResponse, error) {
	return c.powClient.Remove(ctx, &proto.RemoveRequest{Cid: cid})
}

// WatchLogs pushes human-friendly messages about Cid executions. The method is blocking
// and will continue to send messages until the context is canceled. The provided channel
// is owned by the method and must not be closed.
func (c *Client) WatchLogs(ctx context.Context, ch chan<- WatchLogsEvent, cid string, opts ...WatchLogsOption) error {
	r := &proto.WatchLogsRequest{Cid: cid}
	for _, opt := range opts {
		opt(r)
	}
	stream, err := c.powClient.WatchLogs(ctx, r)
	if err != nil {
		return err
	}
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF || status.Code(err) == codes.Canceled {
				close(ch)
				break
			}
			if err != nil {
				ch <- WatchLogsEvent{Err: err}
				close(ch)
				break
			}
			ch <- WatchLogsEvent{Res: res}
		}
	}()
	return nil
}

// CidInfo returns information about cids managed by the FFS instance.
func (c *Client) CidInfo(ctx context.Context, cids ...string) (*proto.CidInfoResponse, error) {
	return c.powClient.CidInfo(ctx, &proto.CidInfoRequest{Cids: cids})
}

// Close closes the client's grpc connection and cancels any active requests.
func (c *Client) Close() error {
	return c.conn.Close()
}

func newDecoratedIPFSAPI(proxyAddr, ffsToken string) (*httpapi.HttpApi, error) {
	ipport := strings.Split(proxyAddr, ":")
	if len(ipport) != 2 {
		return nil, fmt.Errorf("ipfs addr is invalid")
	}
	cm, err := multiaddr.NewComponent("dns4", ipport[0])
	if err != nil {
		return nil, err
	}
	cp, err := multiaddr.NewComponent("tcp", ipport[1])
	if err != nil {
		return nil, err
	}
	useHTTPS := ipport[1] == "443"
	ipfsMaddr := cm.Encapsulate(cp)
	customClient := http.DefaultClient
	customClient.Transport = newFFSHeaderDecorator(ffsToken, useHTTPS)
	ipfs, err := httpapi.NewApiWithClient(ipfsMaddr, customClient)
	if err != nil {
		return nil, err
	}
	return ipfs, nil
}

type ffsHeaderDecorator struct {
	ffsToken string
	useHTTPS bool
}

func newFFSHeaderDecorator(ffsToken string, useHTTPS bool) *ffsHeaderDecorator {
	return &ffsHeaderDecorator{
		ffsToken: ffsToken,
		useHTTPS: useHTTPS,
	}
}

func (fhd ffsHeaderDecorator) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header["x-ipfs-ffs-auth"] = []string{fhd.ffsToken}
	if fhd.useHTTPS {
		req.URL.Scheme = "https"
	}

	return http.DefaultTransport.RoundTrip(req)
}
