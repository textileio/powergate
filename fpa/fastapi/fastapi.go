package fastapi

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/textileio/fil-tools/deals"
	ftypes "github.com/textileio/fil-tools/fpa/types"
)

var (
	dsKeyInstance = ds.NewKey("instance")
	dsKeyCid      = ds.NewKey("cid")

	defaultWalletType = "bls"

	log = logging.Logger("fastapi")
)

type Instance struct {
	ipfs    iface.CoreAPI
	dm      *deals.Module
	ms      ftypes.MinerSelector
	wm      ftypes.WalletManager
	auditer ftypes.Auditer

	lock sync.Mutex
	info info
	ds   ds.Datastore
}

func New(ctx context.Context,
	store ds.Datastore,
	ipfs iface.CoreAPI,
	dm *deals.Module,
	ms ftypes.MinerSelector,
	a ftypes.Auditer,
	wm ftypes.WalletManager) (*Instance, error) {
	addr, err := wm.NewWallet(ctx, defaultWalletType)
	if err != nil {
		return nil, fmt.Errorf("creating new wallet addr: %s", err)
	}

	info := info{
		ID:         NewID(),
		WalletAddr: addr,
	}
	i := &Instance{
		ds:      store,
		ipfs:    ipfs,
		dm:      dm,
		ms:      ms,
		auditer: a,
		wm:      wm,
		info:    info,
	}
	if err := i.saveInfo(); err != nil {
		return nil, fmt.Errorf("saving new instance %s: %s", i.info.ID, err)
	}
	return i, nil
}

func LoadFromID(store ds.Datastore,
	ipfsClient iface.CoreAPI,
	dm *deals.Module,
	ms ftypes.MinerSelector,
	a ftypes.Auditer,
	wm ftypes.WalletManager,
	id ID) (*Instance, error) {
	buf, err := store.Get(makeKeyInstance(id))
	if err != nil && err == ds.ErrNotFound {
		return nil, fmt.Errorf("instance doesn't exist")
	}
	if err != nil {
		return nil, fmt.Errorf("loading instance %s: %s", id, err)
	}
	var info info
	if err := json.Unmarshal(buf, &info); err != nil {
		return nil, fmt.Errorf("loading instance %s from datastore: %s", id, err)
	}
	return &Instance{
		ds:      store,
		ipfs:    ipfsClient,
		dm:      dm,
		ms:      ms,
		wm:      wm,
		auditer: a,
		info:    info,
	}, nil
}

func (i *Instance) ID() ID {
	return i.info.ID
}

func (i *Instance) WalletAddr() string {
	return i.info.WalletAddr
}

func (i *Instance) Show(c cid.Cid) (CidInfo, error) {
	var ci CidInfo
	info, exist, err := i.getCidInfo(c)
	if err != nil {
		return ci, fmt.Errorf("getting cid information: %s", err)
	}
	if !exist {
		return ci, ErrNotStored
	}
	return info, nil
}

func (i *Instance) Info(ctx context.Context) (Info, error) {
	inf := Info{
		ID: i.info.ID,
		Wallet: WalletInfo{
			Address: i.info.WalletAddr,
		},
	}

	var err error
	inf.Pins, err = i.getPinnedCids()
	if err != nil {
		return inf, fmt.Errorf("getting pins from instance: %s", err)
	}

	inf.Wallet.Balance, err = i.wm.Balance(ctx, i.info.WalletAddr)
	if err != nil {
		return inf, fmt.Errorf("getting balance of %s: %s", i.info.WalletAddr, err)
	}

	return inf, nil
}

func (i *Instance) saveCidInfo(cinfo CidInfo) error {
	buf, err := json.Marshal(cinfo)
	if err != nil {
		return err
	}
	if err := i.ds.Put(makeKeyCid(i.info.ID, cinfo.Cid), buf); err != nil {
		return err
	}
	return nil
}

func (i *Instance) getCidInfo(c cid.Cid) (CidInfo, bool, error) {
	var ci CidInfo
	buf, err := i.ds.Get(makeKeyCid(i.info.ID, c))
	if err == ds.ErrNotFound {
		return ci, false, nil
	}
	if err != nil {
		return ci, false, err
	}
	if err := json.Unmarshal(buf, &ci); err != nil {
		return ci, false, err
	}
	return ci, true, err
}

func (i *Instance) getPinnedCids() ([]cid.Cid, error) {
	q := query.Query{
		Prefix:   makeKeyInstance(i.info.ID).Child(dsKeyCid).String(),
		KeysOnly: true,
	}
	res, err := i.ds.Query(q)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var cids []cid.Cid
	for r := range res.Next() {
		strCid := ds.RawKey(r.Key).Name()
		c, _ := cid.Decode(strCid)
		cids = append(cids, c)
	}
	return cids, nil
}

func (i *Instance) saveInfo() error {
	i.lock.Lock()
	defer i.lock.Unlock()

	k := makeKeyInstance(i.info.ID)
	buf, err := json.Marshal(i.info)
	if err != nil {
		return err
	}
	if err := i.ds.Put(k, buf); err != nil {
		return err
	}
	return nil
}

func makeKeyCid(iid ID, c cid.Cid) datastore.Key {
	return makeKeyInstance(iid).Child(dsKeyCid).ChildString(c.String())
}

func makeKeyInstance(id ID) ds.Key {
	return dsKeyInstance.ChildString(id.String())
}
