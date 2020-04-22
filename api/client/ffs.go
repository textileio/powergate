package client

import (
	"context"
	"io"
	"time"

	cid "github.com/ipfs/go-cid"
	ff "github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ffs struct {
	client rpc.FFSClient
}

// JobEvent represents an event for Watching a job
type JobEvent struct {
	Job ff.Job
	Err error
}

// NewAddressOption is a function that changes a NewAddressConfig
type NewAddressOption func(r *rpc.NewAddrRequest)

// WithMakeDefault specifies if the new address should become the default
func WithMakeDefault(makeDefault bool) NewAddressOption {
	return func(r *rpc.NewAddrRequest) {
		r.MakeDefault = makeDefault
	}
}

// WithAddressType specifies the type of address to create
func WithAddressType(addressType string) NewAddressOption {
	return func(r *rpc.NewAddrRequest) {
		r.AddressType = addressType
	}
}

// PushConfigOption mutates a push request.
type PushConfigOption func(r *rpc.PushConfigRequest)

// WithCidConfig overrides the Api default Cid configuration.
func WithCidConfig(c ff.CidConfig) PushConfigOption {
	return func(r *rpc.PushConfigRequest) {
		r.HasConfig = true
		r.Config = &rpc.CidConfig{
			Cid:  c.Cid.String(),
			Hot:  toRpcHotConfig(c.Hot),
			Cold: toRpcColdConfig(c.Cold),
		}
	}
}

// WithOverride allows a new push configuration to override an existing one.
// It's used as an extra security measure to avoid unwanted configuration changes.
func WithOverride(override bool) PushConfigOption {
	return func(r *rpc.PushConfigRequest) {
		r.HasOverrideConfig = true
		r.OverrideConfig = override
	}
}

// WatchLogsOption is a function that changes GetLogsConfig.
type WatchLogsOption func(r *rpc.WatchLogsRequest)

// WithJidFilter filters only log messages of a Cid related to
// the Job with id jid.
func WithJidFilter(jid ff.JobID) WatchLogsOption {
	return func(r *rpc.WatchLogsRequest) {
		r.Jid = jid.String()
	}
}

// LogEvent represents an event for watching cid logs
type LogEvent struct {
	LogEntry ff.LogEntry
	Err      error
}

func (f *ffs) Create(ctx context.Context) (string, string, error) {
	r, err := f.client.Create(ctx, &rpc.CreateRequest{})
	if err != nil {
		return "", "", err
	}
	return r.ID, r.Token, nil
}

func (f *ffs) ID(ctx context.Context) (ff.APIID, error) {
	resp, err := f.client.ID(ctx, &rpc.IDRequest{})
	if err != nil {
		return ff.EmptyInstanceID, err
	}
	return ff.APIID(resp.ID), nil
}

func (f *ffs) Addrs(ctx context.Context) ([]api.AddrInfo, error) {
	resp, err := f.client.Addrs(ctx, &rpc.AddrsRequest{})
	if err != nil {
		return nil, err
	}
	addrs := make([]api.AddrInfo, len(resp.Addrs))
	for i, addr := range resp.Addrs {
		addrs[i] = api.AddrInfo{
			Name: addr.Name,
			Addr: addr.Addr,
		}
	}
	return addrs, nil
}

