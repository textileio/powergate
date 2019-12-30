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

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
