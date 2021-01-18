package api

import (
	"errors"

	"github.com/textileio/powergate/v2/ffs"
)

var (
	// ErrNotFound returned when instance configuration doesn't exist.
	ErrNotFound = errors.New("stored item not found")
)

// InstanceConfig has general information about a Api instance.
type InstanceConfig struct {
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
