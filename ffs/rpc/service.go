package rpc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/ipfs/go-cid"
	logger "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/manager"
)

var (
	// ErrEmptyAuthToken is returned when the provided auth-token is unknown.
	ErrEmptyAuthToken = errors.New("auth token can't be empty")

	log = logger.Logger("ffs-grpc-service")
)

// Service implements the proto service definition of FFS.
type Service struct {
	UnimplementedFFSAPIServer

	m   *manager.Manager
	hot ffs.HotStorage
}

// NewService returns a new Service.
func NewService(m *manager.Manager, hot ffs.HotStorage) *Service {
	return &Service{
		m:   m,
		hot: hot,
	}
}

// Create creates a new Api.
func (s *Service) Create(ctx context.Context, req *CreateRequest) (*CreateReply, error) {
	id, token, err := s.m.Create(ctx)
	if err != nil {
		log.Errorf("creating instance: %s", err)
		return nil, err
	}
	return &CreateReply{
		ID:    id.String(),
		Token: token,
	}, nil
}

// ID returns the API instance id
func (s *Service) ID(ctx context.Context, req *IDRequest) (*IDReply, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	id := i.ID()
	return &IDReply{ID: id.String()}, nil
}

// WalletAddr returns the wallet address
func (s *Service) WalletAddr(ctx context.Context, req *WalletAddrRequest) (*WalletAddrReply, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	addr := i.WalletAddr()
	return &WalletAddrReply{Addr: addr}, nil
}

// GetDefaultCidConfig returns the default cid config prepped for the provided cid
func (s *Service) GetDefaultCidConfig(ctx context.Context, req *GetDefaultCidConfigRequest) (*GetDefaultCidConfigReply, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	c, err := cid.Decode(req.Cid)
	if err != nil {
		return nil, err
	}
	config := i.GetDefaultCidConfig(c)
	return &GetDefaultCidConfigReply{
		Config: &CidConfig{
			Cid: config.Cid.String(),
			Hot: &HotConfig{
				Enabled:       config.Hot.Enabled,
				AllowUnfreeze: config.Hot.AllowUnfreeze,
				Ipfs: &IpfsConfig{
					AddTimeout: int64(config.Hot.Ipfs.AddTimeout),
				},
			},
			Cold: &ColdConfig{
				Enabled: config.Cold.Enabled,
				Filecoin: &FilConfig{
					RepFactor:      int64(config.Cold.Filecoin.RepFactor),
					DealDuration:   int64(config.Cold.Filecoin.DealDuration),
					ExcludedMiners: config.Cold.Filecoin.ExcludedMiners,
					CountryCodes:   config.Cold.Filecoin.CountryCodes,
					Renew: &FilRenew{
						Enabled:   config.Cold.Filecoin.Renew.Enabled,
						Threshold: int64(config.Cold.Filecoin.Renew.Threshold),
					},
				},
			},
		},
	}, nil
}

// GetCidConfig returns the cid config for the provided cid
func (s *Service) GetCidConfig(ctx context.Context, req *GetCidConfigRequest) (*GetCidConfigReply, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	c, err := cid.Decode(req.Cid)
	if err != nil {
		return nil, err
	}
	config, err := i.GetCidConfig(c)
	if err != nil {
		return nil, err
	}
	return &GetCidConfigReply{
		Config: &CidConfig{
			Cid: config.Cid.String(),
			Hot: &HotConfig{
				Enabled:       config.Hot.Enabled,
				AllowUnfreeze: config.Hot.AllowUnfreeze,
				Ipfs: &IpfsConfig{
					AddTimeout: int64(config.Hot.Ipfs.AddTimeout),
				},
			},
			Cold: &ColdConfig{
				Enabled: config.Cold.Enabled,
				Filecoin: &FilConfig{
					RepFactor:      int64(config.Cold.Filecoin.RepFactor),
					DealDuration:   int64(config.Cold.Filecoin.DealDuration),
					ExcludedMiners: config.Cold.Filecoin.ExcludedMiners,
					CountryCodes:   config.Cold.Filecoin.CountryCodes,
					Renew: &FilRenew{
						Enabled:   config.Cold.Filecoin.Renew.Enabled,
						Threshold: int64(config.Cold.Filecoin.Renew.Threshold),
					},
				},
			},
		},
	}, nil
}

// SetDefaultCidConfig sets a new config to be used by default
func (s *Service) SetDefaultCidConfig(ctx context.Context, req *SetDefaultCidConfigRequest) (*SetDefaultCidConfigReply, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	defaultConfig := ffs.DefaultCidConfig{
		Hot: ffs.HotConfig{
			Enabled:       req.Config.Hot.Enabled,
			AllowUnfreeze: req.Config.Hot.AllowUnfreeze,
			Ipfs: ffs.IpfsConfig{
				AddTimeout: int(req.Config.Hot.Ipfs.AddTimeout),
			},
		},
		Cold: ffs.ColdConfig{
			Enabled: req.Config.Cold.Enabled,
			Filecoin: ffs.FilConfig{
				RepFactor:      int(req.Config.Cold.Filecoin.RepFactor),
				DealDuration:   req.Config.Cold.Filecoin.DealDuration,
				ExcludedMiners: req.Config.Cold.Filecoin.ExcludedMiners,
				CountryCodes:   req.Config.Cold.Filecoin.CountryCodes,
				Renew: ffs.FilRenew{
					Enabled:   req.Config.Cold.Filecoin.Renew.Enabled,
					Threshold: int(req.Config.Cold.Filecoin.Renew.Threshold),
				},
			},
		},
	}
	if err := i.SetDefaultCidConfig(defaultConfig); err != nil {
		return nil, err
	}
	return &SetDefaultCidConfigReply{}, nil
}

