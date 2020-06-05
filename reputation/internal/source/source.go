package source

import (
	"context"
	"time"

	ma "github.com/multiformats/go-multiaddr"
)

// Source is an external source of reputation information.
type Source struct {
	ID          string
	Weight      float64
	Scores      map[string]int
	Maddr       ma.Multiaddr
	LastFetched *time.Time
}

// Refresh pulls fresh information from source.
func (s *Source) Refresh(ctx context.Context) error {
	// ToDo: pull from Maddr fresh reputation information
	return nil
}
