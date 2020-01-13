package source

import (
	"context"
	"time"

	ma "github.com/multiformats/go-multiaddr"
)

type Source struct {
	Id          string
	Weight      float64
	Scores      map[string]int
	Maddr       ma.Multiaddr
	LastFetched *time.Time
}

func (s *Source) Refresh(ctx context.Context) error {
	// ToDo: pull from Maddr fresh reputation information
	return nil
}
