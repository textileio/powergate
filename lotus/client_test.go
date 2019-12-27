package lotus

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/textileio/filecoin/tests"
)

func TestClientVersion(t *testing.T) {
	addr, token := tests.ClientConfig(t)
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

	addr, token := tests.ClientConfig(t)
	c, cls, err := New(addr, token)
	checkErr(t, err)
	defer cls()

	cid, err := c.ClientImport(context.Background(), f.Name())
	checkErr(t, err)
	if !cid.Defined() {
		t.Errorf("undefined cid from import")
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
