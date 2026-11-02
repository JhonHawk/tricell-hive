package viewer

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const timelineStorageVersion = 3
const timelineMaxRecordBytes = 4 << 20

// TimelineError contains public, value-free failure semantics.
type TimelineError struct {
	Code           string
	Status         int
	RemainingFiles *int
	CleanupPending bool
}

func (e *TimelineError) Error() string { return e.Code }
func timelineFailure(code string, status int) error {
	return &TimelineError{Code: code, Status: status}
}

type timelineRecord struct {
	Audit          []timelinePrivateAudit `json:"audit,omitempty"`
	StorageVersion int                    `json:"storage_version"`
	ID             string                 `json:"id"`
	Source         TimelineSource         `json:"source"`
	ProjectKey     string                 `json:"project_key"`
	ProjectName    string                 `json:"project_name"`
	Roots          []string               `json:"roots"`
	// WorkKey is the effective key behind ID: "work:<key>" or
	// "operation:<id>". Records written before work keys carry none and keep
	// the original derivation.
	WorkKey         string                     `json:"work_key,omitempty"`
	Session         *TimelineSession           `json:"session,omitempty"`
	Receipts        []TimelineOperationReceipt `json:"receipts"`
	DeletePending   bool                       `json:"delete_pending,omitempty"`
	DeleteOperation string                     `json:"delete_operation,omitempty"`
	DeleteDigest    string                     `json:"delete_digest,omitempty"`
	DeleteResult    *TimelineOperationResult   `json:"delete_result,omitempty"`
}

func (s *Store) timelineDir() string { return filepath.Join(s.vdir, "timeline-api-v3") }
func (s *Store) timelineRecordPath(id string) string {
	return filepath.Join(s.timelineDir(), id+".json")
}
func (s *Store) timelineBlobDir(id string) string { return filepath.Join(s.timelineDir(), id) }

// timelineEffectiveWorkKey names the work a create request addresses.
func timelineEffectiveWorkKey(request TimelineCreateSessionRequest) string {
	if request.WorkKey != nil {
		return "work:" + *request.WorkKey
	}
	return "operation:" + request.OperationID
}

// timelineWorkKeyMatches reports whether a session's public work key agrees
// with the effective key its record holds.
func timelineWorkKeyMatches(effective string, public *string) bool {
	if value, ok := strings.CutPrefix(effective, "work:"); ok {
		return public != nil && *public == value
	}
	return public == nil
}

// timelineSessionKey derives the viewer session ID. An empty effective key is
// the derivation of records stored before work keys; stored IDs stay stable.
func timelineSessionKey(project string, source TimelineSource, effective string) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	parts := []any{"timeline-api-v3", project, source.Host, source.NativeSessionID}
	if kind, value, ok := strings.Cut(effective, ":"); ok {
		parts = append(parts, kind, value)
	}
	_ = enc.Encode(parts)
	raw := strings.TrimSuffix(b.String(), "\n")
	raw = strings.ReplaceAll(strings.ReplaceAll(raw, `\u2028`, "\u2028"), `\u2029`, "\u2029")
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
func timelineRandomID(prefix string) (string, error) {
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		return "", e
	}
	return prefix + hex.EncodeToString(b[:]), nil
}
func timelineDigest(kind, id, event string, request any, content []byte) string {
	raw, _ := json.Marshal(request)
	h := sha256.New()
	for _, v := range []string{kind, id, event, string(raw)} {
		_, _ = io.WriteString(h, v)
		h.Write([]byte{0})
	}
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}
func timelineClone(r *timelineRecord) *timelineRecord {
	raw, _ := json.Marshal(r)
	var copy timelineRecord
	_ = json.Unmarshal(raw, &copy)
	return &copy
}
func (s *Store) timelineTime() string { return s.now().UTC().Format(time.RFC3339Nano) }
func (s *Store) timelineFaultAt(phase, path string) error {
	if s.timelineFault != nil {
		return s.timelineFault(phase, path)
	}
	return nil
}
func (s *Store) timelineSyncDir(path string) error {
	if err := s.timelineFaultAt("before_directory_sync", path); err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err = f.Sync(); err != nil {
		return err
	}
	return s.timelineFaultAt("after_directory_sync", path)
}

