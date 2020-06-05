package api

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

// ListPayChannels lists all payment channels associated with this FFS
func (i *API) ListPayChannels(ctx context.Context) ([]ffs.PaychInfo, error) {
	addrInfos := i.Addrs()
	addrs := make([]string, len(addrInfos))
	for i, info := range addrInfos {
		addrs[i] = info.Addr
	}
	res, err := i.pm.List(ctx, addrs...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// CreatePayChannel creates a new payment channel
func (i *API) CreatePayChannel(ctx context.Context, from string, to string, amount uint64) (ffs.PaychInfo, cid.Cid, error) {
	if !i.isManagedAddress(from) {
		return ffs.PaychInfo{}, cid.Undef, fmt.Errorf("%v is not managed by ffs instance", from)
	}
	return i.pm.Create(ctx, from, to, amount)
}

// RedeemPayChannel redeems a payment channel
func (i *API) RedeemPayChannel(ctx context.Context, addr string) error {
	channels, err := i.ListPayChannels(ctx)
	if err != nil {
		return err
	}
	managed := false
	for _, channel := range channels {
		if channel.Addr == addr {
			managed = true
			break
		}
	}
	if !managed {
		return fmt.Errorf("paych %v is not managed by ffs instance", addr)
	}
	return i.pm.Redeem(ctx, addr)
}
