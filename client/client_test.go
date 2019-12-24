package client

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/textileio/filecoin/tests"
)

var (
	authToken = ""
)

func TestMain(m *testing.M) {
	var err error
	authToken, err = tests.GetLotusToken()
	if err != nil {
		fmt.Println("couldn't get/generate lotus authtoken")
		os.Exit(-1)
	}
	os.Exit(m.Run())
}

func TestClientVersion(t *testing.T) {
	c, cls, err := New(tests.DaemonAddr, authToken)
	checkErr(t, err)
	defer cls()
	if _, err := c.Version(context.Background()); err != nil {
		t.Fatalf("error when getting client version: %s", err)
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
