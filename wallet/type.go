package wallet

import (
	"context"
)

// Module provides wallet management access to a Filecoin client.
type Module interface {
	NewAddress(ctx context.Context, typ string) (string, error)
	List(ctx context.Context) ([]string, error)
	Balance(ctx context.Context, addr string) (uint64, error)
	FundFromFaucet(ctx context.Context, addr string) error
}
