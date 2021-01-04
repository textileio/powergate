package sjstore

import (
	"math/rand"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

type apiJobs struct {
	apiid     string
	queued    int
	executing int
	final     int
}
type test struct {
	name    string
	apiJobs []apiJobs
	cb      func(t *testing.T, s *Store)
}

var testsx = []test{
	{
		name: "enqueue",
		apiJobs: []apiJobs{
			{
				apiid:  "api1",
				queued: 1,
			},
		},
		cb: func(t *testing.T, s *Store) {
			res, _, _, err := s.List(ListConfig{Select: Queued})
			require.NoError(t, err)
			require.Len(t, res, 1)
		},
	},
	{
		name: "final",
		apiJobs: []apiJobs{
			{
				apiid: "api1",
				final: 3,
			},
		},
		cb: func(t *testing.T, s *Store) {
			res, _, _, err := s.List(ListConfig{Select: Final})
			require.NoError(t, err)
			require.Len(t, res, 3)
		},
	},
}

func TestTests(t *testing.T) {
	for _, test := range testsx {
		t.Logf("running test: %s", test.name)
		s := create(t)
		for _, apiJobs := range test.apiJobs {
			total := apiJobs.queued + apiJobs.executing + apiJobs.final
			numToExecute := apiJobs.executing + apiJobs.final
			for i := 0; i < total; i++ {
				j := createJob(t, apiJobs.apiid)
				err := s.Enqueue(j)
				require.NoError(t, err)
			}
			var executing []*ffs.StorageJob
			for i := 0; i < numToExecute; i++ {
				j, err := s.Dequeue()
				require.NoError(t, err)
				executing = append(executing, j)
			}
			for i := 0; i < apiJobs.final; i++ {
				err := s.Finalize(executing[i].ID, ffs.Success, nil, nil)
				require.NoError(t, err)
			}
		}
		test.cb(t, s)
	}
}

func TestEnqueue(t *testing.T) {
	t.Parallel()
	s := create(t)
	j := createJob(t, "apiid")
	err := s.Enqueue(j)
	require.NoError(t, err)
	jQueued, err := s.Get(j.ID)
	require.NoError(t, err)
	j.Status = ffs.Queued
	require.Equal(t, j, jQueued)
}

func TestFoo(t *testing.T) {
	t.Parallel()
	s := create(t)
	j1 := createJob(t, "api1")
	j2 := createJob(t, "api1")
	j3 := createJob(t, "api1")
	j4 := createJob(t, "api1")
	j5 := createJob(t, "api1")
	j6 := createJob(t, "api2")
	j7 := createJob(t, "api2")
	j8 := createJob(t, "api2")
	j9 := createJob(t, "api2")
	err := s.Enqueue(j1)
	require.NoError(t, err)
	err = s.Enqueue(j2)
	require.NoError(t, err)
	err = s.Enqueue(j3)
	require.NoError(t, err)
	err = s.Enqueue(j4)
	require.NoError(t, err)
	err = s.Enqueue(j5)
	require.NoError(t, err)
	err = s.Enqueue(j6)
	require.NoError(t, err)
	err = s.Enqueue(j7)
	require.NoError(t, err)
	err = s.Enqueue(j8)
	require.NoError(t, err)
	err = s.Enqueue(j9)
	require.NoError(t, err)
	jQueued, err := s.Get(j1.ID)
	require.NoError(t, err)
	j1.Status = ffs.Queued
	require.Equal(t, j1, jQueued)

	res, more, last, err := s.List(ListConfig{Limit: 1})
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.True(t, more)
	requireDescending(t, res)

	res, more, last, err = s.List(ListConfig{NextPageToken: last, Limit: 1})
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.True(t, more)
	requireDescending(t, res)

	res, more, last, err = s.List(ListConfig{NextPageToken: last, Limit: 4})
	require.NoError(t, err)
	require.Len(t, res, 4)
	require.True(t, more)
	requireDescending(t, res)

	res, more, last, err = s.List(ListConfig{NextPageToken: last, Limit: 10})
	require.NoError(t, err)
	require.Len(t, res, 3)
	require.False(t, more)
	require.NotNil(t, last)
	requireDescending(t, res)

	res, _, _, err = s.List(ListConfig{APIIDFilter: "doesntexist"})
	require.NoError(t, err)
	require.Empty(t, res)

	res, _, _, err = s.List(ListConfig{APIIDFilter: "api1"})
	require.NoError(t, err)
	require.Len(t, res, 5)

	res, _, _, err = s.List(ListConfig{APIIDFilter: "api2"})
	require.NoError(t, err)
	require.Len(t, res, 4)

	c, err := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u")
	require.NoError(t, err)
	res, _, _, err = s.List(ListConfig{CidFilter: c})
	require.NoError(t, err)
	require.Len(t, res, 1)

	res, _, _, err = s.List(ListConfig{CidFilter: c, APIIDFilter: "api1"})
	require.NoError(t, err)
	require.Len(t, res, 1)

	res, _, _, err = s.List(ListConfig{CidFilter: c, APIIDFilter: "api2"})
	require.NoError(t, err)
	require.Empty(t, res)

	res, _, _, err = s.List(ListConfig{Select: Final})
	require.NoError(t, err)
	require.Len(t, res, 0)
}

func requireDescending(t *testing.T, res []ffs.StorageJob) {
	var last *ffs.StorageJob
	for _, job := range res {
		if last != nil {
			require.Less(t, job.CreatedAt, last.CreatedAt)
		}
		foo := &job
		bar := *foo
		last = &bar
	}
}

func TestDequeue(t *testing.T) {
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		s := create(t)
		j := createJob(t, "apiid")
		err := s.Enqueue(j)
		require.NoError(t, err)
		j2, err := s.Dequeue()
		require.NoError(t, err)
		j.Status = ffs.Executing
		require.NotNil(t, j2)
		require.Equal(t, j, *j2)
	})
	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		s := create(t)
		j, err := s.Dequeue()
		require.NoError(t, err)
		require.Nil(t, j)
	})
	t.Run("ExecutingAndFinalized", func(t *testing.T) {
		t.Parallel()
		s := create(t)

		j1 := createJob(t, "apiid")
		err := s.Enqueue(j1)
		require.NoError(t, err)
		_, err = s.Dequeue()
		require.NoError(t, err)

		j2 := createJob(t, "apiid")
		err = s.Enqueue(j2)
		require.NoError(t, err)

		// Dequeue should be nil since j1 is Executing,
		// and j2 can't be returned until that fininishes since
		// both are jobs for the same Cid.
		jq, err := s.Dequeue()
		require.NoError(t, err)
		require.Nil(t, jq)

		errors := []ffs.DealError{{ProposalCid: cid.Undef, Miner: "t0100", Message: "abcd"}}
		err = s.Finalize(j1.ID, ffs.Success, nil, errors)
		require.NoError(t, err)
		// Dequeue now returns a new job since the Executing one
		// was finalized
		jq, err = s.Dequeue()
		require.NoError(t, err)
		require.NotNil(t, jq)
		j2.Status = ffs.Executing
		require.Equal(t, j2, *jq)
	})
}

