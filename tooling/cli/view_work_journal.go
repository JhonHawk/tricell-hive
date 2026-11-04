package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

const workJournalLimit = 16 << 20

type viewWorkBinding struct {
	WorkID   string                `json:"work_id"`
	Key      string                `json:"key"`
	Project  string                `json:"project"`
	Change   string                `json:"change"`
	Source   viewer.TimelineSource `json:"source"`
	Previous []string              `json:"previous,omitempty"`
	Reason   string                `json:"reason,omitempty"`
}
type viewWorkStep struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	ContentType string `json:"content_type"`
	Body        []byte `json:"body"`
	Response    []byte `json:"response,omitempty"`
	Rejection   string `json:"rejection,omitempty"`
	Attempted   bool   `json:"attempted"`
}
type viewWorkIntent struct {
	ID            string                `json:"id"`
	WorkID        string                `json:"work_id"`
	Source        viewer.TimelineSource `json:"source"`
	Action        string                `json:"action"`
	State         string                `json:"state"`
	Context       json.RawMessage       `json:"context,omitempty"`
	Steps         []viewWorkStep        `json:"steps"`
	ArchiveFile   string                `json:"archive_file,omitempty"`
	ArchiveReason string                `json:"archive_reason,omitempty"`
}
type viewWorkJournal struct {
	SchemaVersion int               `json:"schema_version"`
	Bindings      []viewWorkBinding `json:"bindings"`
	Pending       []*viewWorkIntent `json:"pending"`
	History       []*viewWorkIntent `json:"history"`
}

