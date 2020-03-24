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

	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/lotus-client/chain/types"
	"github.com/textileio/powergate/tests"
)

const (
	tmpDir = "/tmp/powergate"
)

func TestMain(m *testing.M) {
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		os.Mkdir(tmpDir, os.ModeDir)
	}
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestClientVersion(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)

	if _, err := client.Version(context.Background()); err != nil {
		t.Fatalf("error when getting client version: %s", err)
	}
}

func TestClientImport(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)

	f, err := ioutil.TempFile(tmpDir, "")
	checkErr(t, err)
	defer os.Remove(f.Name())
	defer f.Close()
	bts := make([]byte, 4)
	rand.Read(bts)
	io.Copy(f, bytes.NewReader(bts))

	cid, err := client.ClientImport(context.Background(), f.Name())
	checkErr(t, err)
	if !cid.Defined() {
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

func TestStateReadState(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)
	addrs, err := client.StateListMiners(context.Background(), types.EmptyTSK)
	checkErr(t, err)

	for _, a := range addrs {
		actor, err := client.StateGetActor(context.Background(), a, types.EmptyTSK)
		checkErr(t, err)
		s, err := client.StateReadState(context.Background(), actor, types.EmptyTSK)
		checkErr(t, err)
		if s.State == nil {
			t.Fatalf("state of actor %s can't be nil", a)
		}
	}
}

func TestGetPeerID(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)

	miners, err := client.StateListMiners(context.Background(), types.EmptyTSK)
	checkErr(t, err)

	pid, err := client.StateMinerPeerID(context.Background(), miners[0], types.EmptyTSK)
	checkErr(t, err)
	checkErr(t, pid.Validate())
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
