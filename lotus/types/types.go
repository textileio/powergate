package types

import "github.com/ipfs/go-cid"

type DealState = uint64

const (
	DealUnknown  = DealState(iota)
	DealRejected // Provider didn't like the proposal
	DealAccepted // Proposal accepted, data moved
	DealStaged   // Data put into the sector
	DealSealing  // Data in process of being sealed

	DealFailed
	DealComplete

	// Internal

	DealError // deal failed with an unexpected error

	DealNoUpdate = DealUnknown
)

var DealStates = []string{
	"DealUnknown",
	"DealRejected",
	"DealAccepted",
	"DealStaged",
	"DealSealing",
	"DealFailed",
	"DealComplete",
	"DealError",
}

type DealInfo struct {
	ProposalCid cid.Cid
	State       DealState
	Provider    string

	PieceRef []byte // cid bytes
	Size     uint64

	PricePerEpoch BigInt
	Duration      uint64
}

type TipSet struct {
	Cids   []cid.Cid
	Height uint64
}

type SignedStorageAsk struct {
	Ask       *StorageAsk
	Signature *Signature
}

type StorageAsk struct {
	// Price per GiB / Epoch
	Price BigInt

	MinPieceSize uint64
	Miner        string
	Timestamp    uint64
	Expiry       uint64
	SeqNo        uint64
}

type Version struct {
	Version string
	// Seconds
	BlockDelay uint64
}

type HeadChange struct {
	Type string
	Val  *TipSet
}
