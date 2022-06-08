package sjstore

import (
	"math/rand"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/tests"
)

type apiJobs struct {
	apiid     string
	queued    int
	executing int
	final     int
}

type createdJobs struct {
	apiid     string
	queued    []ffs.StorageJob
	executing []ffs.StorageJob
	final     []ffs.StorageJob
}
type test struct {
	name    string
	apiJobs []apiJobs
	cb      func(t *testing.T, s *Store, createdJobs ...createdJobs)
}

var testsList = []test{
	{
		name: "list all apis queued",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{Select: Queued})
			require.NoError(t, err)
			require.Len(t, res, 18)
		},
	},
	{
		name: "list all apis executing",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{Select: Executing})
			require.NoError(t, err)
			require.Len(t, res, 21)
		},
	},
	{
		name: "list all apis final",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{Select: Final})
			require.NoError(t, err)
			require.Len(t, res, 24)
		},
	},
	{
		name: "list all apis all",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{Select: All})
			require.NoError(t, err)
			require.Len(t, res, 63)
		},
	},
	{
		name: "list all apis empty",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{})
			require.NoError(t, err)
			require.Len(t, res, 63)
		},
	},
	{
		name: "list api1 queued",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{APIIDFilter: ffs.APIID("api1"), Select: Queued})
			require.NoError(t, err)
			require.Len(t, res, 3)
		},
	},
	{
		name: "list api1 executing",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{APIIDFilter: ffs.APIID("api1"), Select: Executing})
			require.NoError(t, err)
			require.Len(t, res, 4)
		},
	},
	{
		name: "list api1 final",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{APIIDFilter: ffs.APIID("api1"), Select: Final})
			require.NoError(t, err)
			require.Len(t, res, 5)
		},
	},
	{
		name: "list api1 all",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{APIIDFilter: ffs.APIID("api1"), Select: All})
			require.NoError(t, err)
			require.Len(t, res, 12)
		},
	},
	{
		name: "list api1 empty",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{APIIDFilter: ffs.APIID("api1")})
			require.NoError(t, err)
			require.Len(t, res, 12)
		},
	},
	{
		name: "list all apis with cid filter",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{CidFilter: createdJobs[0].queued[0].Cid})
			require.NoError(t, err)
			require.Len(t, res, 1)
			require.Equal(t, createdJobs[0].queued[0].Cid, res[0].Cid)
		},
	},
	{
		name: "list api1 with cid filter",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{APIIDFilter: ffs.APIID("api1"), CidFilter: createdJobs[0].queued[0].Cid})
			require.NoError(t, err)
			require.Len(t, res, 1)
			require.Equal(t, createdJobs[0].queued[0].Cid, res[0].Cid)
			require.Equal(t, createdJobs[0].queued[0].APIID, res[0].APIID)
		},
	},
	{
		name: "list wrong api filter with cid filter",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{APIIDFilter: ffs.APIID("api2"), CidFilter: createdJobs[0].queued[0].Cid})
			require.NoError(t, err)
			require.Len(t, res, 0)
		},
	},
	{
		name: "list nonexistent apiid",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{APIIDFilter: ffs.APIID("doesntexist")})
			require.NoError(t, err)
			require.Len(t, res, 0)
		},
	},
	{
		name: "list all apis descending",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{Ascending: false})
			require.NoError(t, err)
			require.Len(t, res, 63)
			requireOrder(t, res, false)
		},
	},
	{
		name: "list all apis ascending",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, _, _, err := s.List(ListConfig{Ascending: true})
			require.NoError(t, err)
			require.Len(t, res, 63)
			requireOrder(t, res, true)
		},
	},
	{
		name: "list all apis with limit",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, more, last, err := s.List(ListConfig{Limit: 5})
			require.NoError(t, err)
			require.Len(t, res, 5)
			require.True(t, more)
			require.Greater(t, len(last), 0)
		},
	},
	{
		name: "list api1 queued with limit",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			res, more, last, err := s.List(ListConfig{Limit: 5, APIIDFilter: ffs.APIID("api1"), Select: Queued})
			require.NoError(t, err)
			require.Len(t, res, 3)
			require.False(t, more)
			require.Empty(t, last)
		},
	},
	{
		name: "iterate list with limit and next token",
		apiJobs: []apiJobs{
			{apiid: "api1", queued: 3, executing: 4, final: 5},
			{apiid: "api2", queued: 6, executing: 7, final: 8},
			{apiid: "api3", queued: 9, executing: 10, final: 11},
		},
		cb: func(t *testing.T, s *Store, createdJobs ...createdJobs) {
			numPages := 0
			more := true
			nextToken := ""
			for more {
				c := ListConfig{Limit: 10}
				if nextToken != "" {
					c.NextPageToken = nextToken
				}
				res, m, n, err := s.List(c)
				require.NoError(t, err)
				require.Greater(t, len(res), 0)
				numPages++
				nextToken = n
				more = m
			}
			require.Equal(t, 7, numPages)
		},
	},
}

