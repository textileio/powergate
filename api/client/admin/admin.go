package admin

import (
	adminPb "github.com/textileio/powergate/api/gen/powergate/admin/v1"
)

// Admin provides access to Powergate admin APIs.
type Admin struct {
	StorageJobs *StorageJobs
	Users       *Users
	Wallet      *Wallet
	Data        *Data
}

// NewAdmin creates a new admin API.
func NewAdmin(client adminPb.AdminServiceClient) *Admin {
	return &Admin{
		StorageJobs: &StorageJobs{client: client},
		Users:       &Users{client: client},
		Wallet:      &Wallet{client: client},
		Data:        &Data{client: client},
	}
}