// Show returns information about a particular Cid.
func (s *Service) Show(ctx context.Context, req *ShowRequest) (*ShowReply, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}

	c, err := cid.Decode(req.GetCid())
	if err != nil {
		return nil, err
	}

	info, err := i.Show(c)
	if err != nil {
		return nil, err
	}
	reply := &ShowReply{
		CidInfo: &CidInfo{
			JobID:   info.JobID.String(),
			Cid:     info.Cid.String(),
			Created: info.Created.UnixNano(),
			Hot: &HotInfo{
				Enabled: info.Hot.Enabled,
				Size:    int64(info.Hot.Size),
				Ipfs: &IpfsHotInfo{
					Created: info.Hot.Ipfs.Created.UnixNano(),
				},
			},
			Cold: &ColdInfo{
				Filecoin: &FilInfo{
					DataCid:   info.Cold.Filecoin.DataCid.String(),
					Proposals: make([]*FilStorage, len(info.Cold.Filecoin.Proposals)),
				},
			},
		},
	}
	for i, p := range info.Cold.Filecoin.Proposals {
		reply.CidInfo.Cold.Filecoin.Proposals[i] = &FilStorage{
			ProposalCid:     p.ProposalCid.String(),
			Renewed:         p.Renewed,
			Duration:        p.Duration,
			ActivationEpoch: p.ActivationEpoch,
			Miner:           p.Miner,
		}
	}
	return reply, nil
}

// Info returns an Api information.
func (s *Service) Info(ctx context.Context, req *InfoRequest) (*InfoReply, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}

	info, err := i.Info(ctx)
	if err != nil {
		return nil, err
	}

	reply := &InfoReply{
		Info: &InstanceInfo{
			ID: info.ID.String(),
			DefaultCidConfig: &DefaultCidConfig{
				Hot: &HotConfig{
					Enabled:       info.DefaultCidConfig.Hot.Enabled,
					AllowUnfreeze: info.DefaultCidConfig.Hot.AllowUnfreeze,
					Ipfs: &IpfsConfig{
						AddTimeout: int64(info.DefaultCidConfig.Hot.Ipfs.AddTimeout),
					},
				},
				Cold: &ColdConfig{
					Enabled: info.DefaultCidConfig.Cold.Enabled,
					Filecoin: &FilConfig{
						RepFactor:      int64(info.DefaultCidConfig.Cold.Filecoin.RepFactor),
						DealDuration:   info.DefaultCidConfig.Cold.Filecoin.DealDuration,
						ExcludedMiners: info.DefaultCidConfig.Cold.Filecoin.ExcludedMiners,
						CountryCodes:   info.DefaultCidConfig.Cold.Filecoin.CountryCodes,
						Renew: &FilRenew{
							Enabled:   info.DefaultCidConfig.Cold.Filecoin.Renew.Enabled,
							Threshold: int64(info.DefaultCidConfig.Cold.Filecoin.Renew.Threshold),
						},
					},
				},
			},
			Wallet: &WalletInfo{
				Address: info.Wallet.Address,
				Balance: info.Wallet.Balance,
			},
			Pins: make([]string, len(info.Pins)),
		},
	}
	for i, p := range info.Pins {
		reply.Info.Pins[i] = p.String()
	}
	return reply, nil
}