func TestTests(t *testing.T) {
	for _, test := range testsList {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			s := create(t)
			var allCreatedJobs []createdJobs
			for _, apiJobs := range test.apiJobs {
				cj := createdJobs{apiid: apiJobs.apiid}
				for i := 0; i < apiJobs.queued; i++ {
					j := createJob(t, apiJobs.apiid, cid.Undef)
					err := s.Enqueue(j)
					require.NoError(t, err)
					cj.queued = append(cj.queued, j)
				}
				for i := 0; i < apiJobs.executing; i++ {
					j := createJob(t, apiJobs.apiid, cid.Undef)
					err := s.Enqueue(j)
					require.NoError(t, err)
					dj, err := s.Dequeue(ffs.APIID(apiJobs.apiid))
					require.NoError(t, err)
					cj.executing = append(cj.executing, *dj)
				}
				for i := 0; i < apiJobs.final; i++ {
					j := createJob(t, apiJobs.apiid, cid.Undef)
					err := s.Enqueue(j)
					require.NoError(t, err)
					err = s.Finalize(j.ID, ffs.Success, nil, nil)
					require.NoError(t, err)
					cj.final = append(cj.final, j)
				}
				allCreatedJobs = append(allCreatedJobs, cj)
			}
			test.cb(t, s, allCreatedJobs...)
		})
	}
}

func TestEnqueue(t *testing.T) {
	t.Parallel()
	s := create(t)
	j := createJob(t, "apiid", cid.Undef)
	err := s.Enqueue(j)
	require.NoError(t, err)
	jQueued, err := s.Get(j.ID)
	require.NoError(t, err)
	j.Status = ffs.Queued
	require.Equal(t, j, jQueued)
}

func TestDequeue(t *testing.T) {
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		s := create(t)
		j := createJob(t, "apiid", cid.Undef)
		err := s.Enqueue(j)
		require.NoError(t, err)
		j2, err := s.Dequeue(ffs.EmptyInstanceID)
		require.NoError(t, err)
		j.Status = ffs.Executing
		require.NotNil(t, j2)
		require.Equal(t, j, *j2)
	})
	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		s := create(t)
		j, err := s.Dequeue(ffs.EmptyInstanceID)
		require.NoError(t, err)
		require.Nil(t, j)
	})
	t.Run("ExecutingAndFinalized", func(t *testing.T) {
		t.Parallel()
		s := create(t)

		j1 := createJob(t, "apiid", cid.Undef)
		err := s.Enqueue(j1)
		require.NoError(t, err)
		_, err = s.Dequeue(ffs.EmptyInstanceID)
		require.NoError(t, err)

		j2 := createJob(t, "apiid", j1.Cid)
		err = s.Enqueue(j2)
		require.NoError(t, err)

		// Dequeue should be nil since j1 is Executing,
		// and j2 can't be returned until that fininishes since
		// both are jobs for the same Cid.
		jq, err := s.Dequeue(ffs.EmptyInstanceID)
		require.NoError(t, err)
		require.Nil(t, jq)

		errors := []ffs.DealError{{ProposalCid: cid.Undef, Miner: "t0100", Message: "abcd"}}
		err = s.Finalize(j1.ID, ffs.Success, nil, errors)
		require.NoError(t, err)
		// Dequeue now returns a new job since the Executing one
		// was finalized
		jq, err = s.Dequeue(ffs.EmptyInstanceID)
		require.NoError(t, err)
		require.NotNil(t, jq)
		j2.Status = ffs.Executing
		require.Equal(t, j2, *jq)
	})
	t.Run("APIID", func(t *testing.T) {
		t.Parallel()
		s := create(t)

		j1 := createJob(t, "apiid1", cid.Undef)
		err := s.Enqueue(j1)
		require.NoError(t, err)

		j2 := createJob(t, "apiid2", cid.Undef)
		err = s.Enqueue(j2)
		require.NoError(t, err)

		j, err := s.Dequeue(ffs.APIID("apiid2"))
		require.NoError(t, err)
		require.Equal(t, ffs.APIID("apiid2"), j.APIID)
	})
}

func TestCancelation(t *testing.T) {
	t.Parallel()
	s := create(t)

	j1 := createJob(t, "apiid", cid.Undef)
	err := s.Enqueue(j1)
	require.NoError(t, err)

	j2 := createJob(t, "apiid", j1.Cid)
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

func requireOrder(t *testing.T, res []ffs.StorageJob, ascending bool) {
	var last *ffs.StorageJob
	for _, job := range res {
		if last != nil && !ascending {
			require.Less(t, job.CreatedAt, last.CreatedAt)
		}
		if last != nil && ascending {
			require.Greater(t, job.CreatedAt, last.CreatedAt)
		}
		tmp := &job
		bar := *tmp
		last = &bar
	}
}

func createJob(_ *testing.T, apiid string, c cid.Cid) ffs.StorageJob {
	finalCid := c
	if !finalCid.Defined() {
		data := make([]byte, 20)
		rand.Read(data)
		hash, _ := mh.Sum(data, mh.SHA2_256, -1)
		finalCid = cid.NewCidV1(cid.DagCBOR, hash)
	}

	return ffs.StorageJob{
		ID:        ffs.NewJobID(),
		APIID:     ffs.APIID(apiid),
		Cid:       finalCid,
		CreatedAt: time.Now().UnixNano(),
	}
}

func create(t *testing.T) *Store {
	ds := tests.NewTxMapDatastore()
	store, err := New(ds, tests.NewMockNotifier())
	require.NoError(t, err)
	return store
}