// timelineEnsureDir re-synchronizes the directory entry even when it already
// exists: a preceding mkdir may have succeeded before its parent sync failed.
func (s *Store) timelineEnsureDir(path string) error {
	if path == s.timelineDir() {
		// Open may have made viewer and state ancestors through MkdirAll. No
		// timeline success may assume those entries were already durable.
		ancestor, err := filepath.EvalSymlinks(s.vdir)
		if err != nil {
			return err
		}
		for {
			parent := filepath.Dir(ancestor)
			if parent == ancestor {
				break
			}
			if err = s.timelineSyncDir(parent); err != nil {
				return err
			}
			ancestor = parent
		}
	}
	if info, err := os.Lstat(path); err == nil {
		if !info.IsDir() {
			return timelineFailure("invalid_storage", 500)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	} else if err = os.Mkdir(path, 0o700); err != nil {
		return err
	}
	return s.timelineSyncDir(filepath.Dir(path))
}

// timelineDurableWrite reports post-rename failures as uncertain. The next
// operation reconciles the same receipt and syncs its directory before replay.
func (s *Store) timelineDurableWrite(path string, raw []byte) error {
	if err := s.timelineFaultAt("before_file_create", path); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".timeline-write-")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0o600); err == nil {
		_, err = f.Write(raw)
	}
	if err == nil {
		err = s.timelineFaultAt("before_file_sync", path)
	}
	if err == nil {
		err = f.Sync()
	}
	if err == nil {
		err = s.timelineFaultAt("after_file_sync", path)
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = s.timelineFaultAt("before_rename", path); err != nil {
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		return err
	}
	if err = s.timelineFaultAt("after_rename", path); err == nil {
		err = s.timelineSyncDir(filepath.Dir(path))
	}
	if err != nil {
		s.timelineUncertain = true
		return timelineFailure("outcome_unknown", 503)
	}
	return nil
}
func (s *Store) timelineLoad() error {
	s.timeline = map[string]*timelineRecord{}
	if info, err := os.Lstat(s.timelineDir()); err == nil && !info.IsDir() {
		return timelineFailure("invalid_storage", 500)
	}
	entries, err := os.ReadDir(s.timelineDir())
	if errors.Is(err, os.ErrNotExist) {
		s.timelinePublish()
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		if !keyPattern.MatchString(id) {
			return timelineFailure("corrupt_storage", 500)
		}
		raw, err := readStateFile(s.timelineRecordPath(id), timelineMaxRecordBytes)
		if err != nil {
			return timelineFailure("corrupt_storage", 500)
		}
		var r timelineRecord
		if err = decodeTimelineStoredJSON(raw, &r, timelineMaxRecordBytes); err != nil {
			return timelineFailure("corrupt_storage", 500)
		}
		if r.StorageVersion != timelineStorageVersion || r.ID != id || len(r.Roots) == 0 || r.ProjectKey == "" || timelineSessionKey(r.ProjectKey, r.Source, r.WorkKey) != id {
			return timelineFailure("corrupt_storage", 500)
		}
		if err = timelineSource(r.Source); err != nil {
			return timelineFailure("corrupt_storage", 500)
		}
		roots := map[string]bool{}
		for _, root := range r.Roots {
			if !filepath.IsAbs(root) || len(root) > 4096 || roots[root] {
				return timelineFailure("corrupt_storage", 500)
			}
			roots[root] = true
		}
		if len(r.Roots) > 64 || !roots[r.Source.ProjectRoot] {
			return timelineFailure("corrupt_storage", 500)
		}
		if r.DeletePending || r.Session == nil {
			if r.DeleteResult == nil || !timelineOperationID.MatchString(r.DeleteOperation) || !timelineSHA256.MatchString(r.DeleteDigest) || r.DeleteResult.OperationID != r.DeleteOperation {
				return timelineFailure("corrupt_storage", 500)
			}
		}

		if r.Session != nil {
			if err = ValidateTimelineSession(*r.Session); err != nil || r.Session.ViewerSessionID != id || r.Session.Source != r.Source || r.Session.ProjectKey != r.ProjectKey {
				return timelineFailure("corrupt_storage", 500)
			}
			if !timelineWorkKeyMatches(r.WorkKey, r.Session.WorkKey) {
				return timelineFailure("corrupt_storage", 500)
			}
		}
		if len(r.Receipts) > TimelineMaxReceipts+1 {
			return timelineFailure("corrupt_storage", 500)
		}
		seen := map[string]bool{}
		for _, receipt := range r.Receipts {
			if receipt.Validate() != nil || receipt.Result.ViewerSessionID != id || seen[receipt.OperationID] {
				return timelineFailure("corrupt_storage", 500)
			}
			seen[receipt.OperationID] = true
		}
		if r.DeleteResult != nil && (r.DeleteResult.Validate() != nil || r.DeleteResult.ViewerSessionID != id || r.DeleteResult.Kind != "session_delete") {
			return timelineFailure("corrupt_storage", 500)
		}
		if r.Session == nil && r.DeleteResult == nil {
			return timelineFailure("corrupt_storage", 500)
		}
		if err = validateTimelinePrivateAudit(&r); err != nil {
			return err
		}
		s.timeline[id] = &r
	}
	// Resolve against the complete set, so durable aliases remain available
	// even when the original source metadata cannot currently be read.
	for _, r := range s.timeline {
		if r.Session == nil {
			continue
		}
		sources := []TimelineSource{r.Source}
		for _, root := range r.Roots {
			source := r.Source
			source.ProjectRoot = root
			sources = append(sources, source)
		}
		for _, p := range r.Session.Producers {
			sources = append(sources, p.Source)
		}
		for _, p := range r.Session.PendingJoinRequests {
			sources = append(sources, p.Source)
		}
		for _, source := range sources {
			key, err := s.timelineRestoredProject(r, source)
			if err != nil {
				return err
			}
			if err = ValidateTimelineResolvedProject(*r.Session, source, key); err != nil {
				return err
			}
		}
		// A durable pending delete may already have removed some blobs.
		if r.DeletePending {
			continue
		}
		for _, d := range r.Session.Documents {
			if _, err := s.timelineReadBlob(r.ID, "documents", d.DocumentID+".md", d.SHA256, d.Size); err != nil {
				return err
			}
		}
		for _, a := range r.Session.Artifacts {
			if _, err := s.timelineReadBlob(r.ID, "artifacts", a.ArtifactID+timelineImageExtension(a.MediaType), a.SHA256, a.Size); err != nil {
				return err
			}
		}
	}
	s.timelinePublish()
	return nil
}

