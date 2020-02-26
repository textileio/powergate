package types

import (
	"context"
)

type WalletManager interface {
	NewWallet(ctx context.Context, typ string) (string, error)
	Balance(ctx context.Context, addr string) (uint64, error)
}

type MinerProposal struct {
	Addr       string
	EpochPrice uint64
}

type MinerSelector interface {
	GetTopMiners(n int) ([]MinerProposal, error)
}

type Auditor interface {
	Start(ctx context.Context, instanceID string) OpAuditor
}
type OpAuditor interface {
	ID() string
	Success()
	Errored(error)
	Log(event interface{})
	Close()
}
