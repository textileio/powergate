package wallet

import (
	"context"
	"fmt"
	"math/big"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/crypto"
	"github.com/textileio/lotus-client/api/apistruct"
)

// Module exposes the filecoin wallet api.
type Module struct {
	api        *apistruct.FullNodeStruct
	iAmount    *big.Int
	masterAddr *address.Address
}

// New creates a new wallet module
func New(api *apistruct.FullNodeStruct, maddr *address.Address, iam big.Int) (*Module, error) {
	m := &Module{
		api:        api,
		iAmount:    &iam,
		masterAddr: maddr,
	}
	return m, nil
}

// NewAddress creates a new wallet
func (m *Module) NewAddress(ctx context.Context, typ string) (string, error) {
	ty := crypto.SigTypeUnknown
	if "bls" == typ {
		ty = crypto.SigTypeBLS
	} else if "secp256k1" == typ {
		ty = crypto.SigTypeSecp256k1
	} else {
		return "", fmt.Errorf("unkown wallet type %s", typ)
	}

	addr, err := m.api.WalletNew(ctx, ty)
	if err != nil {
		return "", err
	}

	if m.masterAddr != nil {
		msg := &types.Message{
			From:     *m.masterAddr,
			To:       addr,
			Value:    types.BigInt{Int: m.iAmount},
			GasLimit: 1000,
			GasPrice: types.NewInt(0),
		}

		_, err = m.api.MpoolPushMessage(ctx, msg)
		if err != nil {
			return "", fmt.Errorf("transfering funds to new wallet: %s", err)
		}
	}

	return addr.String(), nil
}

// Balance returns the balance of the specified wallet
func (m *Module) Balance(ctx context.Context, addr string) (uint64, error) {
	a, err := address.NewFromString(addr)
	if err != nil {
		return 0, err
	}
	b, err := m.api.WalletBalance(ctx, a)
	if err != nil {
		return 0, fmt.Errorf("getting balance from lotus: %s", err)
	}
	return b.Uint64(), nil
}
