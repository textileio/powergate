package retrieval

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	cbornode "github.com/ipfs/go-ipld-cbor"
	format "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log/v2"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
	itmanager "github.com/textileio/powergate/ffs/integrationtest/manager"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

type rootType struct {
	MyLinks []cid.Cid
}

func TestMain(m *testing.M) {
	cbornode.RegisterCborType(rootType{})
	util.AvgBlockTime = time.Millisecond * 500
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestPartialRetrievalFlow(t *testing.T) {
	t.SkipNow() // Skip since Lotus isn't ready for this.
	t.Parallel()
	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ctx := context.Background()
		ipfs, _, fapi, cls := itmanager.NewAPI(t, 1)
		defer cls()
		_ = ctx
		_ = fapi

		// Generate some data to run a selector.
		numInternalNodes := 3
		nodes := make([]format.Node, numInternalNodes)
		r := rand.New(rand.NewSource(22))
		for i := 0; i < numInternalNodes; i++ {
			buf := make([]byte, 500)
			_, _ = r.Read(buf)
			n, err := cbornode.WrapObject(buf, mh.SHA2_256, -1)
			require.NoError(t, err)
			nodes[i] = n
		}
		err := ipfs.Dag().AddMany(context.Background(), nodes)
		require.NoError(t, err)

		root := rootType{
			MyLinks: make([]cid.Cid, numInternalNodes),
		}
		for i, v := range nodes {
			root.MyLinks[i] = v.Cid()
		}
		rn, err := cbornode.WrapObject(root, mh.SHA2_256, -1)
		require.NoError(t, err)
		err = ipfs.Dag().Add(context.Background(), rn)
		require.NoError(t, err)

		/*
			c := rn.Cid() // Cid of data.

			// Make a deal with a IPLD graph that makes sense
			// to do partial retrieval.
			jid, err := fapi.PushStorageConfig(c)
			require.NoError(t, err)
			it.RequireJobState(t, fapi, jid, ffs.Success)

			// Current partial retrievals should be 0.
			prs, err := fapi.GetPartialRetrievals(c)
			require.NoError(t, err)
			require.Len(t, 0, prs)
			it.RequireIpfsUnpinnedCid(context.Background(), t, nodes[1].Cid(), ipfs)

			// Do partial retrieval.
			selector := "/Link/2/Hash/Qm...."
			jid, err = fapi.PushPartialRetrieval(c, selector)
			require.NoError(t, err)
			it.RequireJobState(t, fapi, jid, ffs.Success)
			it.RequireIpfsPinnedCid(context.Background(), t, nodes[1].Cid(), ipfs)

			// Current partial retrievals should be 1.
			prs, err = fapi.GetPartialRetrievals(c)
			require.NoError(t, err)
			require.Len(t, 1, prs)
			pr := prs[0]
			require.Equal(t, pr.RootCid, c)
			require.Equal(t, pr.Selector, selector)
			require.Equal(t, nodes[1].Cid(), pr.DataCid) // Change assertion to expected Cid.

			rr, err := fapi.Get(ctx, pr.DataCid)
			require.NoError(t, err)
			fetched, err := ioutil.ReadAll(rr)
			require.NoError(t, err)
			require.True(t, bytes.Equal(nodes[1].RawData(), fetched))

			// Remove it. Check that we have 0 partial retrievals again, and
			// check was unpined from IPFS node.
			err = fapi.RemovePartialRetrieval(c)
			require.NoError(t, err)
			prs, err = fapi.GetPartialRetrievals(c)
			require.NoError(t, err)
			require.Len(t, 0, prs)
			it.RequireIpfsUnpinnedCid(context.Background(), t, nodes[1].Cid(), ipfs)
		*/
	})
}
