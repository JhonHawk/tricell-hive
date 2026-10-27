package viewer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PartialDeleteError reports the actual surviving copies after a delete attempt.
// Its cause is internal; the HTTP response exposes only counts and recovery state.
type PartialDeleteError struct {
	Deleted        int
	Remaining      int
	CleanupPending bool
	cause          error
}

func (e *PartialDeleteError) Error() string { return "session deletion incomplete" }
func (e *PartialDeleteError) Unwrap() error { return e.cause }

func groupMembers(d *indexData, id sessionIdentity) []Session {
	var members []Session
	for _, sess := range d.Sessions {
		if identityOf(sess).key() == id.key() {
			members = append(members, sess)
		}
	}
	return members
}

func (s *Store) groupHasUnknownMembers(d *indexData) bool {
	for _, problem := range d.Errors {
		if problem.Path == s.sessionsDir() || strings.HasPrefix(problem.Path, s.sessionsDir()+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// recoverPendingGroups removes only explicitly deleted groups with no surviving
// copies. With repair false it merely reports pending state, including on GET.
// The caller owns mu, or is Open before the store is shared.
func (s *Store) recoverPendingGroups(d *indexData, repair bool) {
	kept := d.Errors[:0:0]
	for _, problem := range d.Errors {
		if problem.Code != "delete_cleanup_pending" {
			kept = append(kept, problem)
		}
	}
	d.Errors = kept
	membership := map[string]int{}
	for _, sess := range d.Sessions {
		membership[identityOf(sess).key()]++
	}
	var cleanupErr error
	if repair {
		cleanupErr = s.cleanupDeletedCopies(d)
	}
	entries, err := os.ReadDir(s.groupsDir())
	if errors.Is(err, fs.ErrNotExist) {
		return
	}
	if err != nil {
		d.Errors = append(d.Errors, IngestError{Path: s.groupsDir(), Reason: "cannot read session preferences"})
		return
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), writePrefix) {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(s.groupsDir(), entry.Name())
		raw, err := readBounded(path, maxMetaBytes)
		var record sessionGroup
		if err == nil {
			err = json.Unmarshal(raw, &record)
		}
		if err == nil && record.Identity.key()+".json" != entry.Name() {
			err = errors.New("session preference identity mismatch")
		}
		if err == nil {
			record, err = s.readGroup(record.Identity)
		}
		if err != nil {
			reported := false
			for _, problem := range d.Errors {
				if problem.Path == path {
					reported = true
					break
				}
			}
			if !reported {
				d.Errors = append(d.Errors, IngestError{Path: path, Code: "session_preferences_error", Reason: "cannot read session preferences"})
			}
			continue
		}
		if !record.DeletePending || membership[record.Identity.key()] != 0 {
			continue
		}
		if repair && cleanupErr == nil && !s.groupHasUnknownMembers(d) {
			err = s.removeFile(path)
			if err == nil || errors.Is(err, fs.ErrNotExist) {
				continue
			}
		}
		d.Errors = append(d.Errors, IngestError{Path: path, Code: "delete_cleanup_pending", SessionGroup: record.Identity.key(), Reason: "copies deleted; session cleanup pending; automatic retry scheduled"})
	}
}

// cleanupDeletedCopies retries only renamed deletion folders with a committed
// tombstone. It never removes active copy directories or unconfirmed leftovers.
func (s *Store) cleanupDeletedCopies(d *indexData) error {
	entries, err := os.ReadDir(s.sessionsDir())
	if err != nil {
		return err
	}
	var first error
	for _, entry := range entries {
		key, ok := strings.CutPrefix(entry.Name(), tmpPrefix+"del-")
		if !ok || !keyPattern.MatchString(key) {
			continue
		}
		if _, confirmed := d.tombstone(key); !confirmed {
			continue
		}
		if err := s.removeTree(filepath.Join(s.sessionsDir(), entry.Name())); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// clearPendingGroup gates new ingestion so a completed deletion cannot lend its
// old preferences to a new delivery. Surviving copies retain their preferences.
func (s *Store) clearPendingGroup(d *indexData, g sessionGroup) error {
	if !g.DeletePending || len(groupMembers(d, g.Identity)) != 0 {
		return nil
	}
	if s.groupHasUnknownMembers(d) {
		return errors.New("session cleanup waits for unreadable copies")
	}
	if err := s.cleanupDeletedCopies(d); err != nil {
		return err
	}
	if err := s.removeFile(s.groupPath(g.Identity)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("session cleanup pending: %w", err)
	}
	return nil
}

// Delete removes every current member of the logical session located by key.
// All tombstones and the recovery marker are committed before any copy is removed.
func (s *Store) Delete(key string) error {
	if err := validKey(key); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.current().clone()
	i, ok := d.session(key)
	if !ok {
		return ErrNotFound
	}
	if err := s.reconcileGroups(d, true); err != nil {
		return err
	}
	if s.groupHasUnknownMembers(d) {
		return errors.New("cannot resolve session members while stored copies are unreadable")
	}
	id := identityOf(d.Sessions[i])
	members := groupMembers(d, id)
	g, err := s.readGroup(id)
	if err != nil {
		return err
	}
	for _, sess := range members {
		s.addTombstone(d, sess)
	}
	if err := s.commit(d); err != nil {
		return err
	}
	g.DeletePending = true
	if err := s.writeGroup(g); err != nil {
		return err
	}
	var first error
	for _, sess := range members {
		if err := s.removeCopy(sess.Key); err != nil && first == nil {
			first = err
		}
	}
	// Reload from actual disk, retaining the precommitted tombstones even if a
	// final index write fails. A restart performs the same reconstruction.
	actual := s.load()
	remaining := len(groupMembers(actual, id))
	if s.groupHasUnknownMembers(actual) && first == nil {
		first = errors.New("cannot verify surviving session copies")
	}
	if err := s.commit(actual); err != nil && first == nil {
		first = err
	}
	cleanupPending := remaining == 0
	if remaining == 0 && first == nil {
		if err := s.removeFile(s.groupPath(id)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			first = err
		} else {
			cleanupPending = false
		}
	}
	if first != nil || remaining != 0 {
		if first == nil {
			first = errors.New("session members remain")
		}
		s.recoverPendingGroups(actual, false)
		s.viewMu.Lock()
		s.idx = actual
		s.viewMu.Unlock()
		return &PartialDeleteError{Deleted: len(members) - remaining, Remaining: remaining, CleanupPending: cleanupPending, cause: first}
	}
	// Retire the pending diagnostic from the published index only after removal.
	if err := s.commit(actual); err != nil {
		return &PartialDeleteError{Deleted: len(members), Remaining: 0, CleanupPending: false, cause: err}
	}
	return nil
}
