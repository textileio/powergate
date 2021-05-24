package integrationtest

import (
	"bytes"
	"context"
	"fmt"
	lapi "github.com/filecoin-project/lotus/api"
	"math/rand"
	"time"

	"github.com/ipfs/go-cid"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/ffs/api"
	"github.com/textileio/powergate/v2/tests"
	"github.com/textileio/powergate/v2/util"
)

// CreateIPFS creates a docker container running IPFS.
func CreateIPFS(t tests.TestingTWithCleanup) (*httpapi.HttpApi, string) {
	ipfsDocker, cls := tests.LaunchIPFSDocker(t)
	t.Cleanup(cls)
	ipfsAddr := util.MustParseAddr("/ip4/127.0.0.1/tcp/" + ipfsDocker.GetPort("5001/tcp"))
	ipfs, err := httpapi.NewApi(ipfsAddr)
	require.NoError(t, err)
	bridgeIP := ipfsDocker.Container.NetworkSettings.Networks["bridge"].IPAddress
	ipfsDockerMAddr := fmt.Sprintf("/ip4/%s/tcp/5001", bridgeIP)

	return ipfs, ipfsDockerMAddr
}

// RequireIpfsUnpinnedCid checks that a cid is unpinned in the IPFS node.
func RequireIpfsUnpinnedCid(ctx context.Context, t require.TestingT, cid cid.Cid, ipfsAPI *httpapi.HttpApi) {
	pins, err := ipfsAPI.Pin().Ls(ctx, options.Pin.Ls.Recursive())
	require.NoError(t, err)
	for p := range pins {
		require.NotEqual(t, cid, p.Path().Cid(), "Cid isn't unpinned from IPFS node")
	}
}

// RequireIpfsPinnedCid checks that a cid is pinned in the IPFS node.
func RequireIpfsPinnedCid(ctx context.Context, t require.TestingT, cid cid.Cid, ipfsAPI *httpapi.HttpApi) {
	pins, err := ipfsAPI.Pin().Ls(ctx)
	require.NoError(t, err)

	pinned := false
	for p := range pins {
		if p.Path().Cid() == cid {
			pinned = true
			break
		}
	}
	require.True(t, pinned, "Cid should be pinned in IPFS node")
}

// RequireFilUnstored checks that a cid is not stored in the Filecoin network.
func RequireFilUnstored(ctx context.Context, t require.TestingT, client *lapi.FullNodeStruct, c cid.Cid) {
	offers, err := client.ClientFindData(ctx, c, nil)
	require.NoError(t, err)
	require.Empty(t, offers)
}

// RequireFilStored cehcks that a cid is stored in the Filecoin network.
func RequireFilStored(ctx context.Context, t require.TestingT, client *lapi.FullNodeStruct, c cid.Cid) {
	offers, err := client.ClientFindData(ctx, c, nil)
	require.NoError(t, err)
	require.NotEmpty(t, offers)
}

// RequireStorageJobState checks if the current status of a job matches one of the specified statuses.
func RequireStorageJobState(t require.TestingT, fapi *api.API, jid ffs.JobID, statuses ...ffs.JobStatus) ffs.StorageJob {
	job, err := fapi.GetStorageJob(jid)
	require.NoError(t, err)
	require.Contains(t, statuses, job.Status)
	return job
}

// RequireEventualJobState watches a Job for a desired status.
func RequireEventualJobState(t require.TestingT, fapi *api.API, jid ffs.JobID, status ffs.JobStatus) ffs.StorageJob {
	ch := make(chan ffs.StorageJob, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	go func() {
		err = fapi.WatchJobs(ctx, ch, jid)
		close(ch)
	}()
	stop := false
	var res ffs.StorageJob
	for !stop {
		select {
		case <-time.After(120 * time.Second):
			t.Errorf("waiting for job update timeout")
			t.FailNow()
		case job, ok := <-ch:
			require.True(t, ok)
			require.Equal(t, jid, job.ID)
			if job.Status == ffs.Queued || job.Status == ffs.Executing {
				if job.Status == status {
					stop = true
					res = job
				}
				continue
			}
			require.Equal(t, status, job.Status, job.ErrCause)
			stop = true
			res = job
		}
	}
	require.NoError(t, err)
	return res
}

// RequireStorageConfig compares a cid storage config against a target.
func RequireStorageConfig(t require.TestingT, fapi *api.API, c cid.Cid, config *ffs.StorageConfig) {
	if config == nil {
		defConfig := fapi.DefaultStorageConfig()
		config = &defConfig
	}
	currentConfigs, err := fapi.GetStorageConfigs(c)
	require.NoError(t, err)
	require.Equal(t, *config, currentConfigs[c])
}

// RequireStorageDealRecord checks that a storage deal record exist for a cid.
func RequireStorageDealRecord(t require.TestingT, fapi *api.API, c cid.Cid) {
	time.Sleep(time.Second)
	recs, err := fapi.StorageDealRecords(deals.WithIncludeFinal(true))
	require.NoError(t, err)
	require.Len(t, recs, 1)
	require.Equal(t, c, recs[0].RootCid)
}

// RequireRetrievalDealRecord checks that a retrieval deal record exits for a cid.
func RequireRetrievalDealRecord(t require.TestingT, fapi *api.API, c cid.Cid) {
	recs, err := fapi.RetrievalDealRecords()
	require.NoError(t, err)
	require.Len(t, recs, 1)
	require.Equal(t, c, recs[0].DealInfo.RootCid)
}

// RandomBytes returns a slice of random bytes of a desired size.
func RandomBytes(r *rand.Rand, size int) []byte {
	buf := make([]byte, size)
	_, _ = r.Read(buf)
	return buf
}

// AddRandomFile adds a random file to the IPFS node.
func AddRandomFile(t require.TestingT, r *rand.Rand, ipfs *httpapi.HttpApi) (cid.Cid, []byte) {
	return AddRandomFileSize(t, r, ipfs, 1600)
}

// AddRandomFileSize adds a random file with a specified size to the IPFS node.
func AddRandomFileSize(t require.TestingT, r *rand.Rand, ipfs *httpapi.HttpApi, size int) (cid.Cid, []byte) {
	data := RandomBytes(r, size)
	node, err := ipfs.Unixfs().Add(context.Background(), ipfsfiles.NewReaderFile(bytes.NewReader(data)), options.Unixfs.Pin(false))
	if err != nil {
		t.Errorf("error adding random file: %s", err)
		t.FailNow()
	}
	return node.Cid(), data
}
