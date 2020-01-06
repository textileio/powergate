package types

import (
	"github.com/ipfs/go-cid"
)

type Ticket struct {
	VRFProof []byte
}

type EPostTicket struct {
	Partial        []byte
	SectorID       uint64
	ChallengeIndex uint64
}

type EPostProof struct {
	Proof      []byte
	PostRand   []byte
	Candidates []EPostTicket
}

type BlockHeader struct {
	Miner string

	Ticket *Ticket

	EPostProof EPostProof

	Parents []cid.Cid

	ParentWeight BigInt

	Height uint64

	ParentStateRoot cid.Cid

	ParentMessageReceipts cid.Cid

	Messages cid.Cid

	BLSAggregate Signature

	Timestamp uint64

	BlockSig *Signature
}
