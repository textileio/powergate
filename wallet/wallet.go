package wallet

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
)

// API interacts with a Filecoin full-node
type API interface {
	WalletNew(context.Context, string) (address.Address, error)
	WalletBalance(context.Context, address.Address) (types.BigInt, error)
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
	addr, err := m.api.WalletNew(ctx, typ)
	if err != nil {
		return "", err
	}
	return addr.String(), nil
}

// WalletBalance returns the balance of the specified wallet
func (m *Module) WalletBalance(ctx context.Context, addr string) (types.BigInt, error) {
	a, err := address.NewFromString(addr)
	if err != nil {
		return types.EmptyInt, err
	}
	return m.api.WalletBalance(ctx, a)
}
