package client

import (
	"context"
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
	"github.com/textileio/powergate/ffs/rpc"
	"github.com/textileio/powergate/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FFS provides the API to create and interact with an FFS instance.
type FFS struct {
	client rpc.RPCServiceClient
}

// WatchJobsEvent represents an event for Watching a job.
type WatchJobsEvent struct {
	Res *rpc.WatchJobsResponse
	Err error
}

// NewAddressOption is a function that changes a NewAddressConfig.
type NewAddressOption func(r *rpc.NewAddrRequest)

// WithMakeDefault specifies if the new address should become the default.
func WithMakeDefault(makeDefault bool) NewAddressOption {
	return func(r *rpc.NewAddrRequest) {
		r.MakeDefault = makeDefault
	}
}

// WithAddressType specifies the type of address to create.
func WithAddressType(addressType string) NewAddressOption {
	return func(r *rpc.NewAddrRequest) {
		r.AddressType = addressType
	}
}

// PushStorageConfigOption mutates a push request.
type PushStorageConfigOption func(r *rpc.PushStorageConfigRequest)

// WithStorageConfig overrides the Api default Cid configuration.
func WithStorageConfig(c *rpc.StorageConfig) PushStorageConfigOption {
	return func(r *rpc.PushStorageConfigRequest) {
		r.HasConfig = true
		r.Config = c
	}
}

// WithOverride allows a new push configuration to override an existing one.
// It's used as an extra security measure to avoid unwanted configuration changes.
func WithOverride(override bool) PushStorageConfigOption {
	return func(r *rpc.PushStorageConfigRequest) {
		r.HasOverrideConfig = true
		r.OverrideConfig = override
	}
}

// WatchLogsOption is a function that changes GetLogsConfig.
type WatchLogsOption func(r *rpc.WatchLogsRequest)

// WithJidFilter filters only log messages of a Cid related to
// the Job with id jid.
func WithJidFilter(jid string) WatchLogsOption {
	return func(r *rpc.WatchLogsRequest) {
		r.Jid = jid
	}
}

// WithHistory indicates that prior history logs should
// be sent in the channel before getting real time logs.
func WithHistory(enabled bool) WatchLogsOption {
	return func(r *rpc.WatchLogsRequest) {
		r.History = enabled
	}
}

// WatchLogsEvent represents an event for watching cid logs.
type WatchLogsEvent struct {
	Res *rpc.WatchLogsResponse
	Err error
}

// ListDealRecordsOption updates a ListDealRecordsConfig.
type ListDealRecordsOption func(*rpc.ListDealRecordsConfig)

// WithFromAddrs limits the results deals initiated from the provided wallet addresses.
// If WithDataCids is also provided, this is an AND operation.
func WithFromAddrs(addrs ...string) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.FromAddrs = addrs
	}
}

// WithDataCids limits the results to deals for the provided data cids.
// If WithFromAddrs is also provided, this is an AND operation.
func WithDataCids(cids ...string) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.DataCids = cids
	}
}

// WithIncludePending specifies whether or not to include pending deals in the results. Default is false.
// Ignored for ListRetrievalDealRecords.
func WithIncludePending(includePending bool) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.IncludePending = includePending
	}
}

// WithIncludeFinal specifies whether or not to include final deals in the results. Default is false.
// Ignored for ListRetrievalDealRecords.
func WithIncludeFinal(includeFinal bool) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.IncludeFinal = includeFinal
	}
}

// WithAscending specifies to sort the results in ascending order. Default is descending order.
// Records are sorted by timestamp.
func WithAscending(ascending bool) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.Ascending = ascending
	}
}

// ID returns the FFS instance ID.
func (f *FFS) ID(ctx context.Context) (*rpc.IDResponse, error) {
	return f.client.ID(ctx, &rpc.IDRequest{})
}

// Addrs returns a list of addresses managed by the FFS instance.
func (f *FFS) Addrs(ctx context.Context) (*rpc.AddrsResponse, error) {
	return f.client.Addrs(ctx, &rpc.AddrsRequest{})
}

// DefaultStorageConfig returns the default storage config.
func (f *FFS) DefaultStorageConfig(ctx context.Context) (*rpc.DefaultStorageConfigResponse, error) {
	return f.client.DefaultStorageConfig(ctx, &rpc.DefaultStorageConfigRequest{})
}

// SignMessage signs a message with a FFS managed wallet address.
func (f *FFS) SignMessage(ctx context.Context, addr string, message []byte) (*rpc.SignMessageResponse, error) {
	r := &rpc.SignMessageRequest{Addr: addr, Msg: message}
	return f.client.SignMessage(ctx, r)
}

// VerifyMessage verifies a message signature from a wallet address.
func (f *FFS) VerifyMessage(ctx context.Context, addr string, message, signature []byte) (*rpc.VerifyMessageResponse, error) {
	r := &rpc.VerifyMessageRequest{Addr: addr, Msg: message, Signature: signature}
	return f.client.VerifyMessage(ctx, r)
}

