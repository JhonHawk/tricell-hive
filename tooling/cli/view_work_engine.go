package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

type viewWorkEngine struct {
	state   string
	dir     string
	info    serverInfo
	journal *viewWorkJournal
	lock    *viewer.FileLock
}

func canonicalWorkDir(path string) (string, error) {
	p, err := filepath.Abs(path)
	if err != nil {
		return "", workInput("directory")
	}
	p, err = filepath.EvalSymlinks(p)
	if err != nil {
		return "", workInput("directory")
	}
	st, err := os.Stat(p)
	if err != nil || !st.IsDir() {
		return "", workInput("directory")
	}
	return p, nil
}
func openViewWorkEngine(c viewWorkCommand) (*viewWorkEngine, error) {
	_, state, err := management.NormalizeOptions(c.Options)
	if err != nil {
		return nil, err
	}
	state, err = filepath.Abs(state)
	if err != nil {
		return nil, err
	}
	// Every component is checked before descending; publisher state never shares
	// the server-owned viewer inventory or its long-lived lock.
	dir := state
	for _, part := range []string{"publisher", "v1"} {
		dir = filepath.Join(dir, part)
		if err = os.Mkdir(dir, 0700); err != nil && !os.IsExist(err) {
			return nil, err
		}
		st, e := os.Lstat(dir)
		if e != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 || st.Mode().Perm()&0077 != 0 {
			return nil, workStateError("unsafe_publisher_directory")
		}
	}
	lock, err := lockWorkPublisher(dir)
	if err != nil {
		return nil, err
	}
	e := &viewWorkEngine{state: state, dir: dir, lock: lock}
	e.journal, err = readWorkJournal(dir)
	if err != nil {
		e.Close()
		return nil, err
	}
	e.info, err = viewTimelineRunning(newViewPaths(state))
	if err != nil {
		e.Close()
		return nil, err
	}
	return e, nil
}
func (e *viewWorkEngine) Close() {
	if e.lock != nil {
		e.lock.Release()
	}
}
func (e *viewWorkEngine) save() error { return writeWorkJournal(e.dir, e.journal) }
func (e *viewWorkEngine) session(id string) (viewer.TimelineSession, error) {
	if !viewTimelineSessionID.MatchString(id) {
		return viewer.TimelineSession{}, workInput("work_id")
	}
	raw, err := viewTimelineHTTP(e.info, "GET", "/api/timeline/sessions/"+id, "", nil, viewer.TimelineMaxBody)
	if err != nil {
		return viewer.TimelineSession{}, workStateError("work_unavailable")
	}
	r, err := viewer.DecodeTimelineSessionResponse(raw)
	if err != nil || r.Session.ViewerSessionID != id {
		return viewer.TimelineSession{}, workStateError("invalid_response")
	}
	return r.Session, nil
}
func (e *viewWorkEngine) actor(s viewer.TimelineSession, c viewWorkCommand) (viewer.TimelineActor, error) {
	for _, p := range s.Producers {
		if string(p.Source.Host) == c.Host && p.Source.NativeSessionID == c.NativeSessionID {
			if c.Project != "" {
				root, err := canonicalWorkDir(c.Project)
				if err != nil || root != p.Source.ProjectRoot {
					continue
				}
			}
			return viewer.TimelineActor{Kind: "native_producer", ProducerID: p.ProducerID}, nil
		}
	}
	return viewer.TimelineActor{}, workStateError("accepted_producer_required")
}
func (e *viewWorkEngine) prepareIntent(action, id string, source viewer.TimelineSource, context any) (*viewWorkIntent, error) {
	for _, i := range e.journal.Pending {
		if i.WorkID == id && id != "" || i.Source == source && i.WorkID == "" && bytes.Equal(i.Context, mustWorkJSON(context)) {
			return nil, &viewWorkError{Code: "pending_operation", Pending: i.ID, Next: "run hive view work resume with the original native source"}
		}
	}
	if err := e.journal.admitWorkIntent(); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(context)
	if err != nil {
		return nil, err
	}
	i := &viewWorkIntent{ID: workID("pending-"), Action: action, WorkID: id, Source: source, State: "prepared", Context: raw, Steps: []viewWorkStep{}}
	e.journal.Pending = append(e.journal.Pending, i)
	if err = e.save(); err != nil {
		return nil, err
	}
	return i, nil
}

