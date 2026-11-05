package viewer

import (
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Host events are facts a coding host reports through the local hive hook. They
// are maintenance metadata and a mirror of the host's native task list, never an
// agent publication: they take no session version, no receipt among the agent's
// operation receipts and no timeline event, so they cannot invalidate an agent's
// expected version or consume the event budget. Their replay record lives in the
// work's maintenance metadata, apart from the agent's operation receipts, so a
// host operation ID can never collide with an agent one.
const (
	// timelineMaxHostOperations bounds the remembered host operation IDs per work.
	timelineMaxHostOperations = 32
	// timelineHostPlanOrderBase keeps mirrored tasks clear of the small plan orders
	// an agent registers later.
	timelineHostPlanOrderBase int64 = 1_000_000
)

const (
	hostEventApplied      = "applied"
	hostEventReplayed     = "replayed"
	hostEventActivityOnly = "activity_only"
	hostEventNoWork       = "no_work"
)

func (s *Server) hostEvents(w http.ResponseWriter, r *http.Request) {
	if !allow(w, r, http.MethodPost) {
		return
	}
	// A local process calls this route, never the browser: the token is required and
	// any Origin header is refused, the viewer's own included.
	if !s.tokenOK(r) || r.Header.Get("Origin") != "" {
		httpError(w, http.StatusForbidden, "forbidden")
		return
	}
	raw, err := timelineReadBody(w, r)
	if err != nil {
		timelineHTTPError(w, err)
		return
	}
	event, err := DecodeTimelineHostEvent(raw)
	if err != nil {
		timelineHTTPError(w, err)
		return
	}
	result, err := s.store.TimelineHostEvent(event)
	if err != nil {
		timelineHTTPError(w, err)
		return
	}
	writeJSON(w, result)
}

// timelineHostProject resolves the registered project an event's directory
// belongs to: the project root of the directory, or the nearest registered
// ancestor when the directory is not itself a project root.
//
// It is split like TimelineCreate's resolution (timelineLocate, timelineKeyOf).
// timelineHostCandidates runs without the store lock held: it spawns Git
// (ResolveRoot, the remote) and reads the registry once for all the candidate
// directories. timelineHostKey runs under the lock, after the reconcile, and
// reads the keys the stored works carry at that moment.
type timelineHostCandidate struct {
	candidate timelineLocated
	err       error
}

func (s *Store) timelineHostCandidates(source TimelineSource) ([]timelineHostCandidate, error) {
	dirs := []string{}
	seen := map[string]bool{}
	add := func(dir string) {
		if !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}
	if root, err := ResolveRoot(source.ProjectRoot); err == nil {
		add(root)
	}
	for p := filepath.Clean(source.ProjectRoot); ; p = filepath.Dir(p) {
		add(p)
		if p == filepath.Dir(p) {
			break
		}
	}
	registry, err := readRegistry(s.vdir)
	if err != nil {
		return nil, timelineFailure("project_unavailable", 409)
	}
	out := make([]timelineHostCandidate, 0, len(dirs))
	for _, dir := range dirs {
		c, err := s.timelineLocate(registry, dir)
		out = append(out, timelineHostCandidate{candidate: c, err: err})
		if err == nil {
			break // the nearest registered directory decides; farther ones are not located
		}
	}
	return out, nil
}

// timelineHostKey is the key of the first candidate, nearest first, that
// resolves. When none does it returns the error of the nearest one, the most
// specific cause. The caller holds s.mu.
func (s *Store) timelineHostKey(candidates []timelineHostCandidate) (string, error) {
	var first error
	for _, c := range candidates {
		err := c.err
		if err == nil {
			var key string
			if key, err = s.timelineKeyOf(c.candidate); err == nil {
				return key, nil
			}
		}
		if first == nil {
			first = err
		}
	}
	return "", first
}

func timelineHostMember(v *TimelineSession, source TimelineSource) bool {
	for _, p := range v.Producers {
		if timelineNativeSourceIdentity(p.Source) == timelineNativeSourceIdentity(source) {
			return true
		}
	}
	return false
}

func timelineTerminal(v *TimelineSession) bool {
	return v.State == TimelineClosed || v.State == TimelineFailed
}

// TimelineHostEvent applies one validated host event. With a work ID it acts as
// the accepted native producer of that work, in the project the work belongs to;
// every event also moves the last activity of the other open, unarchived works of
// its session to the event's own time, clamped to the server's clock.
func (s *Store) TimelineHostEvent(e TimelineHostEvent) (TimelineHostEventResult, error) {
	if err := e.Validate(); err != nil {
		return TimelineHostEventResult{}, err
	}
	op := e.OperationID()
	result := TimelineHostEventResult{OperationID: op, Status: hostEventNoWork}
	// The directories are resolved before the store lock is taken (it spawns Git);
	// the key the stored works bind is read under the lock, so a bind made in
	// between is seen.
	candidates, projectErr := s.timelineHostCandidates(e.Source)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.timelineReconcile(); err != nil {
		return result, err
	}
	key := ""
	if projectErr == nil {
		key, projectErr = s.timelineHostKey(candidates)
	}
	if e.WorkID != "" {
		old, ok := s.timeline[e.WorkID]
		if !ok {
			return result, ErrNotFound
		}
		if old.Session == nil || old.DeletePending {
			return result, timelineFailure("session_deleted", 410)
		}
		if projectErr != nil {
			return result, projectErr
		}
		if old.ProjectKey != key {
			return result, timelineFailure("foreign_project", 409)
		}
		checked := e
		if timelineTerminal(old.Session) || e.StageEventID != "" && (old.Session.CurrentStageEventID == nil || *old.Session.CurrentStageEventID != e.StageEventID) {
			// A finished work, or a stage that is no longer active, has nothing to bind
			// to: the event still proves host activity, so only its activity moves.
			checked.StageEventID = ""
		}
		actor, err := ValidateTimelineHostEventSession(*old.Session, checked)
		if err != nil {
			return result, err
		}
		status, err := s.timelineHostApply(old, e, op, actor.ProducerID, true)
		if err != nil {
			return result, err
		}
		result.Status = status
	}
	if projectErr != nil {
		return result, nil
	}
	ids := make([]string, 0, len(s.timeline))
	for id := range s.timeline {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		old := s.timeline[id]
		// Only open, unarchived works take an unnamed event: a finished or archived
		// work is rewritten only when it is named as the session's current one.
		if id == e.WorkID || old.Session == nil || old.DeletePending || old.ProjectKey != key || timelineTerminal(old.Session) || old.Session.Archived || !timelineHostMember(old.Session, e.Source) {
			continue
		}
		status, err := s.timelineHostApply(old, e, op, "", false)
		if err != nil {
			return result, err
		}
		if result.Status == hostEventNoWork || result.Status == hostEventReplayed && status != hostEventReplayed {
			result.Status = status
		}
	}
	return result, nil
}

// timelineHostApply moves last activity and, for the named work, mirrors the
// native task list. It never changes the session version, stages, authority or
// work state.
func (s *Store) timelineHostApply(old *timelineRecord, e TimelineHostEvent, op, producerID string, named bool) (string, error) {
	if old.Maintenance != nil && contains(old.Maintenance.HostOperations, op) {
		return hostEventReplayed, nil
	}
	r := timelineClone(old)
	now := s.timelineTime()
	timelineRefreshMaintenance(r, now)
	m, v := r.Maintenance, r.Session
	// The host's own time stamps the activity, so a reconcile can backfill what
	// happened earlier; it never moves ahead of the server's clock or backwards.
	activity := now
	if occurred, err := time.Parse(time.RFC3339Nano, e.OccurredAt); err == nil {
		if nowAt, _ := time.Parse(time.RFC3339Nano, now); occurred.Before(nowAt) {
			activity = occurred.UTC().Format(time.RFC3339Nano)
		}
	}
	m.LastActivityAt = timelineLater(m.LastActivityAt, activity)
	m.HostActivityAt = timelineLater(m.HostActivityAt, activity)
	v.UpdatedAt = timelineLater(v.UpdatedAt, now)
	status := hostEventApplied
	task := strings.HasPrefix(e.EventType, "task_")
	switch {
	case timelineTerminal(v):
		status = hostEventActivityOnly
	case named && e.StageEventID != "" && (v.CurrentStageEventID == nil || e.StageEventID != *v.CurrentStageEventID):
		// The stage the event names is no longer active: nothing is mirrored.
		status = hostEventActivityOnly
	case named && task && hostTaskBeforeWork(e, v):
		// A native task from before this work existed belongs to earlier work.
		status = hostEventActivityOnly
	case named && task && v.CurrentStageEventID != nil && e.StageEventID == *v.CurrentStageEventID:
		occurred, _ := time.Parse(time.RFC3339Nano, e.OccurredAt)
		at := occurred.UnixNano()
		if last, seen := m.HostTaskTimes[e.Data.TaskID]; seen && at < last {
			// A replay or a late event older than the last one applied to this task
			// never moves it backwards.
			status = hostEventActivityOnly
			break
		}
		changed, err := timelineMirrorHostTask(v, e, producerID)
		if err != nil {
			return "", err
		}
		if changed {
			m.ContentBaseline = nil
		}
		if hostTaskMirrored(v, e.Data.TaskID) {
			if m.HostTaskTimes == nil {
				m.HostTaskTimes = map[string]int64{}
			}
			m.HostTaskTimes[e.Data.TaskID] = max(at, m.HostTaskTimes[e.Data.TaskID])
		}
	}
	m.HostOperations = append(m.HostOperations, op)
	if extra := len(m.HostOperations) - timelineMaxHostOperations; extra > 0 {
		m.HostOperations = append([]string{}, m.HostOperations[extra:]...)
	}
	timelineProjectMaintenance(r)
	if err := s.timelineSave(r); err != nil {
		return "", err
	}
	return status, nil
}

// hostTaskBeforeWork says whether a task event happened before the work was
// created, by the host's own time.
func hostTaskBeforeWork(e TimelineHostEvent, v *TimelineSession) bool {
	occurred, err1 := time.Parse(time.RFC3339Nano, e.OccurredAt)
	created, err2 := time.Parse(time.RFC3339Nano, v.CreatedAt)
	return err1 == nil && err2 == nil && occurred.Before(created)
}

func hostTaskMirrored(v *TimelineSession, nativeTaskID string) bool {
	key := TimelineHostTaskKey(nativeTaskID)
	for _, t := range v.Tasks {
		if t.TaskKey == key {
			return true
		}
	}
	return false
}

// timelineMirrorHostTask reflects one native task event as a host:<id> task bound
// to the active stage. A task is added only while every task of the work is
// itself a mirrored one and up to TimelineMaxHostTasks of them; one already
// mirrored keeps following the host after the agent registers tasks.
func timelineMirrorHostTask(v *TimelineSession, e TimelineHostEvent, producerID string) (bool, error) {
	// Once the agent registers tasks, the host tasks already mirrored keep
	// following the host's state, but no new one is added.
	agentTasks := false
	for _, t := range v.Tasks {
		agentTasks = agentTasks || !strings.HasPrefix(t.TaskKey, TimelineHostTaskKeyPrefix)
	}
	var stage *TimelineEvent
	for i := range v.Events {
		if v.Events[i].EventID == *v.CurrentStageEventID && v.Events[i].Kind == "stage" {
			stage = &v.Events[i]
		}
	}
	if stage == nil {
		return false, nil
	}
	d := e.Data
	key := TimelineHostTaskKey(d.TaskID)
	title := []rune(strings.TrimSpace(d.Title))
	if len(title) > 120 {
		title = title[:120]
	}
	for i := range v.Tasks {
		t := &v.Tasks[i]
		if t.TaskKey != key {
			continue
		}
		if t.Withdrawn || len(t.Segments) != 1 {
			return false, nil
		}
		seg := &t.Segments[0]
		changed := false
		if len(title) > 0 && t.Title != string(title) {
			t.Title, changed = string(title), true
		}
		next := seg.State
		switch {
		case e.EventType == "task_completed":
			next = "completed"
		case e.EventType == "task_updated" && d.State != "":
			next = d.State
		}
		if next != seg.State {
			seg.State, changed = next, true
		}
		if !changed {
			return false, nil
		}
		if t.TaskVersion >= TimelineMaxNumber {
			return false, timelineFailure("capacity", 413)
		}
		t.TaskVersion++
		t.State = AggregateTimelineTaskState(t.Segments)
		return true, nil
	}
	if agentTasks || len(v.Tasks) >= TimelineMaxHostTasks || len(v.Tasks) >= TimelineMaxTasks {
		return false, nil
	}
	state := "not_started"
	switch {
	case e.EventType == "task_completed":
		state = "completed"
	case d.State != "":
		state = d.State
	case e.EventType == "task_updated":
		state = "in_progress"
	}
	name := string(title)
	if name == "" {
		name = "Native task " + d.TaskID
	}
	order := timelineHostPlanOrderBase
	for _, t := range v.Tasks {
		order = max(order, t.PlanOrder+1)
	}
	taskID, err := timelineRandomID("task-")
	if err != nil {
		return false, err
	}
	segmentID, err := timelineRandomID("segment-")
	if err != nil {
		return false, err
	}
	v.Tasks = append(v.Tasks, TimelineTask{
		PlanOrder: order, TaskID: taskID, TaskKey: key, Title: name, TaskVersion: 1, State: state,
		Segments: []TimelineTaskSegment{{SegmentID: segmentID, PlannedStage: stage.Stage, Order: order, State: state, AssignedProducerID: producerID, DisplayStageEventID: stage.EventID}},
	})
	return true, nil
}
