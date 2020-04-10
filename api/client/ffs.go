package client

import (
	"context"
	"io"

	cid "github.com/ipfs/go-cid"
	ff "github.com/textileio/powergate/ffs"
	pb "github.com/textileio/powergate/ffs/pb"
)

type ffs struct {
	client pb.APIClient
}

// JobEvent represents an event for Watching a job
type JobEvent struct {
	Job ff.Job
	Err error
}

// PushConfigOption mutates a push configuration.
type PushConfigOption func(o *PushConfig)

// PushConfig contains options for pushing a Cid configuration.
type PushConfig struct {
	Config            ff.CidConfig
	HasConfig         bool
	OverrideConfig    bool
	HasOverrideConfig bool
}

// WithCidConfig overrides the Api default Cid configuration.
func WithCidConfig(c ff.CidConfig) PushConfigOption {
	return func(o *PushConfig) {
		o.Config = c
		o.HasConfig = true
	}
}

// WithOverride allows a new push configuration to override an existing one.
// It's used as an extra security measure to avoid unwanted configuration changes.
func WithOverride(override bool) PushConfigOption {
	return func(o *PushConfig) {
		o.OverrideConfig = override
		o.HasOverrideConfig = true
	}
}

func (f *ffs) Create(ctx context.Context) (string, string, error) {
	r, err := f.client.Create(ctx, &pb.CreateRequest{})
	if err != nil {
		return "", "", err
	}
	return r.ID, r.Token, nil
}

func (f *ffs) ID(ctx context.Context) (ff.ApiID, error) {
	resp, err := f.client.ID(ctx, &pb.IDRequest{})
	if err != nil {
		return ff.EmptyInstanceID, err
	}
	return ff.ApiID(resp.ID), nil
}

func (f *ffs) WalletAddr(ctx context.Context) (string, error) {
	resp, err := f.client.WalletAddr(ctx, &pb.WalletAddrRequest{})
	if err != nil {
		return "", err
	}
	return resp.Addr, nil
}

func (f *ffs) GetDefaultCidConfig(ctx context.Context, c cid.Cid) (*pb.GetDefaultCidConfigReply, error) {
	return f.client.GetDefaultCidConfig(ctx, &pb.GetDefaultCidConfigRequest{Cid: c.String()})
}

func (f *ffs) GetCidConfig(ctx context.Context, c cid.Cid) (*pb.GetCidConfigReply, error) {
	return f.client.GetCidConfig(ctx, &pb.GetCidConfigRequest{Cid: c.String()})
}

func (f *ffs) SetDefaultCidConfig(ctx context.Context, config ff.DefaultCidConfig) error {
	req := &pb.SetDefaultCidConfigRequest{
		Config: &pb.DefaultCidConfig{
			Hot: &pb.HotConfig{
				Enabled:       config.Hot.Enabled,
				AllowUnfreeze: config.Hot.AllowUnfreeze,
				Ipfs: &pb.IpfsConfig{
					AddTimeout: int64(config.Hot.Ipfs.AddTimeout),
				},
			},
			Cold: &pb.ColdConfig{
				Enabled: config.Cold.Enabled,
				Filecoin: &pb.FilConfig{
					RepFactor:      int64(config.Cold.Filecoin.RepFactor),
					DealDuration:   int64(config.Cold.Filecoin.DealDuration),
					ExcludedMiners: config.Cold.Filecoin.ExcludedMiners,
					CountryCodes:   config.Cold.Filecoin.CountryCodes,
					Renew: &pb.FilRenew{
						Enabled:   config.Cold.Filecoin.Renew.Enabled,
						Threshold: int64(config.Cold.Filecoin.Renew.Threshold),
					},
				},
			},
		},
	}
	_, err := f.client.SetDefaultCidConfig(ctx, req)
	return err
}

func (f *ffs) Show(ctx context.Context, c cid.Cid) (*pb.ShowReply, error) {
	return f.client.Show(ctx, &pb.ShowRequest{
		Cid: c.String(),
	})
}

func (f *ffs) Info(ctx context.Context) (*pb.InfoReply, error) {
	return f.client.Info(ctx, &pb.InfoRequest{})
}

func (f *ffs) Watch(ctx context.Context, jids ...ff.JobID) (<-chan JobEvent, func(), error) {
	updates := make(chan JobEvent)
	jidStrings := make([]string, len(jids))
	for i, jid := range jids {
		jidStrings[i] = jid.String()
	}

	ctx, cancel := context.WithCancel(ctx)
	cancelFunc := func() {
		cancel()
		close(updates)
	}

	stream, err := f.client.Watch(ctx, &pb.WatchRequest{Jids: jidStrings})
	if err != nil {
		return nil, nil, err
	}
	go func() {
		for {
			reply, err := stream.Recv()
			if err == io.EOF {
				close(updates)
				break
			}
			if err != nil {
				updates <- JobEvent{Err: err}
				close(updates)
				break
			}
			job := ff.Job{
				ID:         ff.JobID(reply.Job.ID),
				InstanceID: ff.ApiID(reply.Job.InstanceID),
				Status:     ff.JobStatus(reply.Job.Status),
				ErrCause:   reply.Job.ErrCause,
			}
			updates <- JobEvent{Job: job}
		}
	}()
	return updates, cancelFunc, nil
}

func (f *ffs) PushConfig(ctx context.Context, c cid.Cid, opts ...PushConfigOption) (ff.JobID, error) {
	pushConfig := PushConfig{}
	for _, opt := range opts {
		opt(&pushConfig)
	}

	req := &pb.PushConfigRequest{Cid: c.String()}

	if pushConfig.HasConfig {
		req.HasConfig = true
		req.Config = &pb.CidConfig{
			Cid: pushConfig.Config.Cid.String(),
			Hot: &pb.HotConfig{
				Enabled:       pushConfig.Config.Hot.Enabled,
				AllowUnfreeze: pushConfig.Config.Hot.AllowUnfreeze,
				Ipfs: &pb.IpfsConfig{
					AddTimeout: int64(pushConfig.Config.Hot.Ipfs.AddTimeout),
				},
			},
			Cold: &pb.ColdConfig{
				Enabled: pushConfig.Config.Cold.Enabled,
				Filecoin: &pb.FilConfig{
					RepFactor:      int64(pushConfig.Config.Cold.Filecoin.RepFactor),
					DealDuration:   pushConfig.Config.Cold.Filecoin.DealDuration,
					ExcludedMiners: pushConfig.Config.Cold.Filecoin.ExcludedMiners,
					CountryCodes:   pushConfig.Config.Cold.Filecoin.CountryCodes,
					Renew: &pb.FilRenew{
						Enabled:   pushConfig.Config.Cold.Filecoin.Renew.Enabled,
						Threshold: int64(pushConfig.Config.Cold.Filecoin.Renew.Threshold),
					},
				},
			},
		}
	}

	if pushConfig.HasOverrideConfig {
		req.HasOverrideConfig = true
		req.OverrideConfig = pushConfig.OverrideConfig
	}

	resp, err := f.client.PushConfig(ctx, req)
	if err != nil {
		return ff.EmptyJobID, err
	}

	return ff.JobID(resp.JobID), nil
}

func (f *ffs) Get(ctx context.Context, c cid.Cid) (io.Reader, error) {
	stream, err := f.client.Get(ctx, &pb.GetRequest{
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

func (f *ffs) Close(ctx context.Context) error {
	_, err := f.client.Close(ctx, &pb.CloseRequest{})
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
		sendErr := stream.Send(&pb.AddToHotRequest{Chunk: buffer[:bytesRead]})
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