// timelineRestoredProject is load-only: an unavailable original path cannot
// revoke a previously validated durable binding. New publications still use
// timelineProject and require a currently resolvable, registered source root.
func (s *Store) timelineRestoredProject(record *timelineRecord, source TimelineSource) (string, error) {
	_, key, _, err := s.timelineProject(source)
	if err == nil {
		return key, nil
	}
	var failure *TimelineError
	if !errors.As(err, &failure) || failure.Code != "project_unavailable" {
		return "", err
	}
	if _, statErr := os.Stat(source.ProjectRoot); !errors.Is(statErr, os.ErrNotExist) {
		return "", err
	}
	sourceBound := false
	for _, root := range record.Roots {
		sourceBound = sourceBound || root == source.ProjectRoot
	}
	if !sourceBound {
		return "", timelineFailure("project_identity_unavailable", 409)
	}
	registry, readErr := readRegistry(s.vdir)
	if readErr != nil {
		return "", readErr
	}
	canonicalBound := false
	for _, repo := range registry.Repositories {
		bound := false
		for _, root := range record.Roots {
			bound = bound || root == repo.Root
		}
		if !bound {
			continue
		}
		canonicalBound = true
		for _, other := range s.timeline {
			for _, root := range other.Roots {
				if root == repo.Root && other.ProjectKey != record.ProjectKey {
					return "", timelineFailure("project_identity_changed", 409)
				}
			}
		}
		if _, statErr := os.Stat(repo.Root); errors.Is(statErr, os.ErrNotExist) {
			continue
		}
		resolved := source
		resolved.ProjectRoot = repo.Root
		_, canonicalKey, _, resolveErr := s.timelineProject(resolved)
		if resolveErr != nil {
			return "", resolveErr
		}
		if canonicalKey != record.ProjectKey {
			return "", timelineFailure("project_identity_changed", 409)
		}
	}
	if !canonicalBound {
		return "", timelineFailure("project_identity_unavailable", 409)
	}
	return record.ProjectKey, nil
}

