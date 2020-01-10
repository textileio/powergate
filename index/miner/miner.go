package miner

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/textileio/filecoin/fchost"
	"github.com/textileio/filecoin/iplocation"
	"github.com/textileio/filecoin/lotus/types"
	"github.com/textileio/filecoin/signaler"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

const (
	fullRefreshThreshold = 100
	deltaRefreshHeight   = 100
)

var (
	dsBase = datastore.NewKey("/index/miner")

	maxParallelCalc = 10

	log = logging.Logger("index-miner")
)

type API interface {
	StateListMiners(context.Context, *types.TipSet) ([]string, error)
	StateMinerPower(context.Context, string, *types.TipSet) (types.MinerPower, error)
	ChainNotify(context.Context) (<-chan []*types.HeadChange, error)
	ChainHead(context.Context) (*types.TipSet, error)
	ChainGetTipSet(context.Context, types.TipSetKey) (*types.TipSet, error)
	ChainGetTipSetByHeight(context.Context, uint64, *types.TipSet) (*types.TipSet, error)
	StateChangedActors(context.Context, cid.Cid, cid.Cid) (map[string]types.Actor, error)
	StateReadState(ctx context.Context, act *types.Actor, ts *types.TipSet) (*types.ActorState, error)
	StateMinerPeerID(ctx context.Context, m string, ts *types.TipSet) (peer.ID, error)
}

type MinerIndex struct {
	c        API
	ds       datastore.TxnDatastore
	h        *fchost.FilecoinHost
	lr       iplocation.LocationResolver
	signaler *signaler.Signaler

	lock  sync.Mutex
	index Index

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
	closed   bool
}

type Index struct {
	LastUpdated uint64
	Miners      map[string]Info
}

type Info struct {
	Static Static
	Power  Power
}

type Power struct {
	Power    uint64
	Relative float64
}

