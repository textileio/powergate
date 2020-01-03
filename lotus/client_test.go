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

	"github.com/textileio/filecoin/lotus/types"
	"github.com/textileio/filecoin/tests"
)

func TestClientVersion(t *testing.T) {
	addr, token := tests.ClientConfig()
	c, cls, err := New(addr, token)
	checkErr(t, err)
	defer cls()
	if _, err := c.Version(context.Background()); err != nil {
		t.Fatalf("error when getting client version: %s", err)
	}
}

func TestClientImport(t *testing.T) {
	// can't avoid home base path, ipfs checks: cannot add filestore references outside ipfs root (home folder)
	home, err := os.UserHomeDir()
	checkErr(t, err)
	f, err := ioutil.TempFile(home, "")
	checkErr(t, err)
	defer os.Remove(f.Name())
	defer f.Close()
	bts := make([]byte, 4)
	rand.Read(bts)
	io.Copy(f, bytes.NewReader(bts))

	addr, token := tests.ClientConfig()
	c, cls, err := New(addr, token)
	checkErr(t, err)
	defer cls()

	cid, err := c.ClientImport(context.Background(), f.Name())
	checkErr(t, err)
	if !cid.Defined() {
		t.Errorf("undefined cid from import")
	}
}

func TestClientChainNotify(t *testing.T) {
	addr, token := tests.ClientConfig()
	c, cls, err := New(addr, token)
	checkErr(t, err)
	defer cls()

	ch, err := c.ChainNotify(context.Background())
	checkErr(t, err)

	// ch is guaranteed to push always current tipset
	h := <-ch
	if len(h) != 1 {
		t.Fatalf("first pushed notification should have length 1")
	}
	if h[0].Type != "current" || len(h[0].Val.Cids) == 0 || h[0].Val.Height == 0 {
		t.Fatalf("current head has invalid values")
	}

	select {
	case <-time.After(time.Second * 50):
		t.Fatalf("a new block should be received in less than ~45s")
	case <-ch:
		return
	}
}

func TestSyncState(t *testing.T) {
	addr, token := tests.ClientConfig()
	c, cls, err := New(addr, token)
	checkErr(t, err)
	defer cls()

	state, err := c.SyncState(context.Background())
	checkErr(t, err)
	if state.ActiveSyncs[0].Height == 0 {
		t.Fatalf("current height can't be zero")
	}
}

func TestChainHead(t *testing.T) {
	addr, token := tests.ClientConfig()
	c, cls, err := New(addr, token)
	checkErr(t, err)
	defer cls()

	ts, err := c.ChainHead(context.Background())
	checkErr(t, err)
	if len(ts.Cids) == 0 || len(ts.Blocks) == 0 || ts.Height == 0 {
		t.Fatalf("invalid tipset")
	}
}

func TestChainGetTipset(t *testing.T) {
	addr, token := tests.ClientConfig()
	c, cls, err := New(addr, token)
	checkErr(t, err)
	defer cls()

	ts, err := c.ChainHead(context.Background())
	checkErr(t, err)
	pts, err := c.ChainGetTipSet(context.Background(), types.NewTipSetKey(ts.Blocks[0].Parents...))
	if len(pts.Cids) == 0 || len(pts.Blocks) == 0 || pts.Height != ts.Height-1 {
		t.Fatalf("invalid tipset")
	}
}

func TestStateReadState(t *testing.T) {
	addr, token := tests.ClientConfig()
	c, cls, err := New(addr, token)
	checkErr(t, err)
	defer cls()

	addrs, err := c.StateListMiners(context.Background(), nil)
	checkErr(t, err)

	for _, a := range addrs {
		actor, err := c.StateGetActor(context.Background(), a, nil)
		checkErr(t, err)
		s, err := c.StateReadState(context.Background(), actor, nil)
		checkErr(t, err)
		if s.State == nil {
			t.Fatalf("state of actor %s can't be nil", a)
		}
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
