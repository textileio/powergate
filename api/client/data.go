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
	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
	"github.com/textileio/powergate/v2/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Data provides access to Powergate general data APIs.
type Data struct {
	client userPb.UserServiceClient
}

// WatchLogsOption is a function that changes GetLogsConfig.
type WatchLogsOption func(r *userPb.WatchLogsRequest)

// WithJobIDFilter filters only log messages of a Cid related to
// the Job with id jid.
func WithJobIDFilter(jobID string) WatchLogsOption {
	return func(r *userPb.WatchLogsRequest) {
		r.JobId = jobID
	}
}

// WithHistory indicates that prior history logs should
// be sent in the channel before getting real time logs.
func WithHistory(enabled bool) WatchLogsOption {
	return func(r *userPb.WatchLogsRequest) {
		r.History = enabled
	}
}

// WatchLogsEvent represents an event for watching cid logs.
type WatchLogsEvent struct {
	Res *userPb.WatchLogsResponse
	Err error
}

// Stage allows to temporarily stage data in hot storage in preparation for pushing a cid storage config.
func (d *Data) Stage(ctx context.Context, data io.Reader) (*userPb.StageResponse, error) {
	stream, err := d.client.Stage(ctx)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024*32) // 32KB
	for {
		bytesRead, err := data.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		sendErr := stream.Send(&userPb.StageRequest{Chunk: buffer[:bytesRead]})
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

// StageFolder allows to temporarily stage a folder in hot storage in preparation for pushing a cid storage config.
func (d *Data) StageFolder(ctx context.Context, ipfsRevProxyAddr string, folderPath string) (string, error) {
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
		options.Unixfs.Pin(true),
	}
	pth, err := ipfs.Unixfs().Add(context.Background(), files.ToDir(ff), opts...)
	if err != nil {
		return "", err
	}

	_, err = d.client.StageCid(ctx, &userPb.StageCidRequest{Cid: pth.Cid().String()})
	if err != nil {
		return "", fmt.Errorf("stage pinning cid: %s", err)
	}

	return pth.Cid().String(), nil
}

// ReplaceData pushes a StorageConfig for c2 equal to that of c1, and removes c1. This operation
// is more efficient than manually removing and adding in two separate operations.
func (d *Data) ReplaceData(ctx context.Context, cid1, cid2 string) (*userPb.ReplaceDataResponse, error) {
	return d.client.ReplaceData(ctx, &userPb.ReplaceDataRequest{Cid1: cid1, Cid2: cid2})
}

// Get returns an io.Reader for reading a stored Cid from hot storage.
func (d *Data) Get(ctx context.Context, cid string) (io.Reader, error) {
	stream, err := d.client.Get(ctx, &userPb.GetRequest{
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
func (d *Data) GetFolder(ctx context.Context, ipfsRevProxyAddr, cid, outputDir string) error {
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

// WatchLogs pushes human-friendly messages about Cid executions. The method is blocking
// and will continue to send messages until the context is canceled. The provided channel
// is owned by the method and must not be closed.
func (d *Data) WatchLogs(ctx context.Context, ch chan<- WatchLogsEvent, cid string, opts ...WatchLogsOption) error {
	r := &userPb.WatchLogsRequest{Cid: cid}
	for _, opt := range opts {
		opt(r)
	}
	stream, err := d.client.WatchLogs(ctx, r)
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

// CidSummary gives a summary of the storage and jobs state of the specified cid.
func (d *Data) CidSummary(ctx context.Context, cids ...string) (*userPb.CidSummaryResponse, error) {
	return d.client.CidSummary(ctx, &userPb.CidSummaryRequest{Cids: cids})
}

// CidInfo returns information about a cid stored by the user.
func (d *Data) CidInfo(ctx context.Context, cid string) (*userPb.CidInfoResponse, error) {
	return d.client.CidInfo(ctx, &userPb.CidInfoRequest{Cid: cid})
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