func newWorkJournal() *viewWorkJournal {
	return &viewWorkJournal{SchemaVersion: 1, Bindings: []viewWorkBinding{}, Pending: []*viewWorkIntent{}, History: []*viewWorkIntent{}}
}
func workID(prefix string) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("system randomness unavailable")
	}
	return prefix + hex.EncodeToString(b)
}
func workStateError(code string) error {
	return &viewWorkError{Code: code, Next: "preserve publisher state and resolve the reported condition before retrying"}
}
func safeWorkFile(path string) error {
	st, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() || st.Mode().Perm()&0077 != 0 || workFileHasLinks(st) {
		return workStateError("unsafe_publisher_file")
	}
	return nil
}
func lockWorkPublisher(dir string) (*viewer.FileLock, error) {
	path := filepath.Join(dir, "publisher.lock")
	if err := safeWorkFile(path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	l, err := viewer.AcquireLock(path, false)
	if errors.Is(err, viewer.ErrLocked) {
		return nil, workStateError("busy")
	}
	return l, err
}
func readWorkJournal(dir string) (*viewWorkJournal, error) {
	path := filepath.Join(dir, "journal.json")
	if err := safeWorkFile(path); os.IsNotExist(err) {
		return newWorkJournal(), nil
	} else if err != nil {
		return nil, err
	}
	raw, err := readPrivateWorkFile(path)
	if err != nil {
		return nil, err
	}
	j := &viewWorkJournal{}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if dec.Decode(j) != nil || dec.Decode(new(any)) != io.EOF || j.SchemaVersion != 1 || j.Bindings == nil || j.Pending == nil || j.History == nil {
		return nil, workStateError("invalid_journal")
	}
	for n, i := range j.History {
		if i != nil && i.ArchiveFile != "" {
			if i.ArchiveFile != "archive-"+i.ID+".json" || !viewTimelineObjectID.MatchString(i.ID) {
				return nil, workStateError("invalid_journal")
			}
			archive, err := readPrivateWorkFile(filepath.Join(dir, i.ArchiveFile))
			if err != nil {
				return nil, err
			}
			var full viewWorkIntent
			dec := json.NewDecoder(bytes.NewReader(archive))
			dec.DisallowUnknownFields()
			if dec.Decode(&full) != nil || dec.Decode(new(any)) != io.EOF || full.ArchiveFile != "" || full.ID != i.ID || full.WorkID != i.WorkID || full.Source != i.Source || full.State != "archived_rejected" {
				return nil, workStateError("invalid_journal")
			}
			full.ArchiveFile = i.ArchiveFile
			j.History[n] = &full
		}
	}
	if err := validateWorkJournal(j); err != nil {
		return nil, err
	}
	return j, nil
}
func writeWorkJournal(dir string, j *viewWorkJournal) error {
	// Rejected evidence is immutable and retained separately. The index rename
	// is the sole active/archive classification commit; an orphan file from an
	// interrupted archive is never treated as archived without that index entry.
	disk := *j
	disk.History = append([]*viewWorkIntent{}, j.History...)
	for n, i := range disk.History {
		if i.State != "archived_rejected" {
			continue
		}
		full := *i
		full.ArchiveFile = ""
		raw, err := json.Marshal(full)
		if err != nil || len(raw) > workJournalLimit {
			return workStateError("journal_limit")
		}
		name := "archive-" + i.ID + ".json"
		path := filepath.Join(dir, name)
		if existing, err := readPrivateWorkFile(path); err == nil {
			if !bytes.Equal(existing, raw) {
				return workStateError("archive_conflict")
			}
		} else if os.IsNotExist(err) {
			if err = writePrivateWorkFile(dir, path, raw); err != nil {
				return err
			}
		} else {
			return err
		}
		stub := *i
		stub.Context = nil
		stub.Steps = []viewWorkStep{}
		stub.ArchiveFile = name
		disk.History[n] = &stub
	}
	raw, err := json.Marshal(&disk)
	if err != nil || len(raw) > workJournalLimit {
		return workStateError("journal_limit")
	}
	return writePrivateWorkFile(dir, filepath.Join(dir, "journal.json"), raw)
}
func writePrivateWorkFile(dir, path string, raw []byte) error {
	if err := safeWorkFile(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	f, err := os.CreateTemp(dir, ".journal-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(raw)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Rename(name, path); err != nil {
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
func definitiveWorkRejection(code string) bool {
	switch code {
	case "version_conflict", "session_version_conflict", "event_version_conflict", "current_predecessor_required", "predecessor_conflict", "work_closed", "session_closed", "session_deleted", "not_found", "forbidden", "producer_required", "lead_required":
		return true
	// Explicit v3 validation/authority/precondition responses reject the write
	// before a receipt is committed. Keep this finite: transport failures,
	// malformed responses, operation_conflict and unknown codes remain uncertain.
	case "scope_reference_required", "entity_exists", "entity_conflict", "duplicate_plan_order",
		"stage_active", "stage_terminal", "stage_resume_required", "already_in_state", "session_terminal",
		"stage_unavailable", "stage_binding_mismatch", "reentry_binding_required", "duplicate_identity", "duplicate_key", "duplicate_stage_order",
		"producer_not_accepted", "content_owner_required", "segment_assignee_required", "forbidden_actor", "forbidden_for_operation", "local_viewer_required", "native_producer_required",
		"unknown_event", "unknown_task", "unknown_wait", "unknown_ticket", "unknown_delivery", "unknown_candidate",
		"pending_request_required", "immutable_audit", "invalid_transition", "invalid_state_transition", "invalid_producer_transition", "foreign_project", "unknown_or_mixed_attempt", "candidate_mismatch", "attempt_already_known",
		"capacity", "session_limit", "limit", "unsupported_protocol", "invalid_request", "invalid_id":
		return true
	}
	return false
}
func (j *viewWorkJournal) archiveRejected(id, reason string) error {
	for n, i := range j.Pending {
		if i.ID != id {
			continue
		}
		if reason == "" || i.State != "rejected_conflict" {
			return workStateError("outcome_unknown")
		}
		rejected := false
		for _, s := range i.Steps {
			if len(s.Response) > 0 {
				continue
			}
			if definitiveWorkRejection(s.Rejection) {
				rejected = true
				continue
			}
			if s.Attempted {
				return workStateError("outcome_unknown")
			}
		}
		if !rejected {
			return workStateError("outcome_unknown")
		}
		i.State = "archived_rejected"
		i.ArchiveReason = reason
		j.History = append(j.History, i)
		j.Pending = append(j.Pending[:n], j.Pending[n+1:]...)
		return nil
	}
	return workStateError("pending_not_found")
}

func validWorkIntentState(state string) bool {
	switch state {
	case "prepared", "resolving", "uploading", "snapshot_pending", "verifying", "confirmed", "needs_reconciliation", "rejected_conflict", "archived_rejected":
		return true
	}
	return false
}

func workFileHasLinks(st os.FileInfo) bool {
	stat, ok := st.Sys().(*syscall.Stat_t)
	return ok && stat.Nlink != 1
}
func readPrivateWorkFile(path string) ([]byte, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if err = safeWorkFile(path); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(before, opened) || !opened.Mode().IsRegular() || opened.Mode().Perm()&0077 != 0 || workFileHasLinks(opened) {
		return nil, workStateError("unsafe_publisher_file")
	}
	raw, err := io.ReadAll(io.LimitReader(f, workJournalLimit+1))
	after, statErr := f.Stat()
	if err != nil || statErr != nil || len(raw) > workJournalLimit || after.Size() != opened.Size() || after.ModTime() != opened.ModTime() {
		return nil, workStateError("invalid_journal")
	}
	return raw, nil
}
func compactConfirmedWorkIntent(i *viewWorkIntent) *viewWorkIntent {
	compact := *i
	compact.State = "confirmed"
	compact.Steps = append([]viewWorkStep{}, i.Steps...)
	for n := range compact.Steps {
		compact.Steps[n].Body = nil
	}
	switch i.Action {
	case "publish":
		var p workPublicationIntent
		if json.Unmarshal(i.Context, &p) == nil {
			for n := range p.Documents {
				p.Documents[n].Bytes = nil
			}
			p.Command.Options = management.Options{}
			compact.Context, _ = json.Marshal(p)
		} else {
			compact.Context = nil
		}
	case "open":
		compact.Context = nil
	default:
		var c viewWorkLifecycleContext
		if json.Unmarshal(i.Context, &c) == nil {
			c.FirstStep = nil
			c.Command.Options = management.Options{}
			compact.Context, _ = json.Marshal(c)
		} else {
			compact.Context = nil
		}
	}
	return &compact
}

// Confirmed evidence remains authoritative on the server. Local display
// history is bounded by count and bytes; rejected recovery evidence is exempt.
const workConfirmedHistoryLimit = 256
const workConfirmedHistoryBytes = 1 << 20

func (j *viewWorkJournal) compactConfirmedHistory() {
	count, size := 0, 0
	keep := make([]bool, len(j.History))
	for n := len(j.History) - 1; n >= 0; n-- {
		i := j.History[n]
		if i.State != "confirmed" {
			keep[n] = true
			continue
		}
		raw, _ := json.Marshal(i)
		if count < workConfirmedHistoryLimit && size+len(raw) <= workConfirmedHistoryBytes {
			keep[n] = true
			count++
			size += len(raw)
		}
	}
	history := j.History[:0]
	for n, i := range j.History {
		if keep[n] {
			history = append(history, i)
		}
	}
	j.History = history
}
func validWorkStoredPath(path string) bool {
	return filepath.IsAbs(path) && filepath.Clean(path) == path && !strings.ContainsRune(path, 0)
}
func validateWorkSource(source viewer.TimelineSource) bool {
	return validWorkStoredPath(source.ProjectRoot) && (viewer.TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "validate", Source: source, Title: "Validate", Summary: "Validate", Tickets: []viewer.TimelineTicketInput{}}).Validate() == nil
}
func validWorkAction(action string) bool {
	switch action {
	case "open", "publish", "stage", "tasks register", "task update", "close":
		return true
	}
	return false
}
func validateWorkJournal(j *viewWorkJournal) error {
	bad := func() error { return workStateError("invalid_journal") }
	bindings := map[string]bool{}
	dirs := map[string]bool{}
	for _, b := range j.Bindings {
		identity := string(b.Source.Host) + "\x00" + b.Source.NativeSessionID
		if !viewTimelineSessionID.MatchString(b.WorkID) || !validateWorkSource(b.Source) || !validWorkStoredPath(b.Project) || !validWorkStoredPath(b.Change) || b.Source.ProjectRoot != b.Project {
			return bad()
		}
		rel, err := filepath.Rel(b.Project, b.Change)
		if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
			return bad()
		}
		if b.Key != "" {
			key := b.Key
			if (viewer.TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "validate", Source: b.Source, Title: "Validate", Summary: "Validate", Tickets: []viewer.TimelineTicketInput{}, WorkKey: &key}).Validate() != nil {
				return bad()
			}
		}
		if bindings[identity+b.WorkID] || dirs[identity+b.Change] {
			return bad()
		}
		bindings[identity+b.WorkID] = true
		dirs[identity+b.Change] = true
		for _, p := range b.Previous {
			if !validWorkStoredPath(p) || b.Reason == "" {
				return bad()
			}
		}
	}
	ids := map[string]bool{}
	for setIndex, set := range [][]*viewWorkIntent{j.Pending, j.History} {
		for _, i := range set {
			if i == nil || !viewTimelineObjectID.MatchString(i.ID) || ids[i.ID] || !validWorkIntentState(i.State) || !validWorkAction(i.Action) || !validateWorkSource(i.Source) || (i.WorkID != "" && !viewTimelineSessionID.MatchString(i.WorkID)) {
				return bad()
			}
			ids[i.ID] = true
			terminal := i.State == "confirmed" || i.State == "archived_rejected"
			if (setIndex == 0 && terminal) || (setIndex == 1 && !terminal) || ((i.Action != "open" && i.Action != "publish") && i.WorkID == "") || i.Steps == nil {
				return bad()
			}
			if i.State == "archived_rejected" && i.ArchiveReason == "" {
				return bad()
			}
			if !terminal && (len(i.Context) == 0 || !json.Valid(i.Context)) {
				return bad()
			}
			if !terminal {
				if err := validateWorkIntentContext(i); err != nil {
					return err
				}
			}
			for _, step := range i.Steps {
				if err := validateWorkStoredStep(i, step); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// workStoredCreateResponse adapts only a retained pre-revision-2 schema-3
// receipt. It returns a validation/use copy; the journal's original bytes remain
// authoritative. Fresh network responses never pass through this path.
func workStoredCreateResponse(raw []byte) ([]byte, error) {
	if _, err := viewer.DecodeTimelineSessionResponse(raw); err == nil {
		return raw, nil
	}
	bad := func() ([]byte, error) { return nil, workStateError("invalid_journal") }
	var envelope map[string]json.RawMessage
	// The original public bound and scanner preserve the old receipt's size,
	// duplicate-key, null, encoding and exact schema-version guarantees.
	if viewer.DecodeTimelineJSON(raw, &envelope) != nil || string(bytes.TrimSpace(envelope["schema_version"])) != "3" {
		return bad()
	}
	var session map[string]json.RawMessage
	if json.Unmarshal(envelope["session"], &session) != nil || string(bytes.TrimSpace(session["schema_version"])) != "3" {
		return bad()
	}
	defaults := map[string]json.RawMessage{
		"last_activity_at": session["created_at"], "closed_at": json.RawMessage("null"), "autoarchive_at": json.RawMessage("null"), "archive_origin": json.RawMessage("null"),
	}
	if bytes.Equal(bytes.TrimSpace(session["archived"]), []byte("true")) {
		defaults["archive_origin"] = json.RawMessage(`"manual"`)
	}
	missing := false
	for field, value := range defaults {
		if _, present := session[field]; !present {
			session[field] = value
			missing = true
		}
	}
	if !missing {
		return bad()
	}
	var err error
	envelope["session"], err = json.Marshal(session)
	if err != nil {
		return bad()
	}
	normalized, err := json.Marshal(envelope)
	if err != nil {
		return bad()
	}
	if _, err = viewer.DecodeTimelineSessionResponse(normalized); err != nil {
		return bad()
	}
	return normalized, nil
}

func validateWorkStoredStep(i *viewWorkIntent, s viewWorkStep) error {
	bad := func() error { return workStateError("invalid_journal") }
	operation := ""
	path := s.Path
	if s.Method == "POST" && path == "/api/timeline/sessions" && (i.Action == "open" || i.Action == "publish") {
		operation = "create_session"
	} else {
		prefix := "/api/timeline/sessions/" + i.WorkID
		if i.WorkID == "" || !strings.HasPrefix(path, prefix) {
			return bad()
		}
		tail := strings.TrimPrefix(path, prefix)
		parts := strings.Split(strings.TrimPrefix(tail, "/"), "/")
		switch {
		case tail == "" && s.Method == "PATCH":
			operation = "session_patch"
		case tail == "/events" && s.Method == "POST":
			operation = "stage_start"
			if i.Action == "close" {
				operation = "append_event"
			}
		case tail == "/documents" && s.Method == "POST":
			operation = "document_add"
		case tail == "/tasks" && s.Method == "POST":
			operation = "task_create"
		case len(parts) == 2 && viewTimelineObjectID.MatchString(parts[1]) && s.Method == "PATCH" && parts[0] == "events":
			operation = "stage_update"
		case len(parts) == 2 && viewTimelineObjectID.MatchString(parts[1]) && s.Method == "PATCH" && parts[0] == "tasks":
			operation = "task_patch"
		case len(parts) == 3 && viewTimelineObjectID.MatchString(parts[1]) && parts[0] == "events" && parts[2] == "document-snapshots" && s.Method == "POST":
			operation = "document_snapshot"
		case len(parts) == 3 && viewTimelineObjectID.MatchString(parts[1]) && parts[0] == "tasks" && parts[2] == "transitions" && s.Method == "POST":
			operation = "task_transition"
		default:
			return bad()
		}
	}
	if s.Rejection != "" && (!s.Attempted || !definitiveWorkRejection(s.Rejection) || len(s.Response) > 0) {
		return bad()
	}
	if len(s.Response) > 0 && !s.Attempted {
		return bad()
	}
	if len(s.Response) > 0 {
		if operation == "create_session" {
			normalized, err := workStoredCreateResponse(s.Response)
			if err != nil {
				return bad()
			}
			r, err := viewer.DecodeTimelineSessionResponse(normalized)
			if err != nil || r.Session.ViewerSessionID != i.WorkID || r.Session.Source != i.Source {
				return bad()
			}
		} else {
			r, err := viewer.DecodeTimelineOperationResult(s.Response)
			if err != nil || r.ViewerSessionID != i.WorkID || (len(s.Body) > 0 && r.OperationID != workStepOperationID(s)) {
				return bad()
			}
		}
	}
	if i.State == "confirmed" && len(s.Body) == 0 {
		return nil
	}
	if len(s.Body) == 0 {
		return bad()
	}
	input := s.Body
	if media, params, err := mime.ParseMediaType(s.ContentType); err == nil && media == "multipart/form-data" {
		if operation != "document_add" {
			return bad()
		}
		reader := multipart.NewReader(bytes.NewReader(input), params["boundary"])
		found := false
		for {
			part, err := reader.NextPart()
			if err != nil {
				break
			}
			if part.FormName() == "metadata" {
				input, _ = io.ReadAll(io.LimitReader(part, viewer.TimelineMaxBody))
				found = true
				break
			}
			part.Close()
		}
		if !found {
			return bad()
		}
	} else if s.ContentType != "application/json" {
		return bad()
	}
	if err := validateViewTimelineInput(operation, input); err != nil {
		return &viewWorkError{Code: "invalid_journal", Field: "step." + operation + "." + viewer.TimelineContractCode(err)}
	}
	if operation == "append_event" {
		r, err := viewer.DecodeTimelineAppendEvent(input)
		if err != nil || r.Kind != "closure" {
			return bad()
		}
	}
	if operation == "create_session" {
		r, err := viewer.DecodeTimelineCreateSession(input)
		if err != nil || r.Source != i.Source {
			return bad()
		}
	}
	return nil
}

// Each admitted intent reserves 12 MiB: the 2 MiB frozen document set and
// multipart requests (base64 encoded), a maximum 2 MiB create response, and
// bounded typed receipts. This leaves room to durably record unknown outcomes
// and acknowledgments even when earlier rejected publications are retained.
const workIntentReservation = 12 << 20

func (j *viewWorkJournal) admitWorkIntent() error {
	base := *j
	base.Pending = []*viewWorkIntent{}
	base.History = append([]*viewWorkIntent{}, j.History...)
	for n, i := range base.History {
		if i.State == "archived_rejected" {
			stub := *i
			stub.Context = nil
			stub.Steps = []viewWorkStep{}
			stub.ArchiveFile = "archive-" + i.ID + ".json"
			base.History[n] = &stub
		}
	}
	raw, err := json.Marshal(&base)
	if err != nil || len(raw)+(len(j.Pending)+1)*workIntentReservation > workJournalLimit {
		return &viewWorkError{Code: "publisher_capacity", Next: "resume or reconcile an existing pending intent before starting another; retained rejected evidence remains preserved"}
	}
	return nil
}

func validateWorkIntentContext(i *viewWorkIntent) error {
	invalid := func() error { return &viewWorkError{Code: "invalid_journal", Field: "intent.context"} }
	var command viewWorkCommand
	switch i.Action {
	case "open":
		var c viewWorkOpenContext
		if json.Unmarshal(i.Context, &c) != nil || c.Binding.Source != i.Source || c.Binding.Project != i.Source.ProjectRoot || !validWorkStoredPath(c.Binding.Change) {
			return invalid()
		}
		rel, err := filepath.Rel(c.Binding.Project, c.Binding.Change)
		if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
			return invalid()
		}
		r, err := viewer.DecodeTimelineCreateSession(c.Request)
		if err != nil || r.Source != i.Source || r.WorkKey == nil || *r.WorkKey != c.Binding.Key {
			return invalid()
		}
		return nil
	case "publish":
		var p workPublicationIntent
		if json.Unmarshal(i.Context, &p) != nil || len(p.Documents) == 0 || len(p.Documents) > workDocumentLimit || !validWorkStoredPath(p.Command.Change) || p.Command.Project != i.Source.ProjectRoot {
			return invalid()
		}
		switch p.Phase {
		case "create", "stage", "upload", "snapshot", "verify":
		default:
			return invalid()
		}
		command = p.Command
	default:
		var c viewWorkLifecycleContext
		if json.Unmarshal(i.Context, &c) != nil {
			return invalid()
		}
		command = c.Command
		if c.FirstStep != nil {
			if err := validateWorkStoredStep(i, *c.FirstStep); err != nil {
				return err
			}
		}
	}
	if command.Action != i.Action || command.Host != string(i.Source.Host) || command.NativeSessionID != i.Source.NativeSessionID || (command.WorkID != i.WorkID && !(i.Action == "publish" && command.WorkID == "")) {
		return invalid()
	}
	return nil
}