func New(ds datastore.TxnDatastore, c API, h *fchost.FilecoinHost, lr iplocation.LocationResolver) *MinerIndex {
	initMetrics()
	ctx, cancel := context.WithCancel(context.Background())
	mi := &MinerIndex{
		c:        c,
		ds:       ds,
		signaler: signaler.New(),
		h:        h,
		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	go mi.start()
	return mi
}

func (mi *MinerIndex) Get() Index {
	mi.lock.Lock()
	ii := Index{
		LastUpdated: mi.index.LastUpdated,
		Miners:      make(map[string]Info, len(mi.index.Miners)),
	}
	for addr, v := range mi.index.Miners {
		ii.Miners[addr] = v
	}
	mi.lock.Unlock()
	return ii
}

func (mi *MinerIndex) start() {
	defer close(mi.finished)
	n, err := mi.c.ChainNotify(mi.ctx)
	if err != nil {
		log.Fatalf("error when getting notify channel from lotus: %s", err)
	}
	for {
		select {
		case <-mi.ctx.Done():
			log.Info("graceful shutdown of background miner index")
			return
		case _, ok := <-n:
			if !ok {
				log.Fatalf("lotus notify channel closed: %s", err)
			}
			log.Info("updating miner index...")
			if err := mi.update(); err != nil {
				log.Errorf("error when updating miner index: %s", err)
				continue
			}
			log.Info("miner index updated")
		}
	}
}

func (mi *MinerIndex) Listen() <-chan struct{} {
	return mi.signaler.Listen()
}

func (mi *MinerIndex) Unregister(c chan struct{}) {
	mi.signaler.Unregister(c)
}

func (mi *MinerIndex) update() error {
	lastHeight := uint64(1)
	dsLastHeight := dsBase.ChildString("height")
	b, err := mi.ds.Get(dsLastHeight)
	if err != nil && err != datastore.ErrNotFound {
		return fmt.Errorf("couldn't get stored last height: %s", err)
	}
	if err != datastore.ErrNotFound {
		lastHeight, _ = binary.Uvarint(b)
	}
	log.Debugf("updating from height: %d", lastHeight)

	ts, err := mi.c.ChainHead(mi.ctx) // ToDo: think if this is correct since we could have applys, rollbacks, etc
	if err != nil {
		return fmt.Errorf("error when getting heaviest head from chain: %s", err)
	}

	ctx, _ := tag.New(context.Background())
	start := time.Now()
	if ts.Height-lastHeight > fullRefreshThreshold {
		ctx, _ = tag.New(ctx, tag.Insert(kRefreshType, "full"))
		log.Infof("doing full refresh, current %d,  last %d", ts.Height, lastHeight)
		if err := mi.fullRefresh(ts); err != nil {
			return fmt.Errorf("error doing full refresh: %s", err)
		}
	} else {
		ctx, _ = tag.New(ctx, tag.Insert(kRefreshType, "delta"))
		log.Debugf("doing delta refresh, current %d, last %d", ts.Height, lastHeight)
		if err := mi.deltaRefresh(ts, deltaRefreshHeight); err != nil {
			return fmt.Errorf("error doing delta refresh: %s", err)
		}
		log.Debug("delta refresh done")
	}
	stats.Record(ctx, mRefreshTime.M(int64(time.Since(start).Milliseconds())))

	// ToDo: save miner index in datastore
	var buf [32]byte
	n := binary.PutUvarint(buf[:], mi.index.LastUpdated)
	if err := mi.ds.Put(dsLastHeight, buf[:n]); err != nil {
		return fmt.Errorf("error saving new height in datastore: %s", err)
	}
	mi.signaler.Signal()
	stats.Record(context.Background(), mUpdatedHeight.M(int64(mi.index.LastUpdated)))

	return nil
}

func (mi *MinerIndex) deltaRefresh(ts *types.TipSet, deltaHeight int) error {
	baseHeight := ts.Height - uint64(deltaHeight)
	if ts.Height < baseHeight {
		log.Fatal("reversed node or chain rollback, current height %d, base height %d", ts.Height, baseHeight)
	}

	currHeight := ts.Height
	var err error
	baseTs, err := mi.c.ChainGetTipSetByHeight(mi.ctx, baseHeight, ts)
	if err != nil {
		return err
	}
	chg, err := mi.c.StateChangedActors(mi.ctx, baseTs.Blocks[0].ParentStateRoot, ts.Blocks[0].ParentStateRoot)
	if err != nil {
		return err
	}
	changedMiners := make(map[string]struct{})
	for addr := range chg {
		changedMiners[addr] = struct{}{}
	}
	addrs := make([]string, 0, len(changedMiners))
	for addr := range changedMiners {
		addrs = append(addrs, addr)
	}
	mi.updateForAddrs(addrs, currHeight)

	return nil
}

func (mi *MinerIndex) fullRefresh(head *types.TipSet) error {
	addrs, err := mi.c.StateListMiners(mi.ctx, nil)
	if err != nil {
		return err
	}
	mi.lock.Lock()
	mi.index.Miners = make(map[string]Info)
	mi.lock.Unlock()

	mi.updateForAddrs(addrs, head.Height)

	return nil
}

func (mi *MinerIndex) updateForAddrs(addrs []string, height uint64) {
	rateLim := make(chan struct{}, maxParallelCalc)
	for i, a := range addrs {
		rateLim <- struct{}{}
		go func(addr string) {
			defer func() { <-rateLim }()
			if err := mi.updateMinerInformation(addr); err != nil {
				log.Debugf("error generating reputation for miner %s: %s", addr, err)
				return
			}
		}(a)
		stats.Record(context.Background(), mRefreshProgress.M(float64(i)/float64(len(addrs))))
	}
	for i := 0; i < maxParallelCalc; i++ {
		rateLim <- struct{}{}
	}
	stats.Record(context.Background(), mRefreshProgress.M(1))
	mi.lock.Lock()
	var online int64
	for _, v := range mi.index.Miners {
		if v.Static.Online {
			online++
		}
	}
	offline := int64(len(mi.index.Miners)) - online
	mi.lock.Unlock()

	ctx, _ := tag.New(context.Background(), tag.Insert(kOnline, "online"))
	stats.Record(ctx, mMinerCount.M(online))
	ctx, _ = tag.New(context.Background(), tag.Insert(kOnline, "offline"))
	stats.Record(ctx, mMinerCount.M(offline))
}

func (mi *MinerIndex) updateMinerInformation(addr string) error {
	r, err := getPower(mi.ctx, mi.c, addr)
	if err != nil {
		return fmt.Errorf("error getting power: %s", err)
	}
	si, err := getStaticInfo(mi.ctx, mi.c, mi.h, mi.lr, addr)
	if err != nil {
		log.Debugf("error getting static info: %s", err)
	}

	mi.lock.Lock()
	m, ok := mi.index.Miners[addr]
	if !ok {
		m = Info{}
		mi.index.Miners[addr] = m
	}
	m.Power = r
	m.Static = si
	mi.lock.Unlock()
	return nil
}

type Static struct {
	UserAgent string
	Location  Location
	Online    bool
}

type Location struct {
	Country   string
	Longitude float32
	Latitude  float32
}

func (mi *MinerIndex) Close() error {
	mi.lock.Lock()
	defer mi.lock.Unlock()
	if mi.closed {
		return nil
	}
	mi.cancel()
	<-mi.finished
	mi.closed = true
	return nil
}

func getPower(ctx context.Context, c API, addr string) (Power, error) {
	mp, err := c.StateMinerPower(ctx, addr, nil)
	if err != nil {
		return Power{}, err
	}
	p := mp.MinerPower.Uint64()
	tp := mp.TotalPower
	return Power{
		Power:    p,
		Relative: float64(p) / float64(tp.Uint64()),
	}, nil
}

func getStaticInfo(ctx context.Context, c API, h *fchost.FilecoinHost, lr iplocation.LocationResolver, addr string) (Static, error) {
	si := Static{}
	pid, err := c.StateMinerPeerID(ctx, addr, nil)
	if err != nil {
		return si, err
	}
	_, err = h.NewStream(ctx, pid, identify.ID)
	if err != nil {
		return si, err
	}
	si.Online = true

	if v, err := h.Peerstore().Get(pid, "AgentVersion"); err == nil {
		agent, ok := v.(string)
		if ok {
			si.UserAgent = agent
		}
	}
	if l, err := lr.Resolve(h.Peerstore().Addrs(pid)); err == nil {
		si.Location = Location{
			Country:   l.Country,
			Latitude:  l.Latitude,
			Longitude: l.Longitude,
		}
	}
	return si, nil
}
