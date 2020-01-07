package wallet

import (
	"context"
	"time"

	"github.com/textileio/filecoin/lotus/types"
)

const (
	initialWait        = time.Second * 5
	chanWriteTimeout   = time.Second
	askRefreshInterval = time.Second * 10
)

// API interacts with a Filecoin full-node
type API interface {
	WalletNew(ctx context.Context, typ string) (string, error)
	WalletBalance(ctx context.Context, addr string) (types.BigInt, error)
}

// Module exposes the filecoin wallet api.
type Module struct {
	api API
}

// New creates a new wallet module
func New(api API) *Module {
	m := &Module{
		api: api,
	}
	return m
}

// NewWallet creates a new wallet
func (m *Module) NewWallet(ctx context.Context, typ string) (string, error) {
	return m.api.WalletNew(ctx, typ)
}

// WalletBalance returns the balance of the specified wallet
func (m *Module) WalletBalance(ctx context.Context, addr string) (types.BigInt, error) {
	return m.api.WalletBalance(ctx, addr)
}
