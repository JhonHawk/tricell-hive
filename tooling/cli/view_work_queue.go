package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

// A semantic declaration (`work open`, `work sync`, `work evidence`) made while
// the viewer is stopped or silent is queued in the publisher journal and sent,
// in order, by the next semantic command that finds the viewer running. The
// queue holds the semantic input and an operation identity; the requests are
// built, and their exact bytes recorded, only when an entry is sent.
//
// A queued open cannot know the viewer's work ID: the viewer derives it from
// the project's server-side identity as well as from the work key. The work
// is addressed meanwhile by a local handle derived from the key and the
// session, shaped like a viewer ID. The journal maps the handle to the real ID
// once the open is sent; the handle stays the lock identity of the work.

const viewWorkQueueLimit = 256

var viewWorkFrozenName = regexp.MustCompile(`^queue-[A-Za-z0-9_-]{1,80}-[0-9]{1,3}$`)

// viewWorkCheckpoint is called at the points where a killed process leaves a
// journal state worth resuming; tests replace it to stop the process there.
var viewWorkCheckpoint = func(point string) {}

type viewWorkQueued struct {
	ID      string                `json:"id"`
	Action  string                `json:"action"`
	WorkID  string                `json:"work_id"`
	Source  viewer.TimelineSource `json:"source"`
	State   string                `json:"state"` // queued, started, rejected
	Command viewWorkCommand       `json:"command"`
	// Frozen names the evidence captures copied into the publisher folder.
	Frozen []string `json:"frozen,omitempty"`
	Reason string   `json:"reason,omitempty"`
}

type viewWorkAlias struct {
	Alias  string `json:"alias"`
	WorkID string `json:"work_id"`
}

// viewWorkQueueReport tells what the queue did before the command's own
// declaration: how many entries were sent and which no longer applied.
type viewWorkQueueReport struct {
	Sent            int                      `json:"sent,omitempty"`
	Rejected        []viewWorkQueueRejection `json:"rejected,omitempty"`
	RejectedOmitted int                      `json:"rejected_omitted,omitempty"`
}
type viewWorkQueueRejection struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

type viewWorkQueuedResult struct {
	WorkID      string `json:"work_id"`
	Result      string `json:"result"`
	OperationID string `json:"operation_id"`
	Position    int    `json:"position"`
}

