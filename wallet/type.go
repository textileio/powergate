package wallet

import (
	"context"
	"errors"
	"math/big"

	"github.com/ipfs/go-cid"
)

var (
	// ErrNoVerifiedClient indicates that the provided
	// wallet-address isn't a verified client.
	ErrNoVerifiedClient = errors.New("the wallet-address isn't a verified client")
)

// VerifiedClientInfo contains information for a wallet address
// that is a verified-client.
type VerifiedClientInfo struct {
	RemainingDatacapBytes *big.Int
}

// Module provides wallet management access to a Filecoin client.
type Module interface {
	// NewAddress returns a new wallet address of the specified type.
	NewAddress(ctx context.Context, typ string) (string, error)
	// List lists all known wallet addresses.
	List(ctx context.Context) ([]string, error)
	// GetVerifiedClientInfo returns details about a wallet-address that's
	// a verified client. If the wallet address isn't a verified client,
	// it will return ErrNoVerifiedClient.
	GetVerifiedClientInfo(ctx context.Context, addr string) (VerifiedClientInfo, error)
	// Balance returns the balance of a wallet-address.
	Balance(ctx context.Context, addr string) (*big.Int, error)
	// SendFil sends `amount` attoFIL from `from` to `to` and returns the
	// message Cid.
	SendFil(ctx context.Context, from string, to string, amount *big.Int) (cid.Cid, error)
	// FundFromFaucet funds addr from the master address (if configured).
	FundFromFaucet(ctx context.Context, addr string) error
}
