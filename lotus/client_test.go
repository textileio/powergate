package lotus_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/tests"
)

const (
	tmpDir = "/tmp/powergate/lotusclienttest"
)

func TestMain(m *testing.M) {
	if err := os.RemoveAll(tmpDir); err != nil {
		panic(err)
	}
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
			panic("can't create temp dir")
		}
	}
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestClientImport(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)

	f, err := ioutil.TempFile(tmpDir, "")
	checkErr(t, err)
	defer func() {
		require.NoError(t, f.Close())
		require.NoError(t, os.Remove(f.Name()))
	}()
	bts := make([]byte, 4)
	_, err = rand.Read(bts)
	checkErr(t, err)
	_, err = io.Copy(f, bytes.NewReader(bts))
	checkErr(t, err)

	ref := api.FileRef{
		Path: f.Name(),
	}
	res, err := client.ClientImport(context.Background(), ref)
	checkErr(t, err)
	if !res.Root.Defined() {
		t.Errorf("undefined cid from import")
	}
}

func TestClientChainNotify(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)

	ch, err := client.ChainNotify(context.Background())
	checkErr(t, err)

	// ch is guaranteed to push always current tipset
	h := <-ch
	if len(h) != 1 {
		t.Fatalf("first pushed notification should have length 1")
	}
	if h[0].Type != "current" || len(h[0].Val.Cids()) == 0 || h[0].Val.Height() == 0 {
		t.Fatalf("current head has invalid values")
	}

	select {
	case <-time.After(time.Second * 10):
		t.Fatalf("a new block should be received in less than ~10s")
	case <-ch:
		return
	}
}

func TestChainHead(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)

	ts, err := client.ChainHead(context.Background())
	checkErr(t, err)
	if len(ts.Cids()) == 0 || len(ts.Blocks()) == 0 || ts.Height() == 0 {
		t.Fatalf("invalid tipset")
	}
}

func TestChainGetTipset(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)

	ts, err := client.ChainHead(context.Background())
	checkErr(t, err)
	pts, err := client.ChainGetTipSet(context.Background(), types.NewTipSetKey(ts.Blocks()[0].Parents...))
	checkErr(t, err)
	if len(pts.Cids()) == 0 || len(pts.Blocks()) == 0 || pts.Height() != ts.Height()-1 {
		t.Fatalf("invalid tipset")
	}
}

func TestGetPeerID(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)

	miners, err := client.StateListMiners(context.Background(), types.EmptyTSK)
	checkErr(t, err)

	mi, err := client.StateMinerInfo(context.Background(), miners[0], types.EmptyTSK)
	checkErr(t, err)
	checkErr(t, mi.PeerId.Validate())
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
