package sendstore

import (
	"math/big"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
	"github.com/textileio/powergate/wallet"
)

var createAddress = address.NewForTestGetter()

// func createTxn(t *testing.T) wallet.SendFilTxn {
// 	c, err := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u")
// 	require.NoError(t, err)

// 	return w
// }

func TestPut(t *testing.T) {
	t.Parallel()
	s := create(t)
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u", address.TestAddress, address.TestAddress2, "100")
}

func TestGet(t *testing.T) {
	t.Parallel()
	s := create(t)
	txn := requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u", address.TestAddress, address.TestAddress2, "100")
	res, err := s.Get(txn.Cid)
	require.NoError(t, err)
	require.Equal(t, txn.Cid, res.Cid)
	require.Equal(t, txn.Amount, res.Amount)
	require.Equal(t, txn.From, res.From)
	require.Equal(t, txn.To, res.To)
	require.True(t, res.Time.Equal(txn.Time))
}

func TestAllFrom(t *testing.T) {
	t.Parallel()
	s := create(t)
	addr1 := createAddress()
	addr2 := createAddress()
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u", addr1, addr2, "100")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8v", addr1, addr2, "200")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8w", addr1, addr2, "300")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8x", addr2, addr1, "400")
	all, err := s.From(addr1)
	require.NoError(t, err)
	require.Len(t, all, 3)
}

func TestAllTo(t *testing.T) {
	t.Parallel()
	s := create(t)
	addr1 := createAddress()
	addr2 := createAddress()
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u", addr1, addr2, "100")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8v", addr1, addr2, "200")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8w", addr1, addr2, "300")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8x", addr2, addr1, "400")
	all, err := s.To(addr2)
	require.NoError(t, err)
	require.Len(t, all, 3)
}

func TestAllFromTo(t *testing.T) {
	t.Parallel()
	s := create(t)
	addr1 := createAddress()
	addr2 := createAddress()
	addr3 := createAddress()
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u", addr1, addr2, "100")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8v", addr1, addr2, "200")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8w", addr1, addr2, "300")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8x", addr2, addr1, "400")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8y", addr3, addr1, "500")
	all, err := s.FromTo(addr1, addr2)
	require.NoError(t, err)
	require.Len(t, all, 3)
	all, err = s.FromTo(addr2, addr1)
	require.NoError(t, err)
	require.Len(t, all, 1)
}

func TestBetween(t *testing.T) {
	t.Parallel()
	s := create(t)
	addr1 := createAddress()
	addr2 := createAddress()
	addr3 := createAddress()
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u", addr1, addr2, "100")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8v", addr1, addr2, "200")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8w", addr1, addr2, "300")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8x", addr2, addr1, "400")
	requirePut(t, s, "QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8y", addr3, addr1, "500")
	all, err := s.Between(addr1, addr2)
	require.NoError(t, err)
	require.Len(t, all, 4)
}

func requireCid(t *testing.T, cid string) cid.Cid {
	c, err := util.CidFromString(cid)
	require.NoError(t, err)
	return c
}

func requireBigInt(t *testing.T, s string) *big.Int {
	res, ok := new(big.Int).SetString(s, 10)
	require.True(t, ok)
	return res
}

func requirePut(t *testing.T, s *SendStore, cid string, from, to address.Address, amt string) *wallet.SendFilEvent {
	c := requireCid(t, cid)
	a := requireBigInt(t, amt)
	txn, err := s.Put(c, from, to, a)
	require.NoError(t, err)
	require.Equal(t, c, txn.Cid)
	require.Equal(t, a, txn.Amount)
	require.Equal(t, from, txn.From)
	require.Equal(t, to, txn.To)
	require.True(t, txn.Time.Before(time.Now()))
	return txn
}

func create(t *testing.T) *SendStore {
	ds := tests.NewTxMapDatastore()
	store := New(ds)
	require.NotNil(t, store)
	return store
}
