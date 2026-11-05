package viewer

import (
	"net/http"
	"path/filepath"
	"sort"
	"strings"
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
func (s *Store) timelineHostProject(source TimelineSource) (string, error) {
	var last error
	candidates := []string{}
	if root, err := ResolveRoot(source.ProjectRoot); err == nil {
		candidates = append(candidates, root)
	}
	for p := filepath.Clean(source.ProjectRoot); ; p = filepath.Dir(p) {
		candidates = append(candidates, p)
		if p == filepath.Dir(p) {
			break
		}
	}
	for _, dir := range candidates {
		probe := source
		probe.ProjectRoot = dir
		_, key, _, err := s.timelineProject(probe)
		if err == nil {
			return key, nil
		}
		if last == nil {
			last = err
		}
	}
	return "", last
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
// every event also moves the last activity of the other works of its session.
func (s *Store) TimelineHostEvent(e TimelineHostEvent) (TimelineHostEventResult, error) {
	if err := e.Validate(); err != nil {
		return TimelineHostEventResult{}, err
	}
	op := e.OperationID()
	result := TimelineHostEventResult{OperationID: op, Status: hostEventNoWork}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.timelineReconcile(); err != nil {
		return result, err
	}
	key, projectErr := s.timelineHostProject(e.Source)
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
		if timelineTerminal(old.Session) {
			// A finished work has no active stage to bind to; only its activity moves.
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
		if id == e.WorkID || old.Session == nil || old.DeletePending || old.ProjectKey != key || !timelineHostMember(old.Session, e.Source) {
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
	m.LastActivityAt = timelineLater(m.LastActivityAt, now)
	v.UpdatedAt = timelineLater(v.UpdatedAt, now)
	status := hostEventApplied
	if timelineTerminal(v) {
		status = hostEventActivityOnly
	} else if named && strings.HasPrefix(e.EventType, "task_") && v.CurrentStageEventID != nil && e.StageEventID == *v.CurrentStageEventID {
		changed, err := timelineMirrorHostTask(v, e, producerID)
		if err != nil {
			return "", err
		}
		if changed {
			m.ContentBaseline = nil
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

// timelineMirrorHostTask reflects one native task event as a host:<id> task bound
// to the active stage, only while every task of the work is itself a mirrored
// one and up to TimelineMaxHostTasks of them.
func timelineMirrorHostTask(v *TimelineSession, e TimelineHostEvent, producerID string) (bool, error) {
	for _, t := range v.Tasks {
		if !strings.HasPrefix(t.TaskKey, TimelineHostTaskKeyPrefix) {
			return false, nil
		}
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
	if len(v.Tasks) >= TimelineMaxHostTasks || len(v.Tasks) >= TimelineMaxTasks {
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