func queuedWorkHandle(project, host, native, key string) string {
	raw, _ := json.Marshal([]string{"queued-work", project, host, native, key})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func viewWorkQueueable(c viewWorkCommand) bool {
	switch c.Action {
	case "sync", "evidence":
		return true
	case "open":
		return c.WorkID == ""
	}
	return false
}

// viewWorkQueueFlushes lists the actions that send the queue of their work
// before their own declaration.
func viewWorkQueueFlushes(action string) bool {
	switch action {
	case "open", "sync", "evidence", "stage", "tasks register", "task update", "close":
		return true
	}
	return false
}

func (e *viewWorkEngine) realID(id string) string {
	for _, a := range e.journal.Aliases {
		if a.Alias == id {
			return a.WorkID
		}
	}
	return id
}

// scopeOf is the lock identity of a work: its handle when it has one.
func (e *viewWorkEngine) scopeOf(id string) string {
	return workScopeOf(e.journal, id)
}

func workScopeOf(j *viewWorkJournal, id string) string {
	for _, a := range j.Aliases {
		if a.Alias == id || a.WorkID == id {
			return a.Alias
		}
	}
	return id
}

// queueSource finds the project of a queued declaration for a work that is not
// open in the viewer yet.
func (e *viewWorkEngine) queueSource(c viewWorkCommand) (viewer.TimelineSource, error) {
	source := viewer.TimelineSource{Host: viewer.TimelineHost(c.Host), NativeSessionID: c.NativeSessionID}
	scope := e.scopeOf(c.WorkID)
	for _, q := range e.journal.Queue {
		if e.scopeOf(q.WorkID) == scope && q.Source.ProjectRoot != "" {
			source.ProjectRoot = q.Source.ProjectRoot
		}
	}
	if source.ProjectRoot == "" {
		for _, b := range e.journal.Bindings {
			if b.WorkID == c.WorkID || b.WorkID == e.realID(c.WorkID) {
				source.ProjectRoot = b.Project
			}
		}
	}
	if source.ProjectRoot == "" {
		p, err := defaultViewWorkProject()
		if err != nil {
			return source, workInput("project: run inside the repository")
		}
		source.ProjectRoot = p
	}
	if !validateWorkSource(source) {
		return source, workInput("project")
	}
	return source, nil
}

// queue records a declaration the viewer cannot take now.
func (e *viewWorkEngine) queue(c viewWorkCommand, w io.Writer) error {
	entry := &viewWorkQueued{ID: workID("op-"), Action: c.Action, State: "queued"}
	stored := c
	stored.Options = management.Options{}
	switch c.Action {
	case "open":
		project, err := canonicalWorkDir(c.Project)
		if err != nil {
			return err
		}
		handle := queuedWorkHandle(project, c.Host, c.NativeSessionID, c.Key)
		if done, err := e.reopenKnown(c, project, handle, w); done {
			return err
		}
		if c.Change != "" {
			if stored.Change, err = canonicalWorkDir(c.Change); err != nil {
				return err
			}
		}
		stored.Project = project
		entry.Source = viewer.TimelineSource{ProjectRoot: project, Host: viewer.TimelineHost(c.Host), NativeSessionID: c.NativeSessionID}
		entry.WorkID = handle
		stored.QueuedOperation = entry.ID
	default:
		source, err := e.queueSource(c)
		if err != nil {
			return err
		}
		entry.Source, entry.WorkID = source, c.WorkID
		stored.WorkID = c.WorkID
		if e.queuedFor(entry.WorkID, e.journal) >= viewWorkQueueLimit {
			return viewWorkQueueFull()
		}
		if err := untitledUnknownTask(e.journal, c); err != nil {
			return err
		}
		if c.Action == "evidence" {
			if entry.Frozen, stored.Images, err = e.freezeImages(entry.ID, c); err != nil {
				return err
			}
		}
	}
	entry.Command = stored
	e.journal.Queue = append(e.journal.Queue, entry)
	// The limit is checked on the journal as it is on disk now, under its lock.
	if err := e.saveWith(func(m *viewWorkJournal) error {
		if e.queuedFor(entry.WorkID, m) > viewWorkQueueLimit {
			return viewWorkQueueFull()
		}
		return nil
	}); err != nil {
		e.removeQueued(entry.ID)
		e.dropFrozen(entry)
		return err
	}
	position := e.queuedFor(entry.WorkID, e.journal)
	if c.Action == "open" {
		if err := addCurrentWork(e.state, c.Host, c.NativeSessionID, entry.WorkID); err != nil {
			return err
		}
	}
	return writeViewWorkBoundedJSON(w, viewWorkQueuedResult{WorkID: entry.WorkID, Result: "queued", OperationID: entry.ID, Position: position})
}

// untitledUnknownTask rejects, while the viewer cannot say which tasks exist, a
// queued sync that names without a title a task this journal never saw
// registered for the work: sent later, the whole declaration would be rejected
// with it. A task another session registered asks for a title it does not
// need, which is harmless, since a title is used only to create a task.
func untitledUnknownTask(j *viewWorkJournal, c viewWorkCommand) error {
	if c.Action != "sync" {
		return nil
	}
	var known map[string]bool
	for i, t := range c.SyncTasks {
		if t.Title != "" {
			continue
		}
		if known == nil {
			known = knownTaskKeys(j, c.WorkID)
		}
		if !known[t.Key] {
			return workInput("invalid semantic JSON: " + faultf(fmt.Sprintf("tasks[%d].title", i), "required field").Error())
		}
	}
	return nil
}

// knownTaskKeys lists the task keys this journal saw registered for a work:
// its recorded and pending intents, and its queued declarations that create
// tasks.
func knownTaskKeys(j *viewWorkJournal, workID string) map[string]bool {
	scope, keys := workScopeOf(j, workID), map[string]bool{}
	add := func(c viewWorkCommand) {
		for _, t := range c.Tasks {
			keys[t.Key] = true
		}
		for _, t := range c.SyncTasks {
			if t.Title != "" {
				keys[t.Key] = true
			}
		}
	}
	for _, intents := range [][]*viewWorkIntent{j.History, j.Pending} {
		for _, i := range intents {
			var context struct {
				Command viewWorkCommand `json:"command"`
			}
			if workScopeOf(j, i.WorkID) == scope && json.Unmarshal(i.Context, &context) == nil {
				add(context.Command)
			}
		}
	}
	for _, q := range j.Queue {
		if workScopeOf(j, q.WorkID) == scope {
			add(q.Command)
		}
	}
	return keys
}

func viewWorkQueueFull() error {
	return &viewWorkError{Code: "publisher_capacity", Next: "start the viewer with `hive view start` so the queue is sent"}
}

// queuedFor counts the entries queued for one work, whichever name they use.
func (e *viewWorkEngine) queuedFor(id string, j *viewWorkJournal) int {
	scope, n := workScopeOf(j, id), 0
	for _, q := range j.Queue {
		if workScopeOf(j, q.WorkID) == scope {
			n++
		}
	}
	return n
}

// reopenKnown answers an open whose key already names a work of this session,
// queued or opened, instead of queueing a second open for it.
func (e *viewWorkEngine) reopenKnown(c viewWorkCommand, project, handle string, w io.Writer) (bool, error) {
	for _, q := range e.journal.Queue {
		if q.Action == "open" && q.WorkID == handle {
			return true, writeViewWorkBoundedJSON(w, viewWorkQueuedResult{WorkID: handle, Result: "queued", OperationID: q.ID, Position: e.queuedFor(handle, e.journal)})
		}
	}
	real := ""
	for _, a := range e.journal.Aliases {
		if a.Alias == handle {
			real = a.WorkID
		}
	}
	if real == "" {
		for _, b := range e.journal.Bindings {
			if b.Key == c.Key && b.Project == project && string(b.Source.Host) == c.Host && b.Source.NativeSessionID == c.NativeSessionID {
				real = b.WorkID
			}
		}
	}
	if real == "" {
		return false, nil
	}
	if err := addCurrentWork(e.state, c.Host, c.NativeSessionID, real); err != nil {
		return true, err
	}
	return true, writeViewWorkBoundedJSON(w, viewWorkResult{WorkID: real, Result: "opened"})
}

// noteHandle remembers which work a key opened in this session, so that the
// key resolves to it while the viewer is unavailable.
func (e *viewWorkEngine) noteHandle(c viewWorkCommand, real string) error {
	project, err := canonicalWorkDir(c.Project)
	if err != nil || c.Key == "" || c.Host == "" {
		return nil
	}
	handle := queuedWorkHandle(project, c.Host, c.NativeSessionID, c.Key)
	if handle == real {
		return nil
	}
	for _, a := range e.journal.Aliases {
		if a.Alias == handle {
			return nil
		}
	}
	e.journal.Aliases = append(e.journal.Aliases, viewWorkAlias{Alias: handle, WorkID: real})
	return e.save()
}

// pruneAliases forgets the handle of the engine's work once no queue entry and
// no session's current work refers to the handle or to the work it names.
func (e *viewWorkEngine) pruneAliases() {
	if e.scope == "" || e.scope == noWorkScope {
		return
	}
	refs, err := currentWorkRefs(e.state)
	if err != nil {
		return
	}
	kept := make([]viewWorkAlias, 0, len(e.journal.Aliases))
	for _, a := range e.journal.Aliases {
		if a.Alias == e.scope && !refs[a.Alias] && !refs[a.WorkID] && !e.hasQueueEntryFor(a.Alias) {
			continue
		}
		kept = append(kept, a)
	}
	if len(kept) != len(e.journal.Aliases) {
		e.journal.Aliases = kept
		e.save()
	}
}

func (e *viewWorkEngine) hasQueueEntryFor(handle string) bool {
	for _, q := range e.journal.Queue {
		if q.WorkID == handle {
			return true
		}
	}
	return false
}

// freezeImages copies the captures into the publisher folder so that their
// bytes cannot change before they are sent. The returned specs name the copies
// relative to that folder.
func (e *viewWorkEngine) freezeImages(id string, c viewWorkCommand) ([]string, []string, error) {
	images, err := readViewWorkImages(c)
	if err != nil {
		return nil, nil, err
	}
	var names, specs []string
	for n, image := range images {
		name := "queue-" + id + "-" + itoa(n)
		if err = writePrivateWorkFile(e.dir, filepath.Join(e.dir, name), image.Data); err != nil {
			e.removeFrozen(names)
			return nil, nil, err
		}
		names = append(names, name)
		specs = append(specs, name+":"+strings.ReplaceAll(image.Caption, ":", "_"))
	}
	return names, specs, nil
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func (e *viewWorkEngine) removeFrozen(names []string) {
	for _, name := range names {
		os.Remove(filepath.Join(e.dir, name))
	}
}
func (e *viewWorkEngine) dropFrozen(q *viewWorkQueued) { e.removeFrozen(q.Frozen) }

func (e *viewWorkEngine) removeQueued(id string) {
	kept := e.journal.Queue[:0]
	for _, q := range e.journal.Queue {
		if q.ID != id {
			kept = append(kept, q)
		}
	}
	e.journal.Queue = kept
}

// viewWorkLocalRejection: a declaration refused because what it declares no
// longer applies, as opposed to a failure that may succeed when retried.
func viewWorkRejectionCode(err error) (string, bool) {
	var contract *viewer.TimelineContractError
	if errors.As(err, &contract) {
		// The viewer's own authority rule, applied before anything is sent.
		if contract.Field == "actor" && (contract.Code == "lead_required" || contract.Code == "producer_not_accepted") {
			return contract.Code, true
		}
		return "", false
	}
	var ve *viewWorkError
	if !errors.As(err, &ve) {
		return "", false
	}
	if definitiveWorkRejection(ve.Code) {
		return ve.Code, true
	}
	switch ve.Code {
	case "session_closed", "stage_start_required", "explicit_stage_predecessor_required", "previous_stage_only_for_reentry",
		"stage_active", "active_stage_required", "task_not_found", "task_segment_not_found", "task_title_required", "stage_not_started",
		"supersedes_view_not_found", "confirmed_plan_revision_required", "task_metadata_conflict", "stage_already_terminal", "binding_mismatch",
		// Authority: this session may not act on the work, whatever it declared.
		"accepted_producer_required":
		return ve.Code, true
	}
	return "", false
}

// flushQueue sends the queued declarations of the engine's work in order. It
// stops at the first failure that is not a rejection, which leaves that entry
// and the ones after it queued.
func (e *viewWorkEngine) flushQueue(c viewWorkCommand) error {
	if e.scope == "" || e.scope == noWorkScope {
		return nil
	}
	for _, q := range append([]*viewWorkQueued{}, e.journal.Queue...) {
		if e.scopeOf(q.WorkID) != e.scope {
			continue
		}
		// Each session sends, and hears about, only the declarations it queued.
		if string(q.Source.Host) != c.Host || q.Source.NativeSessionID != c.NativeSessionID {
			continue
		}
		if q.State == "rejected" {
			e.noteRejected(q)
			continue
		}
		err := e.flushEntry(q, c)
		if err == nil {
			e.dropFrozen(q)
			e.removeQueued(q.ID)
			e.report.Sent++
			if err = e.save(); err != nil {
				return err
			}
			continue
		}
		code, rejected := viewWorkRejectionCode(err)
		if !rejected {
			return err
		}
		q.State, q.Reason = "rejected", code
		var failure *viewWorkError
		if errors.As(err, &failure) && failure.Pending != "" {
			e.archiveFlushRejection(failure.Pending, q, code)
		}
		e.dropFrozen(q)
		q.Frozen = nil
		e.noteRejected(q)
		if err = e.save(); err != nil {
			return err
		}
		// A rejected open never became a work: its local handle must not stay
		// current. A work the viewer reports closed or gone is not current either.
		if err = e.releaseRejectedWork(q, code, c); err != nil {
			return err
		}
	}
	return nil
}

// releaseRejectedWork drops the work a rejected entry named from the session's
// current works: the local handle of a rejected open, and the real work (which
// the handle may stand for) once the viewer reports it closed or gone.
func (e *viewWorkEngine) releaseRejectedWork(q *viewWorkQueued, code string, c viewWorkCommand) error {
	gone := code == "session_closed" || code == "work_closed" || code == "session_deleted" || code == "not_found"
	if q.Action != "open" && !gone {
		return nil
	}
	if err := releaseCurrentWork(e.state, c.Host, c.NativeSessionID, q.WorkID); err != nil {
		return err
	}
	if real := e.realID(q.WorkID); gone && real != q.WorkID {
		return releaseCurrentWork(e.state, c.Host, c.NativeSessionID, real)
	}
	return nil
}

// archiveFlushRejection archives the intent a queued declaration left behind
// when the viewer definitively rejected it, as `reconcile --action
// archive-rejected` would, so later lifecycle calls are not blocked. An intent
// with any unknown outcome stays pending.
func (e *viewWorkEngine) archiveFlushRejection(id string, q *viewWorkQueued, code string) {
	for _, i := range e.journal.Pending {
		if i.ID == id && i.State == "rejected_conflict" {
			e.journal.archiveRejected(id, "queued declaration "+q.ID+" rejected: "+code)
			return
		}
	}
}

func (e *viewWorkEngine) hasQueuedOpen(handle string) bool {
	for _, q := range e.journal.Queue {
		if q.Action == "open" && q.WorkID == handle {
			return true
		}
	}
	return false
}

func (e *viewWorkEngine) noteRejected(q *viewWorkQueued) {
	e.rejectedIDs = append(e.rejectedIDs, q.ID)
	if len(e.report.Rejected) >= 5 {
		e.report.RejectedOmitted++
		return
	}
	e.report.Rejected = append(e.report.Rejected, viewWorkQueueRejection{ID: q.ID, Action: q.Action, Reason: q.Reason})
}

// queueReport is nil when the queue had nothing to say.
func (e *viewWorkEngine) queueReport() *viewWorkQueueReport {
	if e.report.Sent == 0 && len(e.report.Rejected) == 0 && e.report.RejectedOmitted == 0 {
		return nil
	}
	r := e.report
	return &r
}

// ackQueue forgets the rejections a command has reported.
func (e *viewWorkEngine) ackQueue() error {
	if len(e.rejectedIDs) == 0 {
		return nil
	}
	for _, id := range e.rejectedIDs {
		e.removeQueued(id)
	}
	e.rejectedIDs = nil
	return e.save()
}

func (e *viewWorkEngine) flushEntry(q *viewWorkQueued, flusher viewWorkCommand) error {
	c := q.Command
	c.Options = flusher.Options
	started := q.State == "started"
	id := e.realID(q.WorkID)
	if q.Action != "open" {
		c.WorkID = id
		if id == q.WorkID && e.hasQueuedOpen(id) {
			// The open that would have created this work did not apply.
			return workStateError("session_closed")
		}
	}
	if !started {
		if q.Action != "open" {
			if rejection := e.queuePrecheck(q, c); rejection != nil {
				return rejection
			}
		}
		q.State = "started"
		if err := e.save(); err != nil {
			return err
		}
		viewWorkCheckpoint("started")
	} else if q.Action != "open" {
		if err := e.resumePending(c); err != nil {
			return err
		}
	}
	switch q.Action {
	case "open":
		e.registerRoot(c)
		s, err := e.open(c)
		if err != nil {
			return err
		}
		if q.WorkID != s.ViewerSessionID {
			e.journal.Aliases = append(e.journal.Aliases, viewWorkAlias{Alias: q.WorkID, WorkID: s.ViewerSessionID})
			if err = e.save(); err != nil {
				return err
			}
		}
		if err = addCurrentWork(e.state, c.Host, c.NativeSessionID, s.ViewerSessionID); err != nil {
			return err
		}
		if q.WorkID != s.ViewerSessionID {
			return releaseCurrentWork(e.state, c.Host, c.NativeSessionID, q.WorkID)
		}
		return nil
	case "sync":
		r, err := e.sync(c)
		if err == nil && r.State == "closed" {
			err = releaseCurrentWork(e.state, c.Host, c.NativeSessionID, c.WorkID)
		}
		return err
	default:
		images := make([]string, len(c.Images))
		for n, spec := range c.Images {
			images[n] = filepath.Join(e.dir, spec)
		}
		c.Images = images
		_, err := e.evidence(c)
		return err
	}
}

// queuePrecheck rejects, before anything is sent, a declaration that the
// work's current state has moved past: the work ended, or the stage it
// declares was already ended by someone else.
func (e *viewWorkEngine) queuePrecheck(q *viewWorkQueued, c viewWorkCommand) error {
	s, err := e.session(c.WorkID)
	if err != nil {
		return err
	}
	if s.State == viewer.TimelineClosed || s.State == viewer.TimelineFailed {
		return workStateError("session_closed")
	}
	if q.Action == "sync" {
		latest := workLifecycleStage(s, c.Stage)
		if latest != nil && (latest.StageState == "completed" || latest.StageState == "failed") && c.PreviousStage == "" {
			return &viewWorkError{Code: "stage_already_terminal", Field: c.Stage, Next: "the stage was already ended; declare a new occurrence with --previous-stage"}
		}
	}
	return nil
}

// resumePending replays, unchanged, the pending intents a killed flush left
// for the work.
func (e *viewWorkEngine) resumePending(c viewWorkCommand) error {
	for _, i := range append([]*viewWorkIntent{}, e.journal.Pending...) {
		if i.WorkID != c.WorkID || i.State == "rejected_conflict" || string(i.Source.Host) != c.Host || i.Source.NativeSessionID != c.NativeSessionID {
			continue
		}
		switch i.Action {
		case "evidence":
			if err := e.resumeEvidenceIntent(i); err != nil {
				return err
			}
		case "stage", "tasks register", "task update", "close":
			if _, err := e.resumeLifecycle(i); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *viewWorkEngine) registerRoot(c viewWorkCommand) {
	if root, err := canonicalWorkDir(c.Project); err == nil {
		viewer.RegisterRepository(e.state, root)
	}
}

// resumeEvidenceIntent replays the exact requests of an interrupted evidence
// step. Running the evidence command again completes what follows.
func (e *viewWorkEngine) resumeEvidenceIntent(i *viewWorkIntent) error {
	for n := range i.Steps {
		if _, err := e.executeStep(i, n); err != nil {
			return err
		}
	}
	return e.finishIntent(i)
}

// ---- journal merge ----

func workBindingKey(b viewWorkBinding) string {
	return "b|" + b.WorkID + "|" + string(b.Source.Host) + "|" + b.Source.NativeSessionID
}

// workJournalElements marks every element of the journal by identity, so a
// save can tell what this process changed.
func workJournalElements(j *viewWorkJournal) map[string][]byte {
	m := map[string][]byte{}
	put := func(key string, v any) { m[key], _ = json.Marshal(v) }
	for _, b := range j.Bindings {
		put(workBindingKey(b), b)
	}
	for _, i := range j.Pending {
		put("p|"+i.ID, i)
	}
	for _, i := range j.History {
		put("h|"+i.ID, i)
	}
	for _, q := range j.Queue {
		put("q|"+q.ID, q)
	}
	for _, a := range j.Aliases {
		put("a|"+a.Alias, a)
	}
	return m
}

func mergeWorkSet[T any](disk, mem []T, prefix string, key func(T) string, base, cur map[string][]byte) []T {
	out := append([]T{}, disk...)
	index := map[string]int{}
	for n, v := range out {
		index[key(v)] = n
	}
	for _, v := range mem {
		k := key(v)
		if b, ok := base[k]; ok && bytes.Equal(b, cur[k]) {
			continue
		}
		if n, ok := index[k]; ok {
			out[n] = v
		} else {
			index[k] = len(out)
			out = append(out, v)
		}
	}
	removed := map[string]bool{}
	for k := range base {
		if _, ok := cur[k]; !ok && strings.HasPrefix(k, prefix) {
			removed[k] = true
		}
	}
	if len(removed) == 0 {
		return out
	}
	kept := out[:0]
	for _, v := range out {
		if !removed[key(v)] {
			kept = append(kept, v)
		}
	}
	return kept
}

// mergeWorkJournal applies what this process changed since it read the journal
// onto the journal as it is on disk now, which other processes, working on
// other works, may have changed meanwhile.
func mergeWorkJournal(disk, mem *viewWorkJournal, base map[string][]byte) *viewWorkJournal {
	cur := workJournalElements(mem)
	out := *disk
	out.Bindings = mergeWorkSet(disk.Bindings, mem.Bindings, "b|", workBindingKey, base, cur)
	out.Pending = mergeWorkSet(disk.Pending, mem.Pending, "p|", func(i *viewWorkIntent) string { return "p|" + i.ID }, base, cur)
	out.History = mergeWorkSet(disk.History, mem.History, "h|", func(i *viewWorkIntent) string { return "h|" + i.ID }, base, cur)
	out.Queue = mergeWorkSet(disk.Queue, mem.Queue, "q|", func(q *viewWorkQueued) string { return "q|" + q.ID }, base, cur)
	out.Aliases = mergeWorkSet(disk.Aliases, mem.Aliases, "a|", func(a viewWorkAlias) string { return "a|" + a.Alias }, base, cur)
	return &out
}

func validateWorkQueue(j *viewWorkJournal) error {
	bad := func() error { return workStateError("invalid_journal") }
	ids := map[string]bool{}
	for _, q := range j.Queue {
		if q == nil || !viewTimelineObjectID.MatchString(q.ID) || ids[q.ID] || !viewTimelineSessionID.MatchString(q.WorkID) || !validateWorkSource(q.Source) {
			return bad()
		}
		ids[q.ID] = true
		switch q.Action {
		case "open", "sync", "evidence":
		default:
			return bad()
		}
		switch q.State {
		case "queued", "started", "rejected":
		default:
			return bad()
		}
		if q.Command.Action != q.Action || q.Command.Host != string(q.Source.Host) || q.Command.NativeSessionID != q.Source.NativeSessionID {
			return bad()
		}
		for _, name := range q.Frozen {
			if !viewWorkFrozenName.MatchString(name) {
				return bad()
			}
		}
		for _, spec := range q.Command.Images {
			if q.Action != "evidence" || !strings.HasPrefix(spec, "queue-") || strings.ContainsAny(strings.SplitN(spec, ":", 2)[0], "/\\") {
				return bad()
			}
		}
	}
	seen := map[string]bool{}
	for _, a := range j.Aliases {
		if !viewTimelineSessionID.MatchString(a.Alias) || !viewTimelineSessionID.MatchString(a.WorkID) || seen[a.Alias] {
			return bad()
		}
		seen[a.Alias] = true
	}
	return nil
}
