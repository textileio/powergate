package grpc

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
	pb "github.com/textileio/powergate/ffs/pb"
)

var (
	// ErrEmptyAuthToken is returned when the provided auth-token is unkown.
	ErrEmptyAuthToken = errors.New("auth token can't be empty")

	log = logger.Logger("ffs-grpc-service")
)

// Service implements the proto service definition of FFS.
type Service struct {
	pb.UnimplementedAPIServer

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

// Info returns an Api information.
func (s *Service) Info(ctx context.Context, req *pb.InfoRequest) (*pb.InfoReply, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}

	info, err := i.Info(ctx)
	if err != nil {
		return nil, err
	}

	reply := &pb.InfoReply{
		Id:   info.ID.String(),
		Pins: make([]string, len(info.Pins)),
		Wallet: &pb.WalletInfo{
			Address: info.Wallet.Address,
			Balance: info.Wallet.Balance,
		},
	}
	for i, p := range info.Pins {
		reply.Pins[i] = p.String()
	}

	return reply, nil
}

// Show returns information about a particular Cid.
func (s *Service) Show(ctx context.Context, req *pb.ShowRequest) (*pb.ShowReply, error) {
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
	reply := &pb.ShowReply{
		Cid:     info.Cid.String(),
		Created: info.Created.UnixNano(),
		Hot: &pb.ShowReply_HotInfo{
			Size: int64(info.Hot.Size),
			Ipfs: &pb.ShowReply_IpfsHotInfo{
				Created: info.Hot.Ipfs.Created.UnixNano(),
			},
		},
		Cold: &pb.ShowReply_ColdInfo{
			Filecoin: &pb.ShowReply_FilInfo{
				DataCid:   info.Cold.Filecoin.DataCid.String(),
				Proposals: make([]*pb.ShowReply_FilStorage, len(info.Cold.Filecoin.Proposals)),
			},
		},
	}
	for i, p := range info.Cold.Filecoin.Proposals {
		reply.Cold.Filecoin.Proposals[i] = &pb.ShowReply_FilStorage{
			ProposalCid:     p.ProposalCid.String(),
			Active:          p.Active,
			Duration:        p.Duration,
			ActivationEpoch: int64(p.ActivationEpoch),
		}
	}

	return reply, nil
}

// AddCid adds a cid to an Api.
func (s *Service) AddCid(ctx context.Context, req *pb.AddCidRequest) (*pb.AddCidReply, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	c, err := cid.Decode(req.GetCid())
	if err != nil {
		return nil, err
	}
	log.Infof("adding cid %s", c)
	jid, err := i.PushConfig(c)
	if err != nil {
		return nil, err
	}

	ch, err := i.Watch(jid)
	if err != nil {
		return nil, fmt.Errorf("watching add cid created job: %s", err)
	}
	defer i.Unwatch(ch)
	for job := range ch {
		if job.Status == ffs.Success {
			break
		} else if job.Status == ffs.Cancelled || job.Status == ffs.Failed {
			return nil, fmt.Errorf("error adding cid: %s", job.ErrCause)
		}
	}
	return &pb.AddCidReply{}, nil
}

func receiveFile(srv pb.API_AddFileServer, writer *io.PipeWriter) {
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
			writer.CloseWithError(writeErr)
		}
	}
}

// AddFile stores data in the Hot Storage and saves it in an Api.
func (s *Service) AddFile(srv pb.API_AddFileServer) error {
	i, err := s.getInstanceByToken(srv.Context())
	if err != nil {
		return err
	}

	reader, writer := io.Pipe()
	defer reader.Close()

	go receiveFile(srv, writer)

	c, err := s.hot.Add(srv.Context(), reader)
	if err != nil {
		return fmt.Errorf("adding data to hot storage: %s", err)
	}

	jid, err := i.PushConfig(c)
	if err != nil {
		return err
	}
	ch, err := i.Watch(jid)
	if err != nil {
		return fmt.Errorf("watching add file created job: %s", err)
	}
	defer i.Unwatch(ch)
	for job := range ch {
		if job.Status == ffs.Success {
			break
		} else if job.Status == ffs.Failed {
			return fmt.Errorf("error adding cid: %s", job.ErrCause)
		}
	}

	return srv.SendAndClose(&pb.AddFileReply{Cid: c.String()})
}

// Get gets the data for a stored Cid.
func (s *Service) Get(req *pb.GetRequest, srv pb.API_GetServer) error {
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
		if sendErr := srv.Send(&pb.GetReply{Chunk: buffer[:bytesRead]}); sendErr != nil {
			return sendErr
		}
		if err == io.EOF {
			return nil
		}
	}
}

// Create creates a new Api.
func (s *Service) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateReply, error) {
	id, addr, err := s.m.Create(ctx)
	if err != nil {
		log.Errorf("creating instance: %s", err)
		return nil, err
	}
	return &pb.CreateReply{
		Id:      id.String(),
		Address: addr,
	}, nil
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