func (s *Store) timelineReconcile() error {
	if !s.timelineUncertain {
		return nil
	}
	if err := s.timelineSyncDir(s.timelineDir()); err != nil {
		return timelineFailure("outcome_unknown", 503)
	}
	if err := s.timelineLoad(); err != nil {
		return err
	}
	s.timelineUncertain = false
	return nil
}
func (s *Store) timelinePublish() {
	entries := []TimelineAPIIndexEntry{}
	for _, r := range s.timeline {
		if r.Session == nil || r.DeletePending {
			continue
		}
		v := r.Session
		entries = append(entries, TimelineAPIIndexEntry{Mode: "timeline_api", ViewerSessionID: v.ViewerSessionID, ProjectKey: v.ProjectKey, ProjectName: v.ProjectName, Source: v.Source, Version: v.Version, Title: v.Title, Summary: v.Summary, State: v.State, Archived: v.Archived, Pinned: v.Pinned, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, ValidationCount: v.ValidationCount, Tickets: append([]TimelineTicket{}, v.Tickets...), CurrentStageEventID: v.CurrentStageEventID})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ViewerSessionID < entries[j].ViewerSessionID })
	s.viewMu.Lock()
	s.timelineOwned = map[string]bool{}
	for id, r := range s.timeline {
		s.timelineOwned[id] = true
		// The native tuple stays owned whatever work key named the work.
		s.timelineOwned[timelineSessionKey(r.ProjectKey, r.Source, "")] = true
	}
	s.timelineIndex = entries
	s.viewMu.Unlock()
}
func (s *Store) TimelineIndex() []TimelineAPIIndexEntry {
	s.viewMu.RLock()
	defer s.viewMu.RUnlock()
	raw, _ := json.Marshal(s.timelineIndex)
	var result []TimelineAPIIndexEntry
	_ = json.Unmarshal(raw, &result)
	if result == nil {
		return []TimelineAPIIndexEntry{}
	}
	return result
}
func (s *Store) timelineOwns(project, host, native string) bool {
	return s.timelineHasID(timelineSessionKey(project, TimelineSource{Host: TimelineHost(host), NativeSessionID: native}, ""))
}
func (s *Store) timelineReplay(r *timelineRecord, op, digest string) (TimelineOperationResult, bool, error) {
	for _, receipt := range r.Receipts {
		if receipt.OperationID == op {
			if err := ValidateTimelineOperationRetry(receipt, op, digest); err != nil {
				return TimelineOperationResult{}, true, timelineFailure("operation_conflict", 409)
			}
			return receipt.Result, true, nil
		}
	}
	if r.DeletePending {
		return TimelineOperationResult{}, false, timelineFailure("delete_pending", 409)
	}
	if r.Session == nil {
		return TimelineOperationResult{}, false, timelineFailure("session_deleted", 410)
	}
	return TimelineOperationResult{}, false, nil
}
func (s *Store) timelineGetLocked(id string) (*timelineRecord, error) {
	if !keyPattern.MatchString(id) {
		return nil, ErrInvalidKey
	}
	if err := s.timelineReconcile(); err != nil {
		return nil, err
	}
	r, ok := s.timeline[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}
func (s *Store) timelineSave(r *timelineRecord) error {
	if err := validateTimelinePrivateAudit(r); err != nil {
		return err
	}
	if r.Session != nil {
		if err := ValidateTimelineSession(*r.Session); err != nil {
			return err
		}
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if len(raw) > timelineMaxRecordBytes {
		return timelineFailure("capacity", 413)
	}
	if err = s.timelineEnsureDir(s.timelineDir()); err != nil {
		return err
	}
	if err = s.timelineDurableWrite(s.timelineRecordPath(r.ID), raw); err != nil {
		return err
	}
	s.timeline[r.ID] = timelineClone(r)
	s.timelinePublish()
	return nil
}
func (s *Store) timelineCommit(r *timelineRecord, kind TimelineOperationKind, op, digest string, result TimelineOperationResult, terminal bool) (TimelineOperationResult, error) {
	if r.Session != nil {
		if err := ValidateTimelineClosureReserve(*r.Session); err != nil {
			return result, err
		}
	}
	lifecycle := terminal || contains([]string{"session_patch", "event_patch", "event_withdraw", "producer_accept", "foreground_transfer", "recover_lead", "task_transition"}, string(kind))
	if kind == "append_event" && r.Session != nil {
		for _, event := range r.Session.Events {
			if event.EventID == result.EventID && (event.Kind == "closure" || event.Kind == "stage") {
				lifecycle = true
			}
		}
	}
	if !terminal && r.Session != nil && r.Session.State != TimelineClosed {
		// Preserve the operation receipts needed to release foreground/reopen,
		// then close. Exact retries replay their existing receipt before this path.
		required := 1
		if r.Session.CurrentStageEventID != nil || r.Session.State == TimelineFailed {
			required++
		}
		if len(r.Receipts)+1+required > TimelineMaxReceipts {
			return result, timelineFailure("capacity", 413)
		}
	}
	if !terminal {
		limit := TimelineContentReceipts
		if lifecycle {
			limit = TimelineMaxReceipts
		}
		if len(r.Receipts) >= limit {
			return result, timelineFailure("capacity", 413)
		}
	}
	if r.Session != nil {
		raw, _ := json.Marshal(r.Session)
		budget := TimelineContentBudget
		if lifecycle {
			budget = TimelineMaxBody
		}
		if len(raw) > budget {
			return result, timelineFailure("capacity", 413)
		}
	}
	if err := result.Validate(); err != nil {
		return result, err
	}
	r.Receipts = append(r.Receipts, TimelineOperationReceipt{OperationID: op, RequestSHA256: digest, CompletedAt: s.timelineTime(), Result: result})
	if !lifecycle {
		raw, err := json.Marshal(r)
		if err != nil {
			return result, err
		}
		if len(raw) > timelineMaxRecordBytes-(TimelineMaxBody-TimelineContentBudget) {
			return result, timelineFailure("capacity", 413)
		}
	}
	if err := validateTimelineSerializedClosureReserve(r); err != nil {
		return result, err
	}
	return result, s.timelineSave(r)
}
func timelineResult(r *timelineRecord, kind TimelineOperationKind, op string) TimelineOperationResult {
	v := r.Session
	return TimelineOperationResult{SchemaVersion: 3, OperationID: op, Kind: kind, ViewerSessionID: r.ID, SessionVersion: v.Version, CurrentStageEventID: v.CurrentStageEventID}
}
func (s *Store) timelineProject(source TimelineSource) (string, string, string, error) {
	root, err := canonicalDir(source.ProjectRoot)
	if err != nil {
		return "", "", "", timelineFailure("project_unavailable", 409)
	}
	registry, err := readRegistry(s.vdir)
	if err != nil {
		return "", "", "", timelineFailure("project_unavailable", 409)
	}
	registered := false
	for _, repo := range registry.Repositories {
		if repo.Root == root {
			registered = true
			break
		}
	}
	if !registered {
		return "", "", "", timelineFailure("project_not_registered", 409)
	}
	bound := ""
	for _, record := range s.timeline {
		for _, alias := range record.Roots {
			if alias == root {
				if bound != "" && bound != record.ProjectKey {
					return "", "", "", timelineFailure("project_identity_changed", 409)
				}
				bound = record.ProjectKey
			}
		}
	}
	key := root
	resolved := false
	if s.remote != nil {
		if _, gitErr := os.Lstat(filepath.Join(root, ".git")); gitErr == nil {
			raw, remoteErr := s.remote(root)
			resolved = remoteErr == nil
			if remoteErr != nil {
				if bound == "" {
					return "", "", "", timelineFailure("project_identity_unavailable", 409)
				}
				key = bound
			} else if raw != "" {
				key = NormalizeRemote(raw)
				if key == "" {
					return "", "", "", timelineFailure("project_identity_unavailable", 409)
				}
			}
		}
	}
	if !resolved && bound != "" {
		key = bound
	}
	if bound != "" && bound != key {
		return "", "", "", timelineFailure("project_identity_changed", 409)
	}
	name := s.resolveProject(root).name
	return root, key, name, nil
}

// TimelineCreate gets or creates exactly one registered native tuple.
func (s *Store) TimelineCreate(request TimelineCreateSessionRequest) (TimelineSessionResponse, error) {
	if err := request.Validate(); err != nil {
		return TimelineSessionResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.timelineReconcile(); err != nil {
		return TimelineSessionResponse{}, err
	}
	root, key, name, err := s.timelineProject(request.Source)
	if err != nil {
		return TimelineSessionResponse{}, err
	}
	effective := timelineEffectiveWorkKey(request)
	id := timelineSessionKey(key, request.Source, effective)
	digest := timelineDigest("create_session", id, "", request, nil)
	if old, ok := s.timeline[id]; ok {
		if old.Session == nil || old.DeletePending {
			return TimelineSessionResponse{}, timelineFailure("session_deleted", 410)
		}
		result, replay, err := s.timelineReplay(old, request.OperationID, digest)
		if err != nil {
			return TimelineSessionResponse{}, err
		}
		if replay {
			return timelineResponse(*timelineClone(old).Session, *result.Created), nil
		}
		r := timelineClone(old)
		found := false
		for _, alias := range r.Roots {
			found = found || alias == root
		}
		if !found {
			if len(r.Roots) >= 64 {
				return TimelineSessionResponse{}, timelineFailure("capacity", 413)
			}
			r.Roots = append(r.Roots, root)
		}
		actor := TimelineActor{Kind: "native_producer", ProducerID: r.Session.ForegroundProducerID}
		for _, producer := range r.Session.Producers {
			if timelineNativeSourceIdentity(producer.Source) == timelineNativeSourceIdentity(request.Source) {
				actor.ProducerID = producer.ProducerID
				break
			}
		}
		if err = s.timelineAudit(r, "create_session", request.OperationID, id, &actor, nil, "Resolved original work tuple"); err != nil {
			return TimelineSessionResponse{}, err
		}
		created := false
		result = timelineResult(r, "create_session", request.OperationID)
		result.Created = &created
		if _, err = s.timelineCommit(r, "create_session", request.OperationID, digest, result, false); err != nil {
			return TimelineSessionResponse{}, err
		}
		return timelineResponse(*r.Session, false), nil
	}
	now := s.timelineTime()
	producer, err := timelineRandomID("producer-")
	if err != nil {
		return TimelineSessionResponse{}, err
	}
	tickets := make([]TimelineTicket, 0, len(request.Tickets))
	for _, input := range request.Tickets {
		tid, e := timelineRandomID("ticket-")
		if e != nil {
			return TimelineSessionResponse{}, e
		}
		input.ExternalScope, _ = CanonicalTimelineIdentityURL(input.ExternalScope)
		tickets = append(tickets, TimelineTicket{TicketID: tid, TimelineTicketInput: input})
	}
	var workKey *string
	if request.WorkKey != nil {
		copied := *request.WorkKey
		workKey = &copied
	}
	v := TimelineSession{SchemaVersion: 3, ViewerSessionID: id, ProjectKey: key, ProjectName: name, Source: request.Source, Version: 1, WorkKey: workKey, Title: request.Title, Summary: request.Summary, State: TimelineActive, CreatedAt: now, UpdatedAt: now, Tickets: tickets, Deliveries: []TimelineDelivery{}, Candidates: []TimelineCandidateRecord{}, Events: []TimelineEvent{}, Artifacts: []TimelineArtifact{}, Documents: []TimelineDocument{}}
	v.Producers = []TimelineProducer{{ProducerID: producer, Source: request.Source, JoinedAt: now}}
	v.ForegroundProducerID = producer
	v.PendingJoinRequests = []TimelineJoinRequest{}
	v.Tasks = []TimelineTask{}
	v.Waits = []TimelineCIWait{}
	if err = ValidateTimelineResolvedProject(v, request.Source, key); err != nil {
		return TimelineSessionResponse{}, err
	}
	r := &timelineRecord{StorageVersion: 3, ID: id, Source: request.Source, ProjectKey: key, ProjectName: name, Roots: timelineSourceRoots(root, request.Source.ProjectRoot), WorkKey: effective, Session: &v, Receipts: []TimelineOperationReceipt{}}
	if err = s.timelineAudit(r, "create_session", request.OperationID, id, &TimelineActor{Kind: "native_producer", ProducerID: producer}, nil, "Created work from actual source"); err != nil {
		return TimelineSessionResponse{}, err
	}
	yes := true
	result := timelineResult(r, "create_session", request.OperationID)
	result.Created = &yes
	if _, err = s.timelineCommit(r, "create_session", request.OperationID, digest, result, false); err != nil {
		return TimelineSessionResponse{}, err
	}
	return timelineResponse(v, true), nil
}
func timelineResponse(v TimelineSession, created bool) TimelineSessionResponse {
	return TimelineSessionResponse{SchemaVersion: 3, Mode: "timeline_api", Session: v, Created: &created}
}
func (s *Store) TimelineGet(id string) (TimelineSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, err := s.timelineGetLocked(id)
	if err != nil {
		return TimelineSession{}, err
	}
	if r.Session == nil || r.DeletePending {
		return TimelineSession{}, timelineFailure("session_deleted", 410)
	}
	return *timelineClone(r).Session, nil
}
func timelineBump(v *TimelineSession, now string) error {
	if v.Version >= TimelineMaxNumber {
		return timelineFailure("capacity", 413)
	}
	v.Version++
	v.UpdatedAt = now
	return nil
}
func timelinePauseForeground(v *TimelineSession, except, now string) {
	if v.CurrentStageEventID == nil || *v.CurrentStageEventID == except {
		return
	}
	for i := range v.Events {
		e := &v.Events[i]
		if e.EventID == *v.CurrentStageEventID {
			e.StageState = "paused"
			e.Version++
			e.UpdatedAt = now
			e.FinishedAt = ""
			break
		}
	}
	v.CurrentStageEventID = nil
}
func (s *Store) TimelinePatch(id string, request TimelineSessionPatchRequest) (TimelineOperationResult, error) {
	if err := request.Validate(); err != nil {
		return TimelineOperationResult{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.timelineGetLocked(id)
	if err != nil {
		return TimelineOperationResult{}, err
	}
	digest := timelineDigest("session_patch", id, "", request, nil)
	if result, done, err := s.timelineReplay(old, request.OperationID, digest); done || err != nil {
		return result, err
	}
	if err = ValidateTimelineSessionPatch(*old.Session, request); err != nil {
		return TimelineOperationResult{}, err
	}
	if request.Actor.Kind == "native_producer" && (old.Session.State == TimelineClosed || old.Session.State == "failed") && (request.State == nil || *request.State != TimelineActive) {
		return TimelineOperationResult{}, timelineFailure("session_closed", 409)
	}
	r := timelineClone(old)
	v := r.Session
	now := s.timelineTime()
	if request.Title != nil {
		title, titleErr := validateSessionTitle(*request.Title)
		if titleErr != nil || title == "" {
			return TimelineOperationResult{}, timelineFailure("invalid_title", 400)
		}
		v.Title = title
	}
	if request.Summary != nil {
		v.Summary = *request.Summary
	}
	if request.Archived != nil {
		v.Archived = *request.Archived
	}
	if request.Pinned != nil {
		v.Pinned = *request.Pinned
	}
	if err = s.timelinePatchState(v, request); err != nil {
		return TimelineOperationResult{}, err
	}
	if request.Actor.Kind == "local_viewer" {
		if err = s.timelineLocalAudit(v, request, *old.Session); err != nil {
			return TimelineOperationResult{}, err
		}
	} else if request.State == nil {
		if err = s.timelineAudit(r, "session_patch", request.OperationID, id, &request.Actor, nil, request.Reason); err != nil {
			return TimelineOperationResult{}, err
		}
	}
	if err = timelineBump(v, now); err != nil {
		return TimelineOperationResult{}, err
	}
	result := timelineResult(r, "session_patch", request.OperationID)
	return s.timelineCommit(r, "session_patch", request.OperationID, digest, result, false)
}
func (s *Store) TimelineAppend(id string, request TimelineAppendEventRequest) (TimelineOperationResult, error) {
	if err := request.Validate(); err != nil {
		return TimelineOperationResult{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.timelineGetLocked(id)
	if err != nil {
		return TimelineOperationResult{}, err
	}
	digest := timelineDigest("append_event", id, "", request, nil)
	if result, done, err := s.timelineReplay(old, request.OperationID, digest); done || err != nil {
		return result, err
	}
	if err = ValidateTimelineEventReferences(*old.Session, request); err != nil {
		return TimelineOperationResult{}, err
	}
	r := timelineClone(old)
	v := r.Session
	associated, err := ResolveTimelineStageEventID(*v, request)
	if err != nil {
		return TimelineOperationResult{}, err
	}
	lifecycle := request.Kind == "stage" || request.Kind == "closure"
	if err = ValidateTimelineEventCapacity(len(v.Events), 1, lifecycle); err != nil {
		return TimelineOperationResult{}, err
	}
	if request.Kind == "stage" {
		if err = s.timelineSuspendOther(v, "", request.Actor, request.Reason); err != nil {
			return TimelineOperationResult{}, err
		}
	}
	eid, err := timelineRandomID("event-")
	if err != nil {
		return TimelineOperationResult{}, err
	}
	now := s.timelineTime()
	ticketRefs, e := CaptureTimelineTicketReferences(*v, request.TicketIDs)
	if e != nil {
		return TimelineOperationResult{}, e
	}
	var deliveryRefs []TimelineDeliveryReference
	if request.DeliveryIDs != nil {
		deliveryRefs, e = CaptureTimelineDeliveryReferences(*v, request.DeliveryIDs)
		if e != nil {
			return TimelineOperationResult{}, e
		}
	}
	event := TimelineEvent{Actor: request.Actor, StageEventID: associated, TaskID: request.TaskID, SegmentID: request.SegmentID, Reason: request.Reason, ReviewStatus: request.ReviewStatus, EventID: eid, Sequence: int64(len(v.Events) + 1), Version: 1, Kind: request.Kind, Title: request.Title, Summary: request.Summary, CreatedAt: now, UpdatedAt: now, Origin: "agent_reported", TicketRefs: ticketRefs, CandidateID: request.CandidateID, RelatedEventID: request.RelatedEventID, Target: request.Target, SourceReference: request.SourceReference, DeliveryRefs: deliveryRefs, Closure: request.Closure, Stage: request.Stage, StageState: request.StageState, OccurredAt: request.OccurredAt}
	if event.Kind == "stage" {
		v.State = TimelineActive
		v.Reason = request.Reason
		event.StartedAt = now
		v.CurrentStageEventID = &event.EventID
	}
	if len(request.Views) > 0 {
		event.EvidenceID, err = timelineRandomID("evidence-")
		if err != nil {
			return TimelineOperationResult{}, err
		}
		union := map[string]bool{}
		for _, ref := range event.TicketRefs {
			union[TimelineTicketIdentity(ref)] = true
		}
		for _, input := range request.Views {
			viewID := input.ViewID
			if viewID == "" {
				viewID, err = timelineRandomID("view-")
				if err != nil {
					return TimelineOperationResult{}, err
				}
			}
			refs, e := CaptureTimelineTicketReferences(*v, input.TicketIDs)
			if e != nil {
				return TimelineOperationResult{}, e
			}
			event.Views = append(event.Views, TimelineView{ViewID: viewID, Title: input.Title, TicketRefs: refs, Steps: input.Steps, ExpectedResult: input.ExpectedResult, ImageRefs: input.ImageRefs, Supersedes: input.Supersedes})
			for _, ref := range refs {
				if !union[TimelineTicketIdentity(ref)] {
					event.TicketRefs = append(event.TicketRefs, ref)
					union[TimelineTicketIdentity(ref)] = true
				}
			}
		}
		v.ValidationCount++
	}
	if event.Kind == "closure" {
		v.State = TimelineClosed
		v.Reason = request.Reason
	}
	v.Events = append(v.Events, event)
	if request.Kind == "closure" {
		if err = s.timelineAuditEvent(v, "state_transition", request.Actor, "", "", "", TimelineAudit{Operation: "closure", Reason: request.Reason, PreviousState: string(old.Session.State), NextState: string(v.State)}, true); err != nil {
			return TimelineOperationResult{}, err
		}
	}
	if err = timelineBumpTask(v, event.TaskID); err != nil {
		return TimelineOperationResult{}, err
	}
	if err = timelineBump(v, now); err != nil {
		return TimelineOperationResult{}, err
	}
	result := timelineResult(r, "append_event", request.OperationID)
	result.EventID = eid
	result.EventVersion = 1
	result.Sequence = event.Sequence
	result.EvidenceID = event.EvidenceID
	if len(event.Views) > 0 {
		result.ViewIDs = []string{}
		for _, view := range event.Views {
			result.ViewIDs = append(result.ViewIDs, view.ViewID)
		}
	}
	return s.timelineCommit(r, "append_event", request.OperationID, digest, result, false)
}
func (s *Store) TimelinePatchEvent(id, eid string, request TimelineEventPatchRequest) (TimelineOperationResult, error) {
	if err := request.Validate(); err != nil {
		return TimelineOperationResult{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.timelineGetLocked(id)
	if err != nil {
		return TimelineOperationResult{}, err
	}
	digest := timelineDigest("event_patch", id, eid, request, nil)
	if result, done, err := s.timelineReplay(old, request.OperationID, digest); done || err != nil {
		return result, err
	}
	if err = ValidateTimelineEventPatch(*old.Session, eid, request); err != nil {
		return TimelineOperationResult{}, err
	}
	if request.StageState != nil && old.Session.State == TimelineFailed {
		return TimelineOperationResult{}, timelineFailure("session_terminal", 409)
	}
	r := timelineClone(old)
	v := r.Session
	for i := range v.Events {
		e := &v.Events[i]
		if e.EventID != eid {
			continue
		}
		if e.Version >= TimelineMaxNumber {
			return TimelineOperationResult{}, timelineFailure("capacity", 413)
		}
		if request.Title != nil {
			e.Title = *request.Title
		}
		if request.Summary != nil {
			e.Summary = *request.Summary
		}
		before := old.Session.Events[i]
		if request.StageState != nil && *request.StageState != before.StageState {
			// Auditing may reallocate Events; preserve the copy edits before transitions.
			if contains([]string{"in_progress", "in_review", "awaiting_validation"}, string(*request.StageState)) {
				if err = s.timelineSuspendOther(v, eid, request.Actor, request.Reason); err != nil {
					return TimelineOperationResult{}, err
				}
			}
			if err = s.timelineSetStage(v, eid, *request.StageState, request.Actor, request.Reason); err != nil {
				return TimelineOperationResult{}, err
			}
			for j := range v.Events {
				if v.Events[j].EventID == eid {
					e = &v.Events[j]
					break
				}
			}
		}
		if !timelineSameEventCopy(before, *e) {
			if err = timelineBumpTask(v, timelineRelatedTasks(*v, eid)...); err != nil {
				return TimelineOperationResult{}, err
			}
			if request.StageState == nil {
				if err = s.timelineAudit(r, "event_patch", request.OperationID, eid, &request.Actor, nil, request.Reason); err != nil {
					return TimelineOperationResult{}, err
				}
			}
		}
		if !timelineSameEventCopy(before, *e) {
			if request.StageState == nil || before.StageState == *request.StageState {
				e.Version++
			}
			e.UpdatedAt = s.timelineTime()
		}
		if err = timelineBump(v, s.timelineTime()); err != nil {
			return TimelineOperationResult{}, err
		}
		result := timelineResult(r, "event_patch", request.OperationID)
		result.EventID = eid
		result.EventVersion = e.Version
		result.Sequence = e.Sequence
		return s.timelineCommit(r, "event_patch", request.OperationID, digest, result, false)
	}
	return TimelineOperationResult{}, ErrNotFound
}
func (s *Store) TimelineWithdraw(id, eid string, request TimelineMutationRequest) (TimelineOperationResult, error) {
	if err := request.Validate(); err != nil {
		return TimelineOperationResult{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.timelineGetLocked(id)
	if err != nil {
		return TimelineOperationResult{}, err
	}
	digest := timelineDigest("event_withdraw", id, eid, request, nil)
	if result, done, err := s.timelineReplay(old, request.OperationID, digest); done || err != nil {
		return result, err
	}
	if err = ValidateTimelineEventWithdrawal(*old.Session, eid, request); err != nil {
		return TimelineOperationResult{}, err
	}
	if old.Session.CurrentStageEventID != nil && *old.Session.CurrentStageEventID == eid {
		return TimelineOperationResult{}, timelineFailure("stage_active", 409)
	}
	r := timelineClone(old)
	v := r.Session
	now := s.timelineTime()
	for i := range v.Events {
		e := &v.Events[i]
		if e.EventID != eid {
			continue
		}
		if !e.Withdrawn && (e.Kind == "validation" || e.Kind == "iteration") {
			v.ValidationCount--
		}
		e.Withdrawn = true
		if !old.Session.Events[i].Withdrawn {
			if err = timelineBumpTask(v, timelineRelatedTasks(*v, eid)...); err != nil {
				return TimelineOperationResult{}, err
			}
			if err = s.timelineAudit(r, "event_withdraw", request.OperationID, eid, &request.Actor, nil, "Explicit content withdrawal"); err != nil {
				return TimelineOperationResult{}, err
			}
		}
		e.Version++
		e.UpdatedAt = now
		if err = timelineBump(v, now); err != nil {
			return TimelineOperationResult{}, err
		}
		result := timelineResult(r, "event_withdraw", request.OperationID)
		result.EventID = eid
		result.EventVersion = e.Version
		result.Sequence = e.Sequence
		yes := true
		result.Withdrawn = &yes
		return s.timelineCommit(r, "event_withdraw", request.OperationID, digest, result, false)
	}
	return TimelineOperationResult{}, ErrNotFound
}
func (s *Store) TimelineSnapshot(id, eid string, request TimelineDocumentSnapshotRequest) (TimelineOperationResult, error) {
	if err := request.Validate(); err != nil {
		return TimelineOperationResult{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.timelineGetLocked(id)
	if err != nil {
		return TimelineOperationResult{}, err
	}
	digest := timelineDigest("document_snapshot", id, eid, request, nil)
	if result, done, err := s.timelineReplay(old, request.OperationID, digest); done || err != nil {
		return result, err
	}
	if err = ValidateTimelineDocumentSnapshot(*old.Session, eid, request); err != nil {
		return TimelineOperationResult{}, err
	}
	for _, event := range old.Session.Events {
		if event.EventID != eid {
			continue
		}
		for _, previous := range event.DocumentSnapshots {
			if previous.DocumentRevisionID == request.PredecessorRevisionID && previous.DocumentKey == request.DocumentKey && previous.Status == request.Status && timelineEqual(previous.Tabs, request.Tabs) && previous.TaskID == request.TaskID && previous.SegmentID == request.SegmentID {
				r := timelineClone(old)
				result := timelineResult(r, "document_snapshot", request.OperationID)
				result.EventID = eid
				result.EventVersion = event.Version
				result.Sequence = event.Sequence
				result.DocumentRevisionID = previous.DocumentRevisionID
				return s.timelineCommit(r, "document_snapshot", request.OperationID, digest, result, false)
			}
		}
	}
	r := timelineClone(old)
	v := r.Session
	rid, err := timelineRandomID("revision-")
	if err != nil {
		return TimelineOperationResult{}, err
	}
	now := s.timelineTime()
	for i := range v.Events {
		e := &v.Events[i]
		if e.EventID != eid {
			continue
		}
		if e.Version >= TimelineMaxNumber {
			return TimelineOperationResult{}, timelineFailure("capacity", 413)
		}
		e.DocumentSnapshots = append(e.DocumentSnapshots, TimelineDocumentSnapshot{Actor: request.Actor, DocumentKey: request.DocumentKey, PredecessorRevisionID: request.PredecessorRevisionID, TaskID: request.TaskID, SegmentID: request.SegmentID, DocumentRevisionID: rid, Status: request.Status, CreatedAt: now, Tabs: request.Tabs})
		if err = timelineBumpTask(v, append(timelineDocumentChainTasks(*v, request.DocumentKey), request.TaskID)...); err != nil {
			return TimelineOperationResult{}, err
		}
		e.Version++
		e.UpdatedAt = now
		if err = timelineBump(v, now); err != nil {
			return TimelineOperationResult{}, err
		}
		result := timelineResult(r, "document_snapshot", request.OperationID)
		result.DocumentRevisionID = rid
		result.EventID = eid
		result.EventVersion = e.Version
		result.Sequence = e.Sequence
		return s.timelineCommit(r, "document_snapshot", request.OperationID, digest, result, false)
	}
	return TimelineOperationResult{}, ErrNotFound
}

func (s *Store) timelineHasID(id string) bool {
	s.viewMu.RLock()
	defer s.viewMu.RUnlock()
	return s.timelineOwned[id]
}
func (s *Store) timelineLegacyHidden(v Session) bool {
	return s.timelineHasID(timelineSessionKey(v.ProjectKey, TimelineSource{Host: TimelineHost(v.Host), NativeSessionID: v.SessionID}, ""))
}

func timelineSourceRoots(root, source string) []string {
	if root == source {
		return []string{root}
	}
	return []string{root, source}
}
