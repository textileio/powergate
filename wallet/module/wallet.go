package module

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	logger "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/v2/lotus"
)

const (
	feeThreshold = 1_000_000
)

var (
	log           = logger.Logger("wallet")
	networkFaucet = map[string]string{}
)

// Module exposes the filecoin wallet api.
type Module struct {
	clientBuilder lotus.ClientBuilder
	iAmount       *big.Int
	masterAddr    address.Address
	networkName   string
}

// New creates a new wallet module.
func New(clientBuilder lotus.ClientBuilder, maddr address.Address, iam big.Int, autocreate bool, networkName string) (*Module, error) {
	m := &Module{
		clientBuilder: clientBuilder,
		iAmount:       &iam,
		masterAddr:    maddr,
		networkName:   networkName,
	}
	if maddr == address.Undef && autocreate {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		newMasterAddr, err := m.NewAddress(ctx, "bls")
		if err != nil {
			return nil, fmt.Errorf("creating and funding master addr: %s", err)
		}
		log.Infof("Auto-created master wallet addr: %s", newMasterAddr)
		if err := m.FundFromFaucet(ctx, newMasterAddr); err != nil {
			return nil, fmt.Errorf("funding new master addr: %s", err)
		}
		log.Info("Autocreated master wallet addr funded successfully")
		maddr, _ = address.NewFromString(newMasterAddr)
		m = &Module{
			clientBuilder: clientBuilder,
			iAmount:       &iam,
			masterAddr:    maddr,
		}
	}

	return m, nil
}

// MasterAddr returns the master address.
// Will return address.Undef is Powergate was started with no master address.
func (m *Module) MasterAddr() address.Address {
	return m.masterAddr
}

// NewAddress creates a new address.
func (m *Module) NewAddress(ctx context.Context, typ string) (string, error) {
	var ty types.KeyType
	if typ == "bls" {
		ty = types.KTBLS
	} else if typ == "secp256k1" {
		ty = types.KTSecp256k1
	} else {
		return "", fmt.Errorf("unknown address type %s", typ)
	}

	client, cls, err := m.clientBuilder(ctx)
	if err != nil {
		return "", fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	addr, err := client.WalletNew(ctx, ty)
	if err != nil {
		return "", err
	}

	if m.masterAddr != address.Undef {
		balance, err := client.WalletBalance(ctx, m.masterAddr)
		if err != nil {
			return "", fmt.Errorf("getting balance from master addr: %s", err)
		}
		if balance.LessThan(types.BigAdd(types.BigInt{Int: m.iAmount}, types.NewInt(feeThreshold))) {
			return "", fmt.Errorf("balance %d is less than allowed threshold", balance)
		}
		go func() {
			client, cls, err := m.clientBuilder(context.Background())
			if err != nil {
				log.Errorf("creating lotus client: %s", err)
				return
			}
			defer cls()

			msg := &types.Message{
				From:  m.masterAddr,
				To:    addr,
				Value: types.BigInt{Int: m.iAmount},
			}
			smsg, err := client.MpoolPushMessage(context.Background(), msg, nil)
			if err != nil {
				log.Errorf("transferring funds to new address: %s", err)
				return
			}
			log.Infof("%s funding transaction message: %s", addr, smsg.Message.Cid())
		}()
	}

	return addr.String(), nil
}

// Sign signs a message with an address.
func (m *Module) Sign(ctx context.Context, addr string, message []byte) ([]byte, error) {
	client, cls, err := m.clientBuilder(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	waddr, err := address.NewFromString(addr)
	if err != nil {
		return nil, fmt.Errorf("parsing wallet address: %s", err)
	}
	sig, err := client.WalletSign(ctx, waddr, message)
	if err != nil {
		return nil, fmt.Errorf("lotus signing message: %s", err)
	}
	sigBytes, err := sig.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshaling signature: %s", err)
	}

	return sigBytes, nil
}

// Verify verifies a message signature from an address.
func (m *Module) Verify(ctx context.Context, addr string, message, signature []byte) (bool, error) {
	client, cls, err := m.clientBuilder(ctx)
	if err != nil {
		return false, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	waddr, err := address.NewFromString(addr)
	if err != nil {
		return false, fmt.Errorf("parsing wallet address: %s", err)
	}
	var sig crypto.Signature
	if err := sig.UnmarshalBinary(signature); err != nil {
		return false, fmt.Errorf("unmarshaling signature: %s", err)
	}
	return client.WalletVerify(ctx, waddr, message, &sig)
}

// List returns all wallet addresses.
func (m *Module) List(ctx context.Context) ([]string, error) {
	client, cls, err := m.clientBuilder(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	addrs, err := client.WalletList(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting wallet addresses: %v", err)
	}
	ret := make([]string, len(addrs))
	for i, addr := range addrs {
		ret[i] = addr.String()
	}
	return ret, nil
}

// Balance returns the balance of the specified address.
func (m *Module) Balance(ctx context.Context, addr string) (*big.Int, error) {
	client, cls, err := m.clientBuilder(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	a, err := address.NewFromString(addr)
	if err != nil {
		return nil, err
	}
	b, err := client.WalletBalance(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("getting balance from lotus: %s", err)
	}
	return b.Int, nil
}

// SendFil sends fil from one address to another.
func (m *Module) SendFil(ctx context.Context, from string, to string, amount *big.Int) (cid.Cid, error) {
	f, err := address.NewFromString(from)
	if err != nil {
		return cid.Cid{}, err
	}
	t, err := address.NewFromString(to)
	if err != nil {
		return cid.Cid{}, err
	}
	msg := &types.Message{
		From:  f,
		To:    t,
		Value: types.BigInt{Int: amount},
	}
	client, cls, err := m.clientBuilder(ctx)
	if err != nil {
		return cid.Cid{}, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	sm, err := client.MpoolPushMessage(ctx, msg, nil)
	if err != nil {
		return cid.Cid{}, err
	}

	return sm.Message.Cid(), err
}

// FundFromFaucet make a faucet call to fund the provided wallet address.
func (m *Module) FundFromFaucet(ctx context.Context, addr string) error {
	faucet, ok := networkFaucet[m.networkName]
	if !ok {
		return fmt.Errorf("unknown faucet for network %s", m.networkName)
	}

	req, err := http.NewRequest("GET", faucet+"/send", nil)
	if err != nil {
		return fmt.Errorf("parsing fountain url: %s", err)
	}
	q := req.URL.Query()
	q.Add("address", addr)
	req.URL.RawQuery = q.Encode()
	r, err := http.Get(req.URL.String())
	if err != nil {
		return fmt.Errorf("calling the fountain: %s", err)
	}
	defer func() { _ = r.Body.Close() }()
	_, _ = io.Copy(ioutil.Discard, r.Body)
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("fountain request not OK: %s", r.Status)
	}
	return nil
}
