package lotus

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin/paych"
	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

// Module provides access to the paych api
type Module struct {
	api *apistruct.FullNodeStruct
}

var _ ffs.PaychManager = (*Module)(nil)

// New creates a new paych module
func New(api *apistruct.FullNodeStruct) *Module {
	return &Module{
		api: api,
	}
}

// List lists all payment channels involving the specified addresses
func (m *Module) List(ctx context.Context, addrs ...string) ([]ffs.PaychInfo, error) {
	filter := make(map[string]struct{}, len(addrs))
	for _, addr := range addrs {
		filter[addr] = struct{}{}
	}

	allAddrs, err := m.api.PaychList(ctx)
	if err != nil {
		return nil, err
	}

	chans := make([]<-chan statusResult, len(allAddrs))
	for i, addr := range allAddrs {
		chans[i] = m.paychStatus(ctx, addr)
	}

	resultsCh := make(chan statusResult, len(chans))
	for _, c := range chans {
		go func(ch <-chan statusResult) {
			res := <-ch
			resultsCh <- res
		}(c)
	}

	results := make([]statusResult, len(chans))
	for i := 0; i < len(chans); i++ {
		results[i] = <-resultsCh
	}

	var final []ffs.PaychInfo
	for _, result := range results {
		if result.err != nil {
			// ToDo: do we want to fail if there was an error checking
			// even one status that may not even be in our filter?
			continue
		}
		_, addrInFilter := filter[result.addr.String()]
		_, ctlAddrInFilter := filter[result.status.ControlAddr.String()]
		if addrInFilter || ctlAddrInFilter {
			var dir ffs.PaychDir
			switch result.status.Direction {
			case api.PCHUndef:
				dir = ffs.PaychDirUnspecified
			case api.PCHInbound:
				dir = ffs.PaychDirInbound
			case api.PCHOutbound:
				dir = ffs.PaychDirOutbound
			default:
				return nil, fmt.Errorf("unknown pay channel direction %v", result.status.Direction)
			}

			info := ffs.PaychInfo{
				CtlAddr:   result.status.ControlAddr.String(),
				Addr:      result.addr.String(),
				Direction: dir,
			}

			final = append(final, info)
		}
	}

	return final, nil
}

// Create creates a new payment channel
func (m *Module) Create(ctx context.Context, from string, to string, amount uint64) (ffs.PaychInfo, cid.Cid, error) {
	// return ffs.PaychInfo{}, cid.Undef, fmt.Errorf("unimplemeted for now, blocked by lotus issue #1611")
	fAddr, err := address.NewFromString(from)
	if err != nil {
		return ffs.PaychInfo{}, cid.Undef, err
	}
	tAddr, err := address.NewFromString(from)
	if err != nil {
		return ffs.PaychInfo{}, cid.Undef, err
	}
	a := types.NewInt(amount)
	info, err := m.api.PaychGet(ctx, fAddr, tAddr, a)
	if err != nil {
		return ffs.PaychInfo{}, cid.Undef, err
	}
	// ToDo: verify these addresses and direction make sense
	res := ffs.PaychInfo{
		CtlAddr:   from,
		Addr:      info.Channel.String(),
		Direction: ffs.PaychDirOutbound,
	}
	return res, info.ChannelMessage, nil
}

// Redeem redeems a payment channel
func (m *Module) Redeem(ctx context.Context, ch string) error {
	chAddr, err := address.NewFromString(ch)
	if err != nil {
		return err
	}
	vouchers, err := m.api.PaychVoucherList(ctx, chAddr)
	if err != nil {
		return err
	}

	var best *paych.SignedVoucher
	for _, v := range vouchers {
		spendable, err := m.api.PaychVoucherCheckSpendable(ctx, chAddr, v, nil, nil)
		if err != nil {
			return err
		}
		if spendable {
			if best == nil || v.Amount.GreaterThan(best.Amount) {
				best = v
			}
		}
	}

	if best == nil {
		return fmt.Errorf("No spendable vouchers for that channel")
	}

	mcid, err := m.api.PaychVoucherSubmit(ctx, chAddr, best)
	if err != nil {
		return err
	}

	mwait, err := m.api.StateWaitMsg(ctx, mcid)
	if err != nil {
		return err
	}

	if mwait.Receipt.ExitCode != 0 {
		return fmt.Errorf("submit voucher message execution failed (exit code %d)", mwait.Receipt.ExitCode)
	}

	return nil
}

type statusResult struct {
	addr   address.Address
	status *api.PaychStatus
	err    error
}

func (m *Module) paychStatus(ctx context.Context, addr address.Address) <-chan statusResult {
	c := make(chan statusResult)
	go func() {
		defer close(c)
		status, err := m.api.PaychStatus(ctx, addr)
		c <- statusResult{addr: addr, status: status, err: err}
	}()
	return c
}
