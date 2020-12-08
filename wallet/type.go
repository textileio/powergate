package wallet

import (
	"context"
	"math/big"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
)

// Module provides wallet management access to a Filecoin client.
type Module interface {
	NewAddress(ctx context.Context, typ string) (string, error)
	List(ctx context.Context) ([]string, error)
	Balance(ctx context.Context, addr string) (*big.Int, error)
	SendFil(ctx context.Context, from string, to string, amount *big.Int) error
	FundFromFaucet(ctx context.Context, addr string) error
}

// SendFilEvent is a record created when calling SendFil.
type SendFilEvent struct {
	Cid    cid.Cid
	From   address.Address
	To     address.Address
	Amount *big.Int
	Time   time.Time
}
