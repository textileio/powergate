package api

import (
	"errors"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

var (
	// ErrNotFound returned when instance configuration doesn't exist.
	ErrNotFound = errors.New("stored item not found")
)

// Config has general information about a Api instance.
// ToDo: rename
type Config struct {
	ID                   ffs.APIID
	Addrs                map[string]AddrInfo
	DefaultStorageConfig ffs.StorageConfig
}

// AddrInfo provides information about a wallet address.
type AddrInfo struct {
	Name string
	Addr string
	Type string
}

// InstanceInfo has general information about a running Api instance.
type InstanceInfo struct {
	ID                   ffs.APIID
	DefaultStorageConfig ffs.StorageConfig
	Balances             []BalanceInfo
	Pins                 []cid.Cid
}

// BalanceInfo contains the balance for the associated wallet address.
type BalanceInfo struct {
	AddrInfo
	Balance uint64
}

// NewAddressConfig contains options for creating a new wallet address.
type NewAddressConfig struct {
	makeDefault bool
	addressType string
}

// GetLogsConfig contains configuration for a stream-log
// of human-friendly messages for a Cid execution.
type GetLogsConfig struct {
	jid     ffs.JobID
	history bool
}