// appendStep freezes the complete wire representation, including multipart
// boundary and upload bytes, before the first request can leave this process.
func (e *viewWorkEngine) appendStep(i *viewWorkIntent, method, path, contentType string, body []byte) (int, error) {
	i.Steps = append(i.Steps, viewWorkStep{Method: method, Path: path, ContentType: contentType, Body: append([]byte(nil), body...)})
	return len(i.Steps) - 1, e.save()
}
func (e *viewWorkEngine) executeStep(i *viewWorkIntent, index int) ([]byte, error) {
	if index < 0 || index >= len(i.Steps) {
		return nil, workStateError("invalid_step")
	}
	s := &i.Steps[index]
	if len(s.Response) > 0 {
		if s.Method == "POST" && s.Path == "/api/timeline/sessions" {
			return workStoredCreateResponse(s.Response)
		}
		return s.Response, nil
	}
	if s.Rejection != "" {
		return nil, &viewWorkError{Code: s.Rejection, Pending: i.ID, Next: "inspect and reconcile the rejected intent"}
	}
	s.Attempted = true
	i.State = "needs_reconciliation"
	if err := e.save(); err != nil {
		return nil, err
	}
	raw, code, err := viewWorkHTTP(e.info, s.Method, s.Path, s.ContentType, s.Body)
	if err != nil {
		if definitiveWorkRejection(code) {
			s.Rejection = code
			i.State = "rejected_conflict"
		}
		if saveErr := e.save(); saveErr != nil {
			return nil, saveErr
		}
		if code == "" {
			code = "outcome_unknown"
		}
		return nil, &viewWorkError{Code: code, Pending: i.ID, Next: "preserve inputs; run hive view work resume with the same native source"}
	}
	// Validate before recording a successful receipt. A malformed success may
	// have committed remotely, and must therefore retain exact replay data.
	if s.Method == "POST" && s.Path == "/api/timeline/sessions" {
		r, err := viewer.DecodeTimelineSessionResponse(raw)
		var request viewer.TimelineCreateSessionRequest
		if err != nil || json.Unmarshal(s.Body, &request) != nil || r.Session.Source != request.Source || !sameWorkKey(r.Session.WorkKey, request.WorkKey) {
			return nil, &viewWorkError{Code: "outcome_unknown", Pending: i.ID, Next: "resume the unchanged request"}
		}
		r.ArtifactURLs = nil
		r.DocumentURLs = nil
		raw, _ = json.Marshal(r)
		i.WorkID = r.Session.ViewerSessionID
	} else {
		if r, err := viewer.DecodeTimelineOperationResult(raw); err != nil || r.ViewerSessionID != i.WorkID || r.OperationID != workStepOperationID(*s) {
			return nil, &viewWorkError{Code: "outcome_unknown", Pending: i.ID, Next: "resume the unchanged request"}
		}
	}
	s.Response = append([]byte(nil), raw...)
	if err = e.save(); err != nil {
		return nil, err
	}
	return raw, nil
}
func (e *viewWorkEngine) finishIntent(i *viewWorkIntent) error {
	i = compactConfirmedWorkIntent(i)
	for n, p := range e.journal.Pending {
		if p.ID == i.ID {
			e.journal.Pending = append(e.journal.Pending[:n], e.journal.Pending[n+1:]...)
			e.journal.History = append(e.journal.History, i)
			e.journal.compactConfirmedHistory()
			return e.save()
		}
	}
	return workStateError("pending_not_found")
}

