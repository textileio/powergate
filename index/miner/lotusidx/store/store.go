package store

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-datastore"
	kt "github.com/ipfs/go-datastore/keytransform"
	"github.com/ipfs/go-datastore/query"
	logger "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/v2/index/miner"
)

var (
	dsMetadata      = datastore.NewKey("meta")
	dsOnChainMiner  = datastore.NewKey("onchain/miner")
	dsOnChainHeight = datastore.NewKey("onchain/height")

	log = logger.Logger("lotusidx-store")
)

// Store is a store to save on-chain and metadata information about the chain.
type Store struct {
	ds *kt.Datastore
}

// New returns a new *Store.
func New(ds *kt.Datastore) (*Store, error) {
	return &Store{
		ds: ds,
	}, nil
}

// SaveMetadata creates/updates metadata information of miners.
func (s *Store) SaveMetadata(index miner.MetaIndex) error {
	for addr, meta := range index.Info {
		key := makeMinerMetadataKey(addr)
		buf, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("marshaling metadata for miner %s: %s", addr, err)
		}
		if err := s.ds.Put(key, buf); err != nil {
			return fmt.Errorf("saving metadata in store: %s", err)
		}
	}
	return nil
}

// SaveOnChain creates/updates on-chain information of miners.
func (s *Store) SaveOnChain(ctx context.Context, index miner.ChainIndex) error {
	var i int64
	b, err := s.ds.Batch()
	if err != nil {
		return fmt.Errorf("creating batch: %s", err)
	}
	for addr, onchain := range index.Miners {
		i++
		if i%1000 == 0 {
			if err := b.Commit(); err != nil {
				return fmt.Errorf("committing batch: %s", err)
			}
			b, err = s.ds.Batch()
			if err != nil {
				return fmt.Errorf("creating batch: %s", err)
			}
		}
		if ctx.Err() != nil {
			return fmt.Errorf("context signal: %s", ctx.Err())
		}
		key := makeMinerOnChainKey(addr)
		buf, err := json.Marshal(onchain)
		if err != nil {
			return fmt.Errorf("marshaling onchain for miner %s: %s", addr, err)
		}
		if err := b.Put(key, buf); err != nil {
			return fmt.Errorf("saving onchain in store: %s", err)
		}
	}
	if err := b.Commit(); err != nil {
		return fmt.Errorf("committing batch: %s", err)
	}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(index.LastUpdated))
	if err := s.ds.Put(dsOnChainHeight, buf); err != nil {
		return fmt.Errorf("saving metadata in store: %s", err)
	}

	return nil
}

// GetIndex gets the complete stored metadata and on-chain index.
func (s *Store) GetIndex() (miner.IndexSnapshot, error) {
	metaIndex, err := s.getMetaIndex()
	if err != nil {
		return miner.IndexSnapshot{}, fmt.Errorf("get meta index: %s", err)
	}
	onChain, err := s.getOnChainIndex()
	if err != nil {
		return miner.IndexSnapshot{}, fmt.Errorf("get onchain index: %s", err)
	}

	return miner.IndexSnapshot{
		Meta:    metaIndex,
		OnChain: onChain,
	}, nil
}

func (s *Store) getMetaIndex() (miner.MetaIndex, error) {
	q := query.Query{Prefix: dsMetadata.String()}
	res, err := s.ds.Query(q)
	if err != nil {
		return miner.MetaIndex{}, fmt.Errorf("executing query: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing meta-index query: %s", err)
		}
	}()

	info := map[string]miner.Meta{}
	for v := range res.Next() {
		if v.Error != nil {
			return miner.MetaIndex{}, fmt.Errorf("fetching query result: %s", v.Error)
		}
		key := datastore.NewKey(v.Key)
		minerAddr := key.Namespaces()[1]
		var m miner.Meta
		if err := json.Unmarshal(v.Value, &m); err != nil {
			return miner.MetaIndex{}, fmt.Errorf("unmarshaling meta info: %s", err)
		}
		info[minerAddr] = m
	}

	return miner.MetaIndex{
		Info: info,
	}, nil
}

func (s *Store) getOnChainIndex() (miner.ChainIndex, error) {
	q := query.Query{Prefix: dsOnChainMiner.String()}
	res, err := s.ds.Query(q)
	if err != nil {
		return miner.ChainIndex{}, fmt.Errorf("executing query: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing on-chain index query: %s", err)
		}
	}()

	info := map[string]miner.OnChainMinerData{}
	for v := range res.Next() {
		if v.Error != nil {
			return miner.ChainIndex{}, fmt.Errorf("fetching query result: %s", v.Error)
		}
		key := datastore.NewKey(v.Key)
		minerAddr := key.Namespaces()[2]
		var m miner.OnChainMinerData
		if err := json.Unmarshal(v.Value, &m); err != nil {
			return miner.ChainIndex{}, fmt.Errorf("unmarshaling onchain info: %s", err)
		}
		info[minerAddr] = m
	}

	buf, err := s.ds.Get(dsOnChainHeight)
	if err != nil && err != datastore.ErrNotFound {
		return miner.ChainIndex{}, fmt.Errorf("get onchain height: %s", err)
	}
	var lastUpdated int64
	if err != datastore.ErrNotFound {
		lastUpdated = int64(binary.LittleEndian.Uint64(buf))
	}

	return miner.ChainIndex{
		Miners:      info,
		LastUpdated: lastUpdated,
	}, nil
}

func makeMinerMetadataKey(addr string) datastore.Key {
	return dsMetadata.ChildString(addr)
}

func makeMinerOnChainKey(addr string) datastore.Key {
	return dsOnChainMiner.ChildString(addr)
}