// NewAddr created a new wallet address managed by the FFS instance.
func (f *FFS) NewAddr(ctx context.Context, name string, options ...NewAddressOption) (*rpc.NewAddrResponse, error) {
	r := &rpc.NewAddrRequest{Name: name}
	for _, opt := range options {
		opt(r)
	}
	return f.client.NewAddr(ctx, r)
}

// SetDefaultStorageConfig sets the default storage config.
func (f *FFS) SetDefaultStorageConfig(ctx context.Context, config *rpc.StorageConfig) (*rpc.SetDefaultStorageConfigResponse, error) {
	req := &rpc.SetDefaultStorageConfigRequest{
		Config: config,
	}
	return f.client.SetDefaultStorageConfig(ctx, req)
}

// CidInfo returns information about cids managed by the FFS instance.
func (f *FFS) CidInfo(ctx context.Context, cids ...string) (*rpc.CidInfoResponse, error) {
	return f.client.CidInfo(ctx, &rpc.CidInfoRequest{Cids: cids})
}

// CancelJob signals that the executing Job with JobID jid should be
// canceled.
func (f *FFS) CancelJob(ctx context.Context, jid string) (*rpc.CancelJobResponse, error) {
	return f.client.CancelJob(ctx, &rpc.CancelJobRequest{Jid: jid})
}

// StorageJob returns the current state of the specified job.
func (f *FFS) StorageJob(ctx context.Context, jid string) (*rpc.StorageJobResponse, error) {
	return f.client.StorageJob(ctx, &rpc.StorageJobRequest{Jid: jid})
}

// QueuedStorageJobs returns a list of queued storage jobs.
func (f *FFS) QueuedStorageJobs(ctx context.Context, cids ...string) (*rpc.QueuedStorageJobsResponse, error) {
	req := &rpc.QueuedStorageJobsRequest{
		Cids: cids,
	}
	return f.client.QueuedStorageJobs(ctx, req)
}

// ExecutingStorageJobs returns a list of executing storage jobs.
func (f *FFS) ExecutingStorageJobs(ctx context.Context, cids ...string) (*rpc.ExecutingStorageJobsResponse, error) {
	req := &rpc.ExecutingStorageJobsRequest{
		Cids: cids,
	}
	return f.client.ExecutingStorageJobs(ctx, req)
}

// LatestFinalStorageJobs returns a list of latest final storage jobs.
func (f *FFS) LatestFinalStorageJobs(ctx context.Context, cids ...string) (*rpc.LatestFinalStorageJobsResponse, error) {
	req := &rpc.LatestFinalStorageJobsRequest{
		Cids: cids,
	}
	return f.client.LatestFinalStorageJobs(ctx, req)
}

// LatestSuccessfulStorageJobs returns a list of latest successful storage jobs.
func (f *FFS) LatestSuccessfulStorageJobs(ctx context.Context, cids ...string) (*rpc.LatestSuccessfulStorageJobsResponse, error) {
	req := &rpc.LatestSuccessfulStorageJobsRequest{
		Cids: cids,
	}
	return f.client.LatestSuccessfulStorageJobs(ctx, req)
}

// StorageJobsSummary returns a summary of storage jobs.
func (f *FFS) StorageJobsSummary(ctx context.Context, cids ...string) (*rpc.StorageJobsSummaryResponse, error) {
	req := &rpc.StorageJobsSummaryRequest{
		Cids: cids,
	}
	return f.client.StorageJobsSummary(ctx, req)
}

// WatchJobs pushes JobEvents to the provided channel. The provided channel will be owned
// by the client after the call, so it shouldn't be closed by the client. To stop receiving
// events, the provided ctx should be canceled. If an error occurs, it will be returned
// in the Err field of JobEvent and the channel will be closed.
func (f *FFS) WatchJobs(ctx context.Context, ch chan<- WatchJobsEvent, jids ...string) error {
	stream, err := f.client.WatchJobs(ctx, &rpc.WatchJobsRequest{Jids: jids})
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
				ch <- WatchJobsEvent{Err: err}
				close(ch)
				break
			}
			ch <- WatchJobsEvent{Res: res}
		}
	}()
	return nil
}

// Replace pushes a StorageConfig for c2 equal to that of c1, and removes c1. This operation
// is more efficient than manually removing and adding in two separate operations.
func (f *FFS) Replace(ctx context.Context, cid1, cid2 string) (*rpc.ReplaceResponse, error) {
	return f.client.Replace(ctx, &rpc.ReplaceRequest{Cid1: cid1, Cid2: cid2})
}

// PushStorageConfig push a new configuration for the Cid in the Hot and Cold layers.
func (f *FFS) PushStorageConfig(ctx context.Context, cid string, opts ...PushStorageConfigOption) (*rpc.PushStorageConfigResponse, error) {
	req := &rpc.PushStorageConfigRequest{Cid: cid}
	for _, opt := range opts {
		opt(req)
	}
	return f.client.PushStorageConfig(ctx, req)
}