// viewWorkHTTP mirrors the authenticated existing transport, preserving only a
// machine error code, never arbitrary server text or authentication material.
func viewWorkHTTP(info serverInfo, method, path, ct string, body []byte) ([]byte, string, error) {
	if !strings.HasPrefix(path, "/api/") || strings.ContainsAny(path, "?#\\") {
		return nil, "", errors.New("invalid route")
	}
	if method != http.MethodGet {
		if _, err := viewTimelineHTTP(info, http.MethodGet, "/api/health", "", nil, viewInfoLimit); err != nil {
			return nil, "reload_required", err
		}
	}
	req, err := http.NewRequest(method, info.url()+path, bytes.NewReader(body))
	if err != nil {
		return nil, "", errors.New("invalid request")
	}
	req.Header.Set("X-Hive-Token", info.Token)
	req.Header.Set(viewer.TimelineRevisionHeader, "2")
	req.Header.Set("Origin", info.url())
	req.Header.Set("X-Hive-Action-Scope", "session")
	req.Header.Set("Content-Type", ct)
	client := *viewClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", errWorkUncertain
	}
	defer resp.Body.Close()
	if resp.Header.Get(viewer.TimelineRevisionHeader) != "2" {
		return nil, "reload_required", errors.New("update Hive and reload viewer; preserve original operation inputs")
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, viewer.TimelineMaxResponseBody+1))
	if err != nil || len(raw) > viewer.TimelineMaxResponseBody {
		return nil, "", errWorkIncomplete
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var failure struct {
			Code string `json:"code"`
		}
		json.Unmarshal(raw, &failure)
		if !viewTimelineObjectID.MatchString(failure.Code) {
			failure.Code = ""
		}
		return nil, failure.Code, errors.New("request rejected")
	}
	return raw, "", nil
}
func (e *viewWorkEngine) binding(id, host, native string) *viewWorkBinding {
	for n := range e.journal.Bindings {
		b := &e.journal.Bindings[n]
		if b.WorkID == id && string(b.Source.Host) == host && b.Source.NativeSessionID == native {
			return b
		}
	}
	return nil
}
func (e *viewWorkEngine) bind(c viewWorkCommand, s viewer.TimelineSession, explicit bool) error {
	if _, err := e.actor(s, c); err != nil {
		return err
	}
	change, err := canonicalWorkDir(c.Change)
	if err != nil {
		return err
	}
	project := c.Project
	b := e.binding(s.ViewerSessionID, c.Host, c.NativeSessionID)
	if project == "" {
		if b != nil {
			project = b.Project
		} else {
			for _, p := range s.Producers {
				if string(p.Source.Host) == c.Host && p.Source.NativeSessionID == c.NativeSessionID {
					project = p.Source.ProjectRoot
					break
				}
			}
		}
	}
	project, err = canonicalWorkDir(project)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(project, change)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return workInput("change must be within project")
	}
	for _, other := range e.journal.Bindings {
		if other.Change == change && other.Source.Host == viewer.TimelineHost(c.Host) && other.Source.NativeSessionID == c.NativeSessionID && other.WorkID != s.ViewerSessionID {
			return workStateError("directory_already_bound")
		}
	}
	if explicit {
		for _, i := range e.journal.Pending {
			if i.WorkID == s.ViewerSessionID {
				return &viewWorkError{Code: "pending_operation", Pending: i.ID}
			}
		}
	}
	if b != nil && b.Change != change {
		if !explicit {
			return workStateError("binding_mismatch")
		}
		for _, i := range e.journal.Pending {
			if i.WorkID == s.ViewerSessionID {
				return &viewWorkError{Code: "pending_operation", Pending: i.ID}
			}
		}
		b.Previous = append(b.Previous, b.Change)
		b.Change = change
		b.Reason = c.Reason
	} else if b == nil {
		key := ""
		if s.WorkKey != nil {
			key = *s.WorkKey
		}
		e.journal.Bindings = append(e.journal.Bindings, viewWorkBinding{WorkID: s.ViewerSessionID, Key: key, Project: project, Change: change, Source: viewer.TimelineSource{ProjectRoot: project, Host: viewer.TimelineHost(c.Host), NativeSessionID: c.NativeSessionID}})
	}
	return e.save()
}
func (e *viewWorkEngine) open(c viewWorkCommand) (viewer.TimelineSession, error) {
	if c.WorkID != "" {
		s, err := e.session(c.WorkID)
		if err != nil {
			return s, err
		}
		err = e.bind(c, s, false)
		return s, err
	}
	project, err := canonicalWorkDir(c.Project)
	if err != nil {
		return viewer.TimelineSession{}, err
	}
	change, err := canonicalWorkDir(c.Change)
	if err != nil {
		return viewer.TimelineSession{}, err
	}
	c.Project, c.Change = project, change
	source := viewer.TimelineSource{ProjectRoot: project, Host: viewer.TimelineHost(c.Host), NativeSessionID: c.NativeSessionID}
	for _, b := range e.journal.Bindings {
		if b.Source == source && b.Change == change {
			if b.Key != c.Key {
				return viewer.TimelineSession{}, workStateError("binding_mismatch")
			}
			for _, i := range e.journal.Pending {
				if i.Action == "open" && i.WorkID == b.WorkID {
					return e.completeOpen(c, i)
				}
			}
			s, err := e.session(b.WorkID)
			return s, err
		}
	}
	// A pending initial create is identified by the complete semantic binding;
	// retries retain its original request, irrespective of edited title/summary.
	for _, i := range e.journal.Pending {
		if i.Action != "open" || i.Source != source {
			continue
		}
		var original viewWorkOpenContext
		if json.Unmarshal(i.Context, &original) != nil {
			return viewer.TimelineSession{}, workStateError("invalid_journal")
		}
		if original.Binding.Change == change && original.Binding.Key != c.Key {
			return viewer.TimelineSession{}, workStateError("binding_mismatch")
		}
		if original.Binding.Change == change && original.Binding.Key == c.Key {
			return e.completeOpen(c, i)
		}
	}
	index, err := viewIndex(e.info)
	if err != nil {
		return viewer.TimelineSession{}, err
	}
	for _, entry := range index.TimelineSessions {
		if entry.Source == source {
			s, err := e.session(entry.ViewerSessionID)
			if err != nil {
				return s, err
			}
			if s.WorkKey != nil && *s.WorkKey == c.Key {
				err = e.bind(c, s, false)
				return s, err
			}
		}
	}
	// Validate the root relationship before any network mutation.
	rel, err := filepath.Rel(project, change)
	if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
		return viewer.TimelineSession{}, workInput("change must be within project")
	}
	req := viewer.TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: workID("op-"), Source: source, WorkKey: &c.Key, Title: c.Title, Summary: c.Summary, Tickets: []viewer.TimelineTicketInput{}}
	raw, err := json.Marshal(req)
	if err != nil {
		return viewer.TimelineSession{}, err
	}
	if err = validateViewTimelineInput("create_session", raw); err != nil {
		return viewer.TimelineSession{}, err
	}
	i, err := e.prepareIntent("open", "", source, viewWorkOpenContext{Binding: viewWorkBinding{Key: c.Key, Project: project, Change: change, Source: source}, Request: raw})
	if err != nil {
		return viewer.TimelineSession{}, err
	}
	return e.completeOpen(c, i)
}
func (e *viewWorkEngine) completeOpen(c viewWorkCommand, i *viewWorkIntent) (viewer.TimelineSession, error) {
	var context viewWorkOpenContext
	if json.Unmarshal(i.Context, &context) != nil {
		return viewer.TimelineSession{}, workStateError("invalid_journal")
	}
	if len(i.Steps) == 0 {
		if _, err := e.appendStep(i, "POST", "/api/timeline/sessions", "application/json", context.Request); err != nil {
			return viewer.TimelineSession{}, err
		}
	}
	if len(i.Steps) != 1 {
		return viewer.TimelineSession{}, workStateError("incomplete_preparation")
	}
	raw, err := e.executeStep(i, 0)
	if err != nil {
		return viewer.TimelineSession{}, err
	}
	r, err := viewer.DecodeTimelineSessionResponse(raw)
	if err != nil {
		return viewer.TimelineSession{}, err
	}
	current, err := e.session(r.Session.ViewerSessionID)
	if err != nil {
		return viewer.TimelineSession{}, err
	}
	r.Session = current
	b := context.Binding
	if b.Project == "" {
		return r.Session, workStateError("invalid_journal")
	}
	c.Change, c.Project = b.Change, b.Project
	if err = e.bind(c, r.Session, false); err != nil {
		return r.Session, err
	}
	err = e.finishIntent(i)
	return r.Session, err
}
func executeViewWork(c viewWorkCommand, w io.Writer) error {
	if isViewWorkLocal(c.Action) {
		// Management commands keep no publisher journal and take no native identity.
		return executeViewWorkLocal(c, w)
	}
	e, err := openViewWorkEngine(c)
	if err != nil {
		return err
	}
	defer e.Close()
	if (c.Action == "open" || c.Action == "publish") && c.WorkID == "" {
		// The viewer rejects work from a root it does not know, so creating work
		// registers its root. A directory that is not a repository stays the
		// viewer's rejection to give.
		if root, err := canonicalWorkDir(c.Project); err == nil {
			if _, err = viewer.RegisterRepository(e.state, root); err != nil && !errors.Is(err, viewer.ErrNotGitRepository) {
				return err
			}
		}
	}
	switch c.Action {
	case "publish":
		if c.WorkID != "" {
			return e.publish(c, w)
		}
		// The creation form fixes the session's current work; its result names it.
		var out bytes.Buffer
		if err := e.publish(c, &out); err != nil {
			return err
		}
		var r viewWorkResult
		if json.Unmarshal(out.Bytes(), &r) == nil && r.WorkID != "" {
			if err := addCurrentWork(e.state, c.Host, c.NativeSessionID, r.WorkID); err != nil {
				return err
			}
		}
		_, err := w.Write(out.Bytes())
		return err
	case "stage", "tasks register", "task update", "close":
		r, err := e.lifecycle(c)
		if err != nil {
			return err
		}
		if c.Action == "close" {
			if err = releaseCurrentWork(e.state, c.Host, c.NativeSessionID, c.WorkID); err != nil {
				return err
			}
		}
		return writeViewWorkResult(w, r)
	case "open":
		s, err := e.open(c)
		if err != nil {
			return err
		}
		if err = addCurrentWork(e.state, c.Host, c.NativeSessionID, s.ViewerSessionID); err != nil {
			return err
		}
		return writeViewWorkResult(w, viewWorkResult{WorkID: s.ViewerSessionID, Result: "opened", SessionVersion: s.Version})
	case "list":
		return e.list(c, w)
	case "show":
		s, err := e.session(c.WorkID)
		if err != nil {
			return err
		}
		var b *viewWorkBinding
		for n := range e.journal.Bindings {
			if e.journal.Bindings[n].WorkID == c.WorkID {
				b = &e.journal.Bindings[n]
				break
			}
		}
		var pending *viewWorkIntent
		for _, i := range e.journal.Pending {
			if i.WorkID == c.WorkID {
				pending = i
				break
			}
		}
		var p any
		if pending != nil {
			p = map[string]string{"id": pending.ID, "state": pending.State}
		}
		var last any
		for n := len(e.journal.History) - 1; n >= 0; n-- {
			i := e.journal.History[n]
			if i.WorkID == c.WorkID {
				last = map[string]string{"id": i.ID, "action": i.Action, "state": i.State}
				break
			}
		}
		return json.NewEncoder(w).Encode(map[string]any{"work_id": c.WorkID, "title": s.Title, "state": s.State, "session_version": s.Version, "binding": b, "pending": p, "last_operation": last})
	case "bind":
		s, err := e.session(c.WorkID)
		if err != nil {
			return err
		}
		if err = e.bind(c, s, true); err != nil {
			return err
		}
		return writeViewWorkResult(w, viewWorkResult{WorkID: c.WorkID, Result: "bound", SessionVersion: s.Version})
	case "resume", "reconcile":
		return e.recover(c, w)
	default:
		return &viewWorkError{Code: "unsupported_runtime", Next: "this work action is not implemented"}
	}
}
func (e *viewWorkEngine) recover(c viewWorkCommand, w io.Writer) error {
	for _, i := range e.journal.Pending {
		if i.WorkID != c.WorkID {
			continue
		}
		if c.Action == "resume" && (string(i.Source.Host) != c.Host || i.Source.NativeSessionID != c.NativeSessionID) {
			return workStateError("source_mismatch")
		}
		if c.Action == "reconcile" && i.ID != c.Pending {
			continue
		}
		// Resolve the actual caller before replaying any frozen producer write.
		// The stored request's actor does not confer authority on its caller.
		observed, err := e.session(c.WorkID)
		if err != nil {
			return err
		}
		if _, err = e.actor(observed, c); err != nil {
			return err
		}
		if c.Action == "resume" && i.Action == "publish" {
			return e.resumePublish(c, i, w)
		}
		if c.Action == "resume" && i.Action != "open" {
			r, err := e.resumeLifecycle(i)
			if err != nil {
				return err
			}
			return writeViewWorkResult(w, r)
		}
		if c.Action == "resume" && i.Action == "open" {
			s, err := e.completeOpen(c, i)
			if err != nil {
				return err
			}
			return writeViewWorkResult(w, viewWorkResult{WorkID: s.ViewerSessionID, Result: "resumed", SessionVersion: s.Version})
		}
		for n := range i.Steps {
			if len(i.Steps[n].Response) > 0 || i.Steps[n].Rejection != "" {
				continue
			}
			if _, err := e.executeStep(i, n); err != nil && i.State != "rejected_conflict" {
				return err
			}
			if i.State == "rejected_conflict" {
				break
			}
		}
		if c.Action == "reconcile" {
			s, err := e.session(c.WorkID)
			if err != nil {
				return err
			}
			if _, err = e.actor(s, c); err != nil {
				return err
			}
			if err = e.journal.archiveRejected(i.ID, c.Reason); err != nil {
				return err
			}
			if err = e.save(); err != nil {
				return err
			}
			return writeViewWorkResult(w, viewWorkResult{WorkID: c.WorkID, Result: "archived_rejected", SessionVersion: s.Version})
		}
		return &viewWorkError{Code: "pending_continuation_required", Pending: i.ID, Next: "resume using the publisher for the original semantic action"}
	}
	return workStateError("pending_not_found")
}
func (e *viewWorkEngine) list(c viewWorkCommand, w io.Writer) error {
	project, err := canonicalWorkDir(c.Project)
	if err != nil {
		return err
	}
	index, err := viewIndex(e.info)
	if err != nil {
		return err
	}
	entries := index.TimelineSessions
	sort.Slice(entries, func(i, j int) bool { return entries[i].ViewerSessionID < entries[j].ViewerSessionID })
	type item struct {
		WorkID  string               `json:"work_id"`
		Key     *string              `json:"key"`
		Title   string               `json:"title"`
		State   viewer.TimelineState `json:"state"`
		Version int64                `json:"session_version"`
		Binding *viewWorkBinding     `json:"binding"`
	}
	items := []item{}
	for _, entry := range entries {
		if entry.Archived && !c.IncludeArchived {
			continue
		}
		s, err := e.session(entry.ViewerSessionID)
		if err != nil {
			return err
		}
		matches := s.Source.ProjectRoot == project
		var b *viewWorkBinding
		for n := range e.journal.Bindings {
			v := &e.journal.Bindings[n]
			if v.WorkID == s.ViewerSessionID && v.Project == project {
				matches = true
				if c.Host == "" || string(v.Source.Host) == c.Host && v.Source.NativeSessionID == c.NativeSessionID {
					b = v
				}
			}
		}
		if !matches {
			continue
		}
		if c.Host != "" {
			found := false
			for _, p := range s.Producers {
				if string(p.Source.Host) == c.Host && p.Source.NativeSessionID == c.NativeSessionID {
					found = true
				}
			}
			if !found {
				continue
			}
		}
		items = append(items, item{s.ViewerSessionID, s.WorkKey, s.Title, s.State, s.Version, b})
	}
	start := 0
	if c.Cursor != "" {
		raw, err := base64.RawURLEncoding.DecodeString(c.Cursor)
		if err != nil {
			return workInput("cursor")
		}
		start, err = strconv.Atoi(string(raw))
		if err != nil || start < 0 || start > len(items) {
			return workInput("cursor")
		}
	}
	limit := c.Limit
	if limit <= 0 {
		limit = 20
	}
	end := min(start+limit, len(items))
	next := ""
	if end < len(items) {
		next = base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprint(end)))
	}
	return json.NewEncoder(w).Encode(map[string]any{"items": items[start:end], "next_cursor": next})
}

type viewWorkOpenContext struct {
	Binding viewWorkBinding `json:"binding"`
	Request json.RawMessage `json:"request"`
}

func mustWorkJSON(v any) []byte { b, _ := json.Marshal(v); return b }
func workStepOperationID(s viewWorkStep) string {
	raw := s.Body
	if media, params, err := mime.ParseMediaType(s.ContentType); err == nil && media == "multipart/form-data" {
		reader := multipart.NewReader(bytes.NewReader(raw), params["boundary"])
		for {
			part, err := reader.NextPart()
			if err != nil {
				return ""
			}
			if part.FormName() == "metadata" {
				raw, _ = io.ReadAll(io.LimitReader(part, viewer.TimelineMaxBody))
				break
			}
			part.Close()
		}
	}
	var input struct {
		OperationID string `json:"operation_id"`
	}
	json.Unmarshal(raw, &input)
	return input.OperationID
}
