package admin

import (
	proto "github.com/textileio/powergate/proto/admin/v1"
)

// Admin provides access to Powergate admin APIs.
type Admin struct {
	Jobs     *Jobs
	Profiles *Profiles
	Wallet   *Wallet
}

func NewAdmin(client proto.PowergateAdminServiceClient) *Admin {
	return &Admin{
		Jobs:     &Jobs{client: client},
		Profiles: &Profiles{client: client},
		Wallet:   &Wallet{client: client},
	}
}
