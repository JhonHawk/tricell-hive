package viewer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func timelineImageExtension(media string) string {
	switch media {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	}
	return ""
}
func (s *Store) timelineWriteBlob(id, sub, name string, content []byte) error {
	if err := s.timelineEnsureDir(s.timelineDir()); err != nil {
		return err
	}
	if err := s.timelineEnsureDir(s.timelineBlobDir(id)); err != nil {
		return err
	}
	dir := filepath.Join(s.timelineBlobDir(id), sub)
	if err := s.timelineEnsureDir(dir); err != nil {
		return err
	}
	return s.timelineDurableWrite(filepath.Join(dir, name), content)
}
func (s *Store) TimelineAddArtifact(id string, request TimelineArtifactMetadata, content []byte) (TimelineOperationResult, error) {
	if err := request.Validate(); err != nil {
		return TimelineOperationResult{}, err
	}
	if len(content) < 1 || len(content) > maxCapture || http.DetectContentType(content) != request.MediaType {
		return TimelineOperationResult{}, timelineFailure("invalid_image", 400)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.timelineGetLocked(id)
	if err != nil {
		return TimelineOperationResult{}, err
	}
	digest := timelineDigest("artifact_add", id, "", request, content)
	if result, done, err := s.timelineReplay(old, request.OperationID, digest); done || err != nil {
		return result, err
	}
	if len(old.Receipts) >= TimelineContentReceipts {
		return TimelineOperationResult{}, timelineFailure("capacity", 413)
	}
	if old.Session.State == TimelineClosed || old.Session.State == "failed" {
		return TimelineOperationResult{}, timelineFailure("session_closed", 409)
	}
	if err := ValidateTimelineAuthority(*old.Session, "artifact_add", request.Actor, ""); err != nil {
		return TimelineOperationResult{}, err
	}
	r := timelineClone(old)
	hash := sha256.Sum256(content)
	sha := hex.EncodeToString(hash[:])
	created := false
	aid := ""
	for _, a := range r.Session.Artifacts {
		if a.SHA256 == sha {
			aid = a.ArtifactID
			break
		}
	}
	if aid == "" {
		aid, err = timelineRandomID("artifact-")
		if err != nil {
			return TimelineOperationResult{}, err
		}
		created = true

		r.Session.Artifacts = append(r.Session.Artifacts, TimelineArtifact{Actor: request.Actor, ArtifactID: aid, SessionID: id, SHA256: sha, Size: int64(len(content)), MediaType: request.MediaType, ReceivedAt: s.timelineTime(), CapturedAt: request.CapturedAt, Candidate: request.Candidate, Viewport: request.Viewport, Theme: request.Theme})
		raw, _ := json.Marshal(r.Session)
		if len(raw) > TimelineContentBudget {
			return TimelineOperationResult{}, timelineFailure("capacity", 413)
		}
		if err = s.timelineWriteBlob(id, "artifacts", aid+timelineImageExtension(request.MediaType), content); err != nil {
			return TimelineOperationResult{}, err
		}
	}
	if err = s.timelineAudit(r, "artifact_add", request.OperationID, aid, &request.Actor, nil, "Immutable content upload"); err != nil {
		return TimelineOperationResult{}, err
	}
	if err = timelineBump(r.Session, s.timelineTime()); err != nil {
		return TimelineOperationResult{}, err
	}
	result := timelineResult(r, "artifact_add", request.OperationID)
	result.ArtifactID = aid
	result.Created = &created
	return s.timelineCommit(r, "artifact_add", request.OperationID, digest, result, false)
}
func (s *Store) TimelineAddDocument(id string, request TimelineDocumentMetadata, content []byte) (TimelineOperationResult, error) {
	if err := request.Validate(); err != nil {
		return TimelineOperationResult{}, err
	}
	if err := ValidateTimelineDocumentBytes(content); err != nil {
		return TimelineOperationResult{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.timelineGetLocked(id)
	if err != nil {
		return TimelineOperationResult{}, err
	}
	digest := timelineDigest("document_add", id, "", request, content)
	if result, done, err := s.timelineReplay(old, request.OperationID, digest); done || err != nil {
		return result, err
	}
	if len(old.Receipts) >= TimelineContentReceipts {
		return TimelineOperationResult{}, timelineFailure("capacity", 413)
	}
	if err = ValidateTimelineDocumentUpload(*old.Session, request, content); err != nil {
		return TimelineOperationResult{}, err
	}
	r := timelineClone(old)
	hash := sha256.Sum256(content)
	sha := hex.EncodeToString(hash[:])
	created := false
	did := ""
	for _, d := range r.Session.Documents {
		if d.SHA256 == sha {
			did = d.DocumentID
			break
		}
	}
	if did == "" {
		did, err = timelineRandomID("document-")
		if err != nil {
			return TimelineOperationResult{}, err
		}
		created = true

		r.Session.Documents = append(r.Session.Documents, TimelineDocument{Actor: request.Actor, DocumentID: did, SessionID: id, SHA256: sha, Size: int64(len(content)), MediaType: "text/markdown", ReceivedAt: s.timelineTime()})
		raw, _ := json.Marshal(r.Session)
		if len(raw) > TimelineContentBudget {
			return TimelineOperationResult{}, timelineFailure("capacity", 413)
		}
		if err = s.timelineWriteBlob(id, "documents", did+".md", content); err != nil {
			return TimelineOperationResult{}, err
		}
	}
	if err = s.timelineAudit(r, "document_add", request.OperationID, did, &request.Actor, nil, "Immutable content upload"); err != nil {
		return TimelineOperationResult{}, err
	}
	if err = timelineBump(r.Session, s.timelineTime()); err != nil {
		return TimelineOperationResult{}, err
	}
	result := timelineResult(r, "document_add", request.OperationID)
	size := int64(len(content))
	result.DocumentID = did
	result.SHA256 = sha
	result.Size = &size
	result.Created = &created
	return s.timelineCommit(r, "document_add", request.OperationID, digest, result, false)
}
func (s *Store) timelineReadBlob(id, sub, name, sha string, size int64) ([]byte, error) {
	for _, dir := range []string{s.timelineDir(), s.timelineBlobDir(id), filepath.Join(s.timelineBlobDir(id), sub)} {
		info, err := os.Lstat(dir)
		if err != nil || !info.IsDir() {
			return nil, timelineFailure("invalid_storage", 500)
		}
	}
	root, err := os.OpenRoot(s.timelineBlobDir(id))
	if err != nil {
		return nil, err
	}
	defer root.Close()
	rel := filepath.Join(sub, name)
	info, err := root.Lstat(rel)
	if err != nil || !info.Mode().IsRegular() || info.Size() != size {
		return nil, timelineFailure("corrupt_artifact", 500)
	}
	file, err := root.Open(rel)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) || validateSingleLink(opened) != nil {
		return nil, timelineFailure("invalid_storage", 500)
	}
	content, err := io.ReadAll(io.LimitReader(file, size+1))
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(content)
	if int64(len(content)) != size || hex.EncodeToString(hash[:]) != sha {
		return nil, timelineFailure("corrupt_artifact", 500)
	}
	return content, nil
}
func (s *Store) TimelineDocument(id, did string) (TimelineDocumentResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.timelineGetLocked(id)
	if err != nil {
		return TimelineDocumentResponse{}, err
	}
	if r.Session == nil || r.DeletePending {
		return TimelineDocumentResponse{}, timelineFailure("session_deleted", 410)
	}
	for _, d := range r.Session.Documents {
		if d.DocumentID == did {
			content, err := s.timelineReadBlob(id, "documents", did+".md", d.SHA256, d.Size)
			if err != nil {
				return TimelineDocumentResponse{}, err
			}
			return TimelineDocumentResponse{SchemaVersion: 2, Document: d, Content: string(content)}, nil
		}
	}
	return TimelineDocumentResponse{}, ErrNotFound
}

// TimelineArtifact returns only a recorded immutable image. Relative filenames
// never select arbitrary content, and every read verifies committed bytes.
func (s *Store) TimelineArtifact(id, rel string) ([]byte, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.timelineGetLocked(id)
	if err != nil {
		return nil, "", err
	}
	if r.Session == nil || r.DeletePending {
		return nil, "", timelineFailure("session_deleted", 410)
	}
	for _, a := range r.Session.Artifacts {
		name := a.ArtifactID + timelineImageExtension(a.MediaType)
		if rel == "artifacts/"+name {
			raw, err := s.timelineReadBlob(id, "artifacts", name, a.SHA256, a.Size)
			return raw, a.MediaType, err
		}
	}
	return nil, "", ErrNotFound
}
func (s *Store) TimelineDelete(id string, request TimelineMutationRequest) (TimelineOperationResult, error) {
	if err := request.Validate(); err != nil {
		return TimelineOperationResult{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.timelineGetLocked(id)
	if err != nil {
		return TimelineOperationResult{}, err
	}
	digest := timelineDigest("session_delete", id, "", request, nil)
	for _, receipt := range old.Receipts {
		if receipt.OperationID == request.OperationID {
			if ValidateTimelineOperationRetry(receipt, request.OperationID, digest) != nil {
				return TimelineOperationResult{}, timelineFailure("operation_conflict", 409)
			}
			return receipt.Result, nil
		}
	}
	if old.DeletePending {
		if old.DeleteOperation != request.OperationID || old.DeleteDigest != digest {
			return TimelineOperationResult{}, timelineFailure("operation_conflict", 409)
		}
		return s.timelineFinishDelete(timelineClone(old))
	}
	if old.Session == nil {
		return TimelineOperationResult{}, timelineFailure("session_deleted", 410)
	}
	if err = ValidateTimelineAuthority(*old.Session, "session_delete", request.Actor, ""); err != nil {
		return TimelineOperationResult{}, err
	}
	if old.Session.Version != request.ExpectedVersion {
		return TimelineOperationResult{}, timelineFailure("version_conflict", 409)
	}
	r := timelineClone(old)
	yes := true
	result := timelineResult(r, "session_delete", request.OperationID)
	result.Deleted = &yes
	result.CurrentStageEventID = nil
	if result.SessionVersion < TimelineMaxNumber {
		result.SessionVersion++
	}
	if err = s.timelineAudit(r, "session_delete", request.OperationID, id, &request.Actor, nil, "Confirmed local viewer deletion"); err != nil {
		return TimelineOperationResult{}, err
	}
	r.DeletePending = true
	r.DeleteOperation = request.OperationID
	r.DeleteDigest = digest
	r.DeleteResult = &result
	// The durable pending marker needs no content to identify its blob namespace.
	// Drop the public payload before persistence so deletion remains available
	// at content capacity and interrupted cleanup cannot expose deleted content.
	r.Session = nil
	if err = s.timelineSave(r); err != nil {
		return TimelineOperationResult{}, err
	}
	return s.timelineFinishDelete(r)
}
func (s *Store) timelineFinishDelete(r *timelineRecord) (TimelineOperationResult, error) {
	if err := s.timelineFaultAt("before_delete_blobs", r.ID); err != nil {
		return TimelineOperationResult{}, s.timelinePendingDeleteError(r.ID)
	}
	if err := s.removeTree(s.timelineBlobDir(r.ID)); err != nil {
		return TimelineOperationResult{}, s.timelinePendingDeleteError(r.ID)
	}
	if err := s.timelineSyncDir(s.timelineDir()); err != nil {
		return TimelineOperationResult{}, s.timelinePendingDeleteError(r.ID)
	}
	result := *r.DeleteResult
	r.Session = nil
	r.DeletePending = false
	return s.timelineCommit(r, "session_delete", r.DeleteOperation, r.DeleteDigest, result, true)
}

// timelineCleanup removes only task-owned incomplete writes and unreferenced
// blobs proven absent from their authoritative record. It never ages content.
func (s *Store) timelineCleanup() error {
	entries, err := os.ReadDir(s.timelineDir())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".timeline-write-") && !e.IsDir() {
			if err = os.Remove(filepath.Join(s.timelineDir(), e.Name())); err != nil {
				return err
			}
		}
	}
	for id, r := range s.timeline {
		if r.Session == nil || r.DeletePending {
			continue
		}
		allowed := map[string]bool{}
		for _, a := range r.Session.Artifacts {
			allowed["artifacts/"+a.ArtifactID+timelineImageExtension(a.MediaType)] = true
		}
		for _, d := range r.Session.Documents {
			allowed["documents/"+d.DocumentID+".md"] = true
		}
		for _, sub := range []string{"artifacts", "documents"} {
			dir := filepath.Join(s.timelineBlobDir(id), sub)
			files, readErr := os.ReadDir(dir)
			if errors.Is(readErr, os.ErrNotExist) {
				continue
			}
			if readErr != nil {
				return readErr
			}
			for _, f := range files {
				if f.IsDir() {
					return timelineFailure("invalid_storage", 500)
				}
				if !allowed[sub+"/"+f.Name()] && (strings.HasPrefix(f.Name(), ".timeline-write-") || timelineOrphanName(f.Name(), sub)) {
					if err = os.Remove(filepath.Join(dir, f.Name())); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func timelineOrphanName(name, sub string) bool {
	prefix := "artifact-"
	extensions := []string{".png", ".jpg", ".webp"}
	if sub == "documents" {
		prefix = "document-"
		extensions = []string{".md"}
	}
	for _, ext := range extensions {
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ext) {
			raw := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ext)
			b, err := hex.DecodeString(raw)
			return err == nil && len(b) == 16
		}
	}
	return false
}

func (s *Store) timelinePendingDeleteError(id string) error {
	count := 0
	err := filepath.WalkDir(s.timelineBlobDir(id), func(_ string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			count++
		}
		return nil
	})
	failure := &TimelineError{Code: "delete_pending", Status: 503, CleanupPending: true}
	if err == nil || errors.Is(err, os.ErrNotExist) {
		failure.RemainingFiles = &count
	}
	return failure
}
