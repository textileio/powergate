package wallet

import (
	"context"
	"math/big"
)

// Module provides wallet management access to a Filecoin client.
type Module interface {
	NewAddress(ctx context.Context, typ string) (string, error)
	Balance(ctx context.Context, addr string) (uint64, error)
	SendFil(ctx context.Context, from string, to string, amount *big.Int) error
	FundFromFaucet(ctx context.Context, addr string) error
}
