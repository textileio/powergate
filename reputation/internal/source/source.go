package source

import (
	"context"
	"time"

	logging "github.com/ipfs/go-log"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	log = logging.Logger("source")
)

type Source struct {
	Id          string
	Maddr       ma.Multiaddr
	LastFetched *time.Time
}

func (s *Source) Refresh(ctx context.Context) error {
	// ToDo: pull from Maddr fresh reputation information
	return nil
}