func (f *ffs) DefaultConfig(ctx context.Context) (ff.DefaultConfig, error) {
	resp, err := f.client.DefaultConfig(ctx, &rpc.DefaultConfigRequest{})
	if err != nil {
		return ff.DefaultConfig{}, err
	}
	return ff.DefaultConfig{
		Hot: ff.HotConfig{
			Enabled:       resp.DefaultConfig.Hot.Enabled,
			AllowUnfreeze: resp.DefaultConfig.Hot.AllowUnfreeze,
			Ipfs: ff.IpfsConfig{
				AddTimeout: int(resp.DefaultConfig.Hot.Ipfs.AddTimeout),
			},
		},
		Cold: ff.ColdConfig{
			Enabled: resp.DefaultConfig.Cold.Enabled,
			Filecoin: ff.FilConfig{
				RepFactor:      int(resp.DefaultConfig.Cold.Filecoin.RepFactor),
				DealDuration:   resp.DefaultConfig.Cold.Filecoin.DealDuration,
				ExcludedMiners: resp.DefaultConfig.Cold.Filecoin.ExcludedMiners,
				CountryCodes:   resp.DefaultConfig.Cold.Filecoin.CountryCodes,
				Renew: ff.FilRenew{
					Enabled:   resp.DefaultConfig.Cold.Filecoin.Renew.Enabled,
					Threshold: int(resp.DefaultConfig.Cold.Filecoin.Renew.Threshold),
				},
				Addr: resp.DefaultConfig.Cold.Filecoin.Addr,
			},
		},
	}, nil
}

func (f *ffs) NewAddr(ctx context.Context, name string, options ...NewAddressOption) (string, error) {
	r := &rpc.NewAddrRequest{Name: name}
	for _, opt := range options {
		opt(r)
	}
	resp, err := f.client.NewAddr(ctx, r)
	return resp.Addr, err
}

func (f *ffs) GetDefaultCidConfig(ctx context.Context, c cid.Cid) (*rpc.GetDefaultCidConfigReply, error) {
	return f.client.GetDefaultCidConfig(ctx, &rpc.GetDefaultCidConfigRequest{Cid: c.String()})
}

func (f *ffs) GetCidConfig(ctx context.Context, c cid.Cid) (*rpc.GetCidConfigReply, error) {
	return f.client.GetCidConfig(ctx, &rpc.GetCidConfigRequest{Cid: c.String()})
}

func (f *ffs) SetDefaultConfig(ctx context.Context, config ff.DefaultConfig) error {
	req := &rpc.SetDefaultConfigRequest{
		Config: &rpc.DefaultConfig{
			Hot:  toRpcHotConfig(config.Hot),
			Cold: toRpcColdConfig(config.Cold),
		},
	}
	_, err := f.client.SetDefaultConfig(ctx, req)
	return err
}

func (f *ffs) Show(ctx context.Context, c cid.Cid) (*rpc.ShowReply, error) {
	return f.client.Show(ctx, &rpc.ShowRequest{
		Cid: c.String(),
	})
}

func (f *ffs) Info(ctx context.Context) (*rpc.InfoReply, error) {
	return f.client.Info(ctx, &rpc.InfoRequest{})
}

func (f *ffs) WatchJobs(ctx context.Context, ch chan<- JobEvent, jids ...ff.JobID) error {
	jidStrings := make([]string, len(jids))
	for i, jid := range jids {
		jidStrings[i] = jid.String()
	}

	stream, err := f.client.WatchJobs(ctx, &rpc.WatchJobsRequest{Jids: jidStrings})
	if err != nil {
		return err
	}
	go func() {
		for {
			reply, err := stream.Recv()
			if err == io.EOF || status.Code(err) == codes.Canceled {
				close(ch)
				break
			}
			if err != nil {
				ch <- JobEvent{Err: err}
				close(ch)
				break
			}
			job := ff.Job{
				ID:       ff.JobID(reply.Job.ID),
				APIID:    ff.APIID(reply.Job.ApiID),
				Status:   ff.JobStatus(reply.Job.Status),
				ErrCause: reply.Job.ErrCause,
			}
			ch <- JobEvent{Job: job}
		}
	}()
	return nil
}

func (f *ffs) Replace(ctx context.Context, c1 cid.Cid, c2 cid.Cid) (ff.JobID, error) {
	resp, err := f.client.Replace(ctx, &rpc.ReplaceRequest{Cid1: c1.String(), Cid2: c2.String()})
	if err != nil {
		return ff.EmptyJobID, err
	}
	return ff.JobID(resp.JobID), nil
}

