package viewer

import (
	"context"
	"time"
)

// RunOptions configures Run.
type RunOptions struct {
	// PollInterval is the time between scans (default 5 s). PurgeInterval is
	// a deprecated compatibility field and has no effect.
	PollInterval, PurgeInterval time.Duration
	// OnError, when set, receives failures of a scan or purge. They never
	// stop Run.
	OnError func(error)
}

// Run validates and cleans owned v2 temporaries, then waits for cancellation.
// API publication is the only ingestion path; no manifest scanner runs.
func (s *Store) Run(ctx context.Context, o RunOptions) error {
	if err := s.Startup(); err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}
