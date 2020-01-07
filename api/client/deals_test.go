package client

import (
	"strings"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/textileio/filecoin/deals"
	pb "github.com/textileio/filecoin/deals/pb"
)

func TestQueryAsk(t *testing.T) {
	d, done := setupDeals(t)
	defer done()

	_, err := d.AvailableAsks(ctx, deals.Query{MaxPrice: 5})
	if err != nil {
		t.Fatalf("failed to call AvailableAsks: %v", err)
	}
}

func TestStore(t *testing.T) {
	d, done := setupDeals(t)
	defer done()

	r := strings.NewReader("store me")
	_, _, err := d.Store(ctx, "an address", r, make([]deals.DealConfig, 0), 1024)
	if err != nil {
		t.Fatalf("failed to call Store: %v", err)
	}
}

func TestWatch(t *testing.T) {
	d, done := setupDeals(t)
	defer done()

	_, err := d.Watch(ctx, make([]cid.Cid, 0))
	if err != nil {
		t.Fatalf("failed to call Watch: %v", err)
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