func TestCancelation(t *testing.T) {
	t.Parallel()
	s := create(t)

	j1 := createJob(t, "apiid")
	err := s.Enqueue(j1)
	require.NoError(t, err)

	j2 := createJob(t, "apiid")
	err = s.Enqueue(j2)
	require.NoError(t, err)

	j1Canceled, err := s.Get(j1.ID)
	require.NoError(t, err)
	require.Equal(t, ffs.Canceled, j1Canceled.Status)

	j2Queued, err := s.Get(j2.ID)
	require.NoError(t, err)
	require.Equal(t, ffs.Queued, j2Queued.Status)
}

func TestStartedDeals(t *testing.T) {
	t.Parallel()
	s := create(t)

	iid := ffs.NewAPIID()
	b, _ := multihash.Encode([]byte("data"), multihash.SHA1)
	cidData := cid.NewCidV1(1, b)

	fds, err := s.GetStartedDeals(iid, cidData)
	require.NoError(t, err)
	require.Equal(t, 0, len(fds))

	b, _ = multihash.Encode([]byte("prop1"), multihash.SHA1)
	cidProp1 := cid.NewCidV1(1, b)
	b, _ = multihash.Encode([]byte("prop2"), multihash.SHA1)
	cidProp2 := cid.NewCidV1(1, b)

	startedDeals := []cid.Cid{cidProp1, cidProp2}
	err = s.AddStartedDeals(iid, cidData, startedDeals)
	require.NoError(t, err)

	fds, err = s.GetStartedDeals(iid, cidData)
	require.NoError(t, err)
	require.Equal(t, startedDeals, fds)

	err = s.RemoveStartedDeals(iid, cidData)
	require.NoError(t, err)

	fds, err = s.GetStartedDeals(iid, cidData)
	require.NoError(t, err)
	require.Equal(t, 0, len(fds))
}

func createJob(_ *testing.T, apiid string) ffs.StorageJob {
	data := make([]byte, 20)
	rand.Read(data)
	hash, _ := mh.Sum(data, mh.SHA2_256, -1)
	c := cid.NewCidV1(cid.DagCBOR, hash)

	return ffs.StorageJob{
		ID:        ffs.NewJobID(),
		APIID:     ffs.APIID(apiid),
		Cid:       c,
		CreatedAt: time.Now().UnixNano(),
	}
}

func create(t *testing.T) *Store {
	ds := tests.NewTxMapDatastore()
	store, err := New(ds)
	require.NoError(t, err)
	return store
}