// WatchJobs calls API.WatchJobs
func (s *Service) WatchJobs(req *WatchJobsRequest, srv FFSAPI_WatchJobsServer) error {
	i, err := s.getInstanceByToken(srv.Context())
	if err != nil {
		return err
	}

	jids := make([]ffs.JobID, len(req.Jids))
	for i, jid := range req.Jids {
		jids[i] = ffs.JobID(jid)
	}

	ch := make(chan ffs.Job, 100)
	go func() {
		err = i.WatchJobs(srv.Context(), ch, jids...)
		close(ch)
	}()
	for job := range ch {
		reply := &WatchJobsReply{
			Job: &Job{
				ID:       job.ID.String(),
				ApiID:    job.APIID.String(),
				Status:   JobStatus(job.Status),
				ErrCause: job.ErrCause,
			},
		}
		if err := srv.Send(reply); err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil
}

// WatchLogs returns a stream of human-readable messages related to executions of a Cid.
// The listener is automatically unsubscribed when the client closes the stream.
func (s *Service) WatchLogs(req *WatchLogsRequest, srv FFSAPI_WatchLogsServer) error {
	i, err := s.getInstanceByToken(srv.Context())
	if err != nil {
		return err
	}

	var opts []api.GetLogsOption
	if req.Jid != "" {
		opts = append(opts, api.WithJidFilter(ffs.JobID(req.Jid)))
	}

	c, err := cid.Decode(req.Cid)
	if err != nil {
		return err
	}
	ch := make(chan ffs.LogEntry, 100)
	go func() {
		err = i.WatchLogs(srv.Context(), ch, c, opts...)
		close(ch)
	}()
	for l := range ch {
		reply := &WatchLogsReply{
			LogEntry: &LogEntry{
				Cid:  c.String(),
				Jid:  l.Jid.String(),
				Time: l.Timestamp.Unix(),
				Msg:  l.Msg,
			},
		}
		if err := srv.Send(reply); err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	return nil
}

// PushConfig applies the provided cid config
func (s *Service) PushConfig(ctx context.Context, req *PushConfigRequest) (*PushConfigReply, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}

	c, err := cid.Decode(req.Cid)
	if err != nil {
		return nil, err
	}

	options := []api.PushConfigOption{}

	if req.HasConfig {
		cid, err := cid.Decode(req.Config.Cid)
		if err != nil {
			return nil, err
		}
		config := ffs.CidConfig{
			Cid: cid,
			Hot: ffs.HotConfig{
				Enabled:       req.Config.Hot.Enabled,
				AllowUnfreeze: req.Config.Hot.AllowUnfreeze,
				Ipfs: ffs.IpfsConfig{
					AddTimeout: int(req.Config.Hot.Ipfs.AddTimeout),
				},
			},
			Cold: ffs.ColdConfig{
				Enabled: req.Config.Cold.Enabled,
				Filecoin: ffs.FilConfig{
					RepFactor:      int(req.Config.Cold.Filecoin.RepFactor),
					DealDuration:   req.Config.Cold.Filecoin.DealDuration,
					ExcludedMiners: req.Config.Cold.Filecoin.ExcludedMiners,
					CountryCodes:   req.Config.Cold.Filecoin.CountryCodes,
					Renew: ffs.FilRenew{
						Enabled:   req.Config.Cold.Filecoin.Renew.Enabled,
						Threshold: int(req.Config.Cold.Filecoin.Renew.Threshold),
					},
				},
			},
		}
		options = append(options, api.WithCidConfig(config))
	}

	if req.HasOverrideConfig {
		options = append(options, api.WithOverride(req.OverrideConfig))
	}

	jid, err := i.PushConfig(c, options...)
	if err != nil {
		return nil, err
	}

	return &PushConfigReply{
		JobID: jid.String(),
	}, nil
}

// Get gets the data for a stored Cid.
func (s *Service) Get(req *GetRequest, srv FFSAPI_GetServer) error {
	i, err := s.getInstanceByToken(srv.Context())
	if err != nil {
		return err
	}
	c, err := cid.Decode(req.GetCid())
	if err != nil {
		return err
	}
	r, err := i.Get(srv.Context(), c)
	if err != nil {
		return err
	}

	buffer := make([]byte, 1024*32)
	for {
		bytesRead, err := r.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if sendErr := srv.Send(&GetReply{Chunk: buffer[:bytesRead]}); sendErr != nil {
			return sendErr
		}
		if err == io.EOF {
			return nil
		}
	}
}

// Close calls API.Close
func (s *Service) Close(ctx context.Context, req *CloseRequest) (*CloseReply, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	if err := i.Close(); err != nil {
		return nil, err
	}
	return &CloseReply{}, nil
}

// AddToHot stores data in the Hot Storage so the resulting cid can be used in PushConfig
func (s *Service) AddToHot(srv FFSAPI_AddToHotServer) error {
	// check that an API instance exists so not just anyone can add data to the hot layer
	if _, err := s.getInstanceByToken(srv.Context()); err != nil {
		return err
	}

	reader, writer := io.Pipe()
	defer func() {
		if err := reader.Close(); err != nil {
			log.Errorf("closing reader: %s", err)
		}
	}()

	go receiveFile(srv, writer)

	c, err := s.hot.Add(srv.Context(), reader)
	if err != nil {
		return fmt.Errorf("adding data to hot storage: %s", err)
	}

	return srv.SendAndClose(&AddToHotReply{Cid: c.String()})
}

func (s *Service) getInstanceByToken(ctx context.Context) (*api.API, error) {
	token := metautils.ExtractIncoming(ctx).Get("X-ffs-Token")
	if token == "" {
		return nil, ErrEmptyAuthToken
	}
	i, err := s.m.GetByAuthToken(token)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func receiveFile(srv FFSAPI_AddToHotServer, writer *io.PipeWriter) {
	for {
		req, err := srv.Recv()
		if err == io.EOF {
			_ = writer.Close()
			break
		} else if err != nil {
			_ = writer.CloseWithError(err)
			break
		}
		_, writeErr := writer.Write(req.GetChunk())
		if writeErr != nil {
			if err := writer.CloseWithError(writeErr); err != nil {
				log.Errorf("closing with error: %s", err)
			}
		}
	}
}
