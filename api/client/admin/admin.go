package admin

import (
	proto "github.com/textileio/powergate/proto/admin/v1"
)

// Admin provides access to Powergate admin APIs.
type Admin struct {
	StorageJobs *StorageJobs
	Profiles    *Profiles
	Wallet      *Wallet
}

// NewAdmin creates a new admin API.
func NewAdmin(client proto.PowergateAdminServiceClient) *Admin {
	return &Admin{
		StorageJobs: &StorageJobs{client: client},
		Profiles:    &Profiles{client: client},
		Wallet:      &Wallet{client: client},
	}
}
