package fastapi

import (
	"time"

	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
)

var (
	EmptyID = ID("")
)

type ID string

func (i ID) Valid() bool {
	_, err := uuid.Parse(string(i))
	return err == nil
}
func (i ID) String() string {
	return string(i)
}

func NewID() ID {
	return ID(uuid.New().String())
}

type info struct {
	ID         ID
	WalletAddr string
}

type CidInfo struct {
	Cid     cid.Cid
	Created time.Time
	Hot     HotInfo
	Cold    ColdInfo
}

type HotInfo struct {
	Size int
	Ipfs IpfsHotInfo
}

type IpfsHotInfo struct {
	Created time.Time
}

type ColdInfo struct {
	Filecoin FilInfo
}

type FilInfo struct {
	PayloadCID cid.Cid
	Duration   uint64
	Proposals  []FilStorage
}

type FilStorage struct {
	ProposalCid cid.Cid
	Failed      bool
}

type Info struct {
	ID     ID
	Wallet WalletInfo
	Pins   []cid.Cid
}

type WalletInfo struct {
	Address string
	Balance uint64
}
