package client

import (
	"strings"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/deals"
	pb "github.com/textileio/powergate/deals/pb"
)

func TestStore(t *testing.T) {
	skipIfShort(t)
	d, done := setupDeals(t)
	defer done()

	r := strings.NewReader("store me")
	_, _, err := d.Store(ctx, "an address", r, make([]deals.StorageDealConfig, 0), 1024)
	if err != nil {
		t.Fatalf("failed to call Store: %v", err)
	}
}

func TestWatch(t *testing.T) {
	skipIfShort(t)
	d, done := setupDeals(t)
	defer done()

	_, err := d.Watch(ctx, make([]cid.Cid, 0))
	if err != nil {
		t.Fatalf("failed to call Watch: %v", err)
	}
}

func TestRetrieve(t *testing.T) {
	skipIfShort(t)
	d, done := setupDeals(t)
	defer done()

	cid, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPA")

	_, err := d.Retrieve(ctx, "an address", cid)
	if err != nil {
		t.Fatalf("failed to call Retrieve: %v", err)
	}
}

func setupDeals(t *testing.T) (*Deals, func()) {
	serverDone := setupServer(t)
	conn, done := setupConnection(t)
	return &Deals{client: pb.NewAPIClient(conn)}, func() {
		done()
		serverDone()
	}
}