// Remove removes a Cid from being tracked as an active storage. The Cid should have
// both Hot and Cold storage disabled, if that isn't the case it will return ErrActiveInStorage.
func (f *FFS) Remove(ctx context.Context, cid string) (*rpc.RemoveResponse, error) {
	return f.client.Remove(ctx, &rpc.RemoveRequest{Cid: cid})
}

// GetFolder retrieves to outputDir a Cid which corresponds to a folder.
func (f *FFS) GetFolder(ctx context.Context, ipfsRevProxyAddr, cid, outputDir string) error {
	token := ctx.Value(AuthKey).(string)
	ipfs, err := newDecoratedIPFSAPI(ipfsRevProxyAddr, token)
	if err != nil {
		return fmt.Errorf("creating decorated IPFS client: %s", err)
	}
	c, err := util.CidFromString(cid)
	if err != nil {
		return fmt.Errorf("decoding cid: %s", err)
	}
	n, err := ipfs.Unixfs().Get(ctx, ipfspath.IpfsPath(c))
	if err != nil {
		return fmt.Errorf("getting folder DAG from IPFS: %s", err)
	}
	err = files.WriteTo(n, outputDir)
	if err != nil {
		return fmt.Errorf("saving folder DAG to output folder: %s", err)
	}
	return nil
}

// Get returns an io.Reader for reading a stored Cid from the Hot Storage.
func (f *FFS) Get(ctx context.Context, cid string) (io.Reader, error) {
	stream, err := f.client.Get(ctx, &rpc.GetRequest{
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

// WatchLogs pushes human-friendly messages about Cid executions. The method is blocking
// and will continue to send messages until the context is canceled. The provided channel
// is owned by the method and must not be closed.
func (f *FFS) WatchLogs(ctx context.Context, ch chan<- WatchLogsEvent, cid string, opts ...WatchLogsOption) error {
	r := &rpc.WatchLogsRequest{Cid: cid}
	for _, opt := range opts {
		opt(r)
	}
	stream, err := f.client.WatchLogs(ctx, r)
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

// SendFil sends fil from a managed address to any another address, returns immediately but funds are sent asynchronously.
func (f *FFS) SendFil(ctx context.Context, from string, to string, amount int64) (*rpc.SendFilResponse, error) {
	req := &rpc.SendFilRequest{
		From:   from,
		To:     to,
		Amount: amount,
	}
	return f.client.SendFil(ctx, req)
}

// Stage allows to temporarily stage data in the Hot Storage in preparation for pushing a cid storage config.
func (f *FFS) Stage(ctx context.Context, data io.Reader) (*rpc.StageResponse, error) {
	stream, err := f.client.Stage(ctx)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024*32) // 32KB
	for {
		bytesRead, err := data.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		sendErr := stream.Send(&rpc.StageRequest{Chunk: buffer[:bytesRead]})
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
func (f *FFS) StageFolder(ctx context.Context, ipfsRevProxyAddr string, folderPath string) (string, error) {
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

// ListPayChannels returns a list of payment channels.
func (f *FFS) ListPayChannels(ctx context.Context) (*rpc.ListPayChannelsResponse, error) {
	return f.client.ListPayChannels(ctx, &rpc.ListPayChannelsRequest{})
}

// CreatePayChannel creates a new payment channel.
func (f *FFS) CreatePayChannel(ctx context.Context, from, to string, amount uint64) (*rpc.CreatePayChannelResponse, error) {
	req := &rpc.CreatePayChannelRequest{
		From:   from,
		To:     to,
		Amount: amount,
	}
	return f.client.CreatePayChannel(ctx, req)
}

// RedeemPayChannel redeems a payment channel.
func (f *FFS) RedeemPayChannel(ctx context.Context, addr string) (*rpc.RedeemPayChannelResponse, error) {
	req := &rpc.RedeemPayChannelRequest{PayChannelAddr: addr}
	return f.client.RedeemPayChannel(ctx, req)
}

// ListStorageDealRecords returns a list of storage deals for the FFS instance according to the provided options.
func (f *FFS) ListStorageDealRecords(ctx context.Context, opts ...ListDealRecordsOption) (*rpc.ListStorageDealRecordsResponse, error) {
	conf := &rpc.ListDealRecordsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	return f.client.ListStorageDealRecords(ctx, &rpc.ListStorageDealRecordsRequest{Config: conf})
}

// ListRetrievalDealRecords returns a list of retrieval deals for the FFS instance according to the provided options.
func (f *FFS) ListRetrievalDealRecords(ctx context.Context, opts ...ListDealRecordsOption) (*rpc.ListRetrievalDealRecordsResponse, error) {
	conf := &rpc.ListDealRecordsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	return f.client.ListRetrievalDealRecords(ctx, &rpc.ListRetrievalDealRecordsRequest{Config: conf})
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
