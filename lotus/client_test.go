package lotus

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/filecoin/ldevnet"
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestClientVersion(t *testing.T) {
	dnet, err := ldevnet.New(t, 1)
	checkErr(t, err)
	defer dnet.Close()

	if _, err := dnet.Client.Version(context.Background()); err != nil {
		t.Fatalf("error when getting client version: %s", err)
	}
}

func TestClientImport(t *testing.T) {
	dnet, err := ldevnet.New(t, 1)
	checkErr(t, err)
	defer dnet.Close()

	// can't avoid home base path, ipfs checks: cannot add filestore references outside ipfs root (home folder)
	home := os.TempDir()
	f, err := ioutil.TempFile(home, "")
	checkErr(t, err)
	defer os.Remove(f.Name())
	defer f.Close()
	bts := make([]byte, 4)
	rand.Read(bts)
	io.Copy(f, bytes.NewReader(bts))

	cid, err := dnet.Client.ClientImport(context.Background(), f.Name())
	checkErr(t, err)
	if !cid.Defined() {
		t.Errorf("undefined cid from import")
	}
}

func TestClientChainNotify(t *testing.T) {
	dnet, err := ldevnet.New(t, 1)
	checkErr(t, err)

	ch, err := dnet.Client.ChainNotify(context.Background())
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
	dnet, err := ldevnet.New(t, 1)
	checkErr(t, err)

	ts, err := dnet.Client.ChainHead(context.Background())
	checkErr(t, err)
	if len(ts.Cids()) == 0 || len(ts.Blocks()) == 0 || ts.Height() == 0 {
		t.Fatalf("invalid tipset")
	}
}

func TestChainGetTipset(t *testing.T) {
	dnet, err := ldevnet.New(t, 1)
	checkErr(t, err)

	ts, err := dnet.Client.ChainHead(context.Background())
	checkErr(t, err)
	pts, err := dnet.Client.ChainGetTipSet(context.Background(), types.NewTipSetKey(ts.Blocks()[0].Parents...))
	checkErr(t, err)
	if len(pts.Cids()) == 0 || len(pts.Blocks()) == 0 || pts.Height() != ts.Height()-1 {
		t.Fatalf("invalid tipset")
	}
}

func TestStateReadState(t *testing.T) {
	dnet, err := ldevnet.New(t, 1)
	checkErr(t, err)

	addrs, err := dnet.Client.StateListMiners(context.Background(), nil)
	checkErr(t, err)

	for _, a := range addrs {
		actor, err := dnet.Client.StateGetActor(context.Background(), a, nil)
		checkErr(t, err)
		s, err := dnet.Client.StateReadState(context.Background(), actor, nil)
		checkErr(t, err)
		if s.State == nil {
			t.Fatalf("state of actor %s can't be nil", a)
		}
	}
}

func TestGetPeerID(t *testing.T) {
	dnet, err := ldevnet.New(t, 1)
	checkErr(t, err)

	miners, err := dnet.Client.StateListMiners(context.Background(), nil)
	checkErr(t, err)

	pid, err := dnet.Client.StateMinerPeerID(context.Background(), miners[0], nil)
	checkErr(t, err)
	checkErr(t, pid.Validate())
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