func (f *ffs) PushConfig(ctx context.Context, c cid.Cid, opts ...PushConfigOption) (ff.JobID, error) {
	req := &rpc.PushConfigRequest{Cid: c.String()}
	for _, opt := range opts {
		opt(req)
	}

	resp, err := f.client.PushConfig(ctx, req)
	if err != nil {
		return ff.EmptyJobID, err
	}

	return ff.JobID(resp.JobID), nil
}

func (f *ffs) Remove(ctx context.Context, c cid.Cid) error {
	_, err := f.client.Remove(ctx, &rpc.RemoveRequest{Cid: c.String()})
	return err
}

func (f *ffs) Get(ctx context.Context, c cid.Cid) (io.Reader, error) {
	stream, err := f.client.Get(ctx, &rpc.GetRequest{
		Cid: c.String(),
	})
	if err != nil {
		return nil, err
	}
	reader, writer := io.Pipe()
	go func() {
		for {
			reply, err := stream.Recv()
			if err == io.EOF {
				_ = writer.Close()
				break
			} else if err != nil {
				_ = writer.CloseWithError(err)
				break
			}
			_, err = writer.Write(reply.GetChunk())
			if err != nil {
				_ = writer.CloseWithError(err)
				break
			}
		}
	}()

	return reader, nil
}

func (f *ffs) WatchLogs(ctx context.Context, ch chan<- LogEvent, c cid.Cid, opts ...WatchLogsOption) error {
	r := &rpc.WatchLogsRequest{Cid: c.String()}
	for _, opt := range opts {
		opt(r)
	}

	stream, err := f.client.WatchLogs(ctx, r)
	if err != nil {
		return err
	}
	go func() {
		for {
			reply, err := stream.Recv()
			if err == io.EOF || status.Code(err) == codes.Canceled {
				close(ch)
				break
			}
			if err != nil {
				ch <- LogEvent{Err: err}
				close(ch)
				break
			}

			cid, err := cid.Decode(reply.LogEntry.Cid)
			if err != nil {
				ch <- LogEvent{Err: err}
				close(ch)
				break
			}

			entry := ff.LogEntry{
				Cid:       cid,
				Timestamp: time.Unix(reply.LogEntry.Time, 0),
				Jid:       ff.JobID(reply.LogEntry.Jid),
				Msg:       reply.LogEntry.Msg,
			}
			ch <- LogEvent{LogEntry: entry}
		}
	}()
	return nil
}

func (f *ffs) Close(ctx context.Context) error {
	_, err := f.client.Close(ctx, &rpc.CloseRequest{})
	return err
}

func (f *ffs) AddToHot(ctx context.Context, data io.Reader) (*cid.Cid, error) {
	stream, err := f.client.AddToHot(ctx)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024*32) // 32KB
	for {
		bytesRead, err := data.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		sendErr := stream.Send(&rpc.AddToHotRequest{Chunk: buffer[:bytesRead]})
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
	reply, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}

	cid, err := cid.Decode(reply.GetCid())
	if err != nil {
		return nil, err
	}
	return &cid, nil
}

func toRpcHotConfig(config ff.HotConfig) *rpc.HotConfig {
	return &rpc.HotConfig{
		Enabled:       config.Enabled,
		AllowUnfreeze: config.AllowUnfreeze,
		Ipfs: &rpc.IpfsConfig{
			AddTimeout: int64(config.Ipfs.AddTimeout),
		},
	}
}

func toRpcColdConfig(config ff.ColdConfig) *rpc.ColdConfig {
	return &rpc.ColdConfig{
		Enabled: config.Enabled,
		Filecoin: &rpc.FilConfig{
			RepFactor:      int64(config.Filecoin.RepFactor),
			DealDuration:   int64(config.Filecoin.DealDuration),
			ExcludedMiners: config.Filecoin.ExcludedMiners,
			CountryCodes:   config.Filecoin.CountryCodes,
			Renew: &rpc.FilRenew{
				Enabled:   config.Filecoin.Renew.Enabled,
				Threshold: int64(config.Filecoin.Renew.Threshold),
			},
			Addr: config.Filecoin.Addr,
		},
	}
}
