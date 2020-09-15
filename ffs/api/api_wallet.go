package api

import (
	"context"
	"fmt"
	"math/big"
)

// Addrs returns the wallet addresses.
func (i *API) Addrs() []AddrInfo {
	i.lock.Lock()
	defer i.lock.Unlock()
	var addrs []AddrInfo
	for _, addr := range i.cfg.Addrs {
		addrs = append(addrs, addr)
	}
	return addrs
}

// NewAddr creates a new address managed by the FFS instance.
func (i *API) NewAddr(ctx context.Context, name string, options ...NewAddressOption) (string, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	conf := &NewAddressConfig{
		makeDefault: false,
		addressType: "bls",
	}
	for _, option := range options {
		option(conf)
	}

	exists := false
	for _, addr := range i.cfg.Addrs {
		if addr.Name == name {
			exists = true
			break
		}
	}

	if exists {
		return "", fmt.Errorf("address with name %s already exists", name)
	}

	addr, err := i.wm.NewAddress(ctx, conf.addressType)
	if err != nil {
		return "", fmt.Errorf("creating new wallet addr: %s", err)
	}
	i.cfg.Addrs[addr] = AddrInfo{
		Name: name,
		Addr: addr,
		Type: conf.addressType,
	}
	if conf.makeDefault {
		i.cfg.DefaultStorageConfig.Cold.Filecoin.Addr = addr
	}
	if err := i.is.putInstanceConfig(i.cfg); err != nil {
		return "", err
	}
	return addr, nil
}

// SendFil sends fil from a managed address to any another address, returns immediately but funds are sent asynchronously.
func (i *API) SendFil(ctx context.Context, from string, to string, amount *big.Int) error {
	if !i.isManagedAddress(from) {
		return fmt.Errorf("%v is not managed by ffs instance", from)
	}
	return i.wm.SendFil(ctx, from, to, amount)
}

func (i *API) isManagedAddress(addr string) bool {
	for managedAddr := range i.cfg.Addrs {
		if managedAddr == addr {
			return true
		}
	}
	return false
}
