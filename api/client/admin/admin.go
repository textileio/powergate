package admin

import (
	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
)

// Admin provides access to Powergate admin APIs.
type Admin struct {
	StorageJobs *StorageJobs
	StorageInfo *StorageInfo
	Users       *Users
	Wallet      *Wallet
	Data        *Data
	Records     *Records
}

// NewAdmin creates a new admin API.
func NewAdmin(client adminPb.AdminServiceClient) *Admin {
	return &Admin{
		StorageJobs: &StorageJobs{client: client},
		StorageInfo: &StorageInfo{client: client},
		Users:       &Users{client: client},
		Wallet:      &Wallet{client: client},
		Data:        &Data{client: client},
		Records:     &Records{client: client},
	}
}
