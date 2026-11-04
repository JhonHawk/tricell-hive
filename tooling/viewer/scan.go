package viewer

import (
	"context"
	"errors"
	"time"
)

// RunOptions configures Run.
type RunOptions struct {
	// PollInterval is the maintenance interval (default and maximum 60 s). PurgeInterval is
	// a deprecated compatibility field and has no effect.
	PollInterval, PurgeInterval time.Duration
	// OnError, when set, receives maintenance failures. They never
	// stop Run.
	OnError func(error)
}

// Run validates stored work, catches up archival, then maintains it until cancellation.
// API publication is the only ingestion path; no manifest scanner runs.
func (s *Store) Run(ctx context.Context, o RunOptions) error {
	if err := s.startupInventory(); err != nil {
		return err
	}
	s.mu.Lock()
	pending := s.timelineStartupMaintenanceError
	s.timelineStartupMaintenanceError = nil
	err := errors.Join(pending, s.timelineMaintainLocked())
	s.mu.Unlock()
	if err != nil && o.OnError != nil {
		o.OnError(err)
	}
	interval := o.PollInterval
	if interval <= 0 || interval > time.Minute {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.mu.Lock()
			err := s.timelineMaintainLocked()
			s.mu.Unlock()
			if err != nil && o.OnError != nil {
				o.OnError(err)
			}
		}
	}
}
