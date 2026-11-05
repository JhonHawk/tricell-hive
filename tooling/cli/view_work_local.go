package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

// The local management commands act as the local_viewer actor with the declared
// origin cli. They carry no native producer identity and keep no publisher
// journal: after an uncertain response they re-read the work and say whether the
// action holds. The version the request expects makes a repeated command fail
// instead of applying twice, and a deleted work answers 410.

// localViewerCapability is announced by /api/health when the viewer accepts the
// origin cli. A viewer without it may reject the field or ignore it.
const localViewerCapability = "local_actions_cli"

// viewWorkLocalResults names the success of each management command.
var viewWorkLocalResults = map[string]string{
	"rename": "renamed", "archive": "archived", "restore": "restored",
	"pin": "pinned", "unpin": "unpinned", "recover": "recovered", "delete": "deleted",
}

func isViewWorkLocal(action string) bool { _, ok := viewWorkLocalResults[action]; return ok }

var (
	errWorkUncertain  = errors.New("uncertain response")
	errWorkIncomplete = errors.New("incomplete response")
)

func localViewerActor() viewer.TimelineActor {
	return viewer.TimelineActor{Kind: "local_viewer", Origin: "cli"}
}

// validateViewWorkLocal checks the semantic input of a management command
// without touching the network.
func validateViewWorkLocal(c *viewWorkCommand) error {
	if c.WorkID == "" {
		return workInput("required semantic field missing")
	}
	switch c.Action {
	case "rename":
		// The viewer stores the trimmed title; send and compare that one.
		c.Title = strings.TrimSpace(c.Title)
		title := c.Title
		r := viewer.TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "validate", Actor: localViewerActor(), ExpectedVersion: 1, Title: &title}
		if strings.TrimSpace(title) == "" || r.Validate() != nil {
			return workInput("invalid title")
		}
	case "recover":
		if strings.TrimSpace(c.RequestID) == "" || strings.TrimSpace(c.Reason) == "" {
			return workInput("required semantic field missing")
		}
		r := viewer.TimelineRecoverLeadRequest{TimelineNativeOperation: viewer.TimelineNativeOperation{SchemaVersion: 3, OperationID: "validate", Actor: localViewerActor(), ExpectedSessionVersion: 1, Reason: c.Reason}, RequestID: c.RequestID}
		if viewer.ValidateTimelineV3Operation(&r) != nil {
			return workInput("invalid request-id or reason")
		}
	case "delete":
		if c.Confirm != c.WorkID {
			return workInput("delete requires --confirm equal to --work-id")
		}
	}
	return nil
}

type viewWorkLocal struct {
	info serverInfo
}

func executeViewWorkLocal(c viewWorkCommand, w io.Writer) error {
	_, state, err := management.NormalizeOptions(c.Options)
	if err != nil {
		return err
	}
	state, err = filepath.Abs(state)
	if err != nil {
		return err
	}
	info, health, err := viewTimelineRunningHealth(newViewPaths(state))
	if err != nil {
		return err
	}
	if err = requireLocalCapability(health); err != nil {
		return err
	}
	l := viewWorkLocal{info: info}
	result, err := l.run(c)
	if err != nil {
		return err
	}
	return writeViewWorkResult(w, result)
}

func requireLocalCapability(health viewTimelineHealth) error {
	for _, name := range health.Capabilities {
		if name == localViewerCapability {
			return nil
		}
	}
	return &viewWorkError{Code: "viewer_update_required", Next: "the running viewer does not announce " + localViewerCapability + "; nothing was sent; the user updates Hive and restarts the viewer (an agent does not restart it), then retry"}
}

// read returns the work, or the stable code the viewer gave when it cannot.
func (l viewWorkLocal) read(id string) (viewer.TimelineSession, string, error) {
	raw, code, err := viewWorkHTTP(l.info, http.MethodGet, "/api/timeline/sessions/"+id, "", nil)
	if err != nil {
		return viewer.TimelineSession{}, code, err
	}
	r, err := viewer.DecodeTimelineSessionResponse(raw)
	if err != nil || r.Session.ViewerSessionID != id {
		return viewer.TimelineSession{}, "", errors.New("invalid response")
	}
	return r.Session, "", nil
}

func (l viewWorkLocal) run(c viewWorkCommand) (viewWorkResult, error) {
	current, code, err := l.read(c.WorkID)
	if err != nil {
		if code == "session_deleted" {
			return viewWorkResult{}, &viewWorkError{Code: code, Next: "the work no longer exists; nothing was sent"}
		}
		return viewWorkResult{}, workStateError("work_unavailable")
	}
	operationID := workID("cli-")
	method, path := http.MethodPatch, "/api/timeline/sessions/"+c.WorkID
	var body any
	holds := func(s viewer.TimelineSession) bool { return false }
	switch c.Action {
	case "rename", "archive", "restore", "pin", "unpin":
		patch := viewer.TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: operationID, Actor: localViewerActor(), ExpectedVersion: current.Version}
		yes, no := true, false
		switch c.Action {
		case "rename":
			patch.Title = &c.Title
			holds = func(s viewer.TimelineSession) bool { return s.Title == c.Title }
		case "archive":
			patch.Archived = &yes
			holds = func(s viewer.TimelineSession) bool { return s.Archived }
		case "restore":
			patch.Archived = &no
			holds = func(s viewer.TimelineSession) bool { return !s.Archived }
		case "pin":
			patch.Pinned = &yes
			holds = func(s viewer.TimelineSession) bool { return s.Pinned }
		case "unpin":
			patch.Pinned = &no
			holds = func(s viewer.TimelineSession) bool { return !s.Pinned }
		}
		if patch.Validate() != nil {
			return viewWorkResult{}, workInput("invalid management request")
		}
		body = patch
	case "recover":
		var requested *viewer.TimelineJoinRequest
		for n := range current.PendingJoinRequests {
			if current.PendingJoinRequests[n].RequestID == c.RequestID {
				requested = &current.PendingJoinRequests[n]
			}
		}
		if requested == nil {
			return viewWorkResult{}, &viewWorkError{Code: "request_not_pending", Next: "nothing was sent; the new session asks to join with `hive view producer request`, then retry with its pending request ID"}
		}
		method, path = http.MethodPost, path+"/recover-lead"
		body = viewer.TimelineRecoverLeadRequest{TimelineNativeOperation: viewer.TimelineNativeOperation{SchemaVersion: 3, OperationID: operationID, Actor: localViewerActor(), ExpectedSessionVersion: current.Version, Reason: c.Reason}, RequestID: c.RequestID}
		source := requested.Source
		holds = func(s viewer.TimelineSession) bool {
			for _, p := range s.PendingJoinRequests {
				if p.RequestID == c.RequestID {
					return false
				}
			}
			for _, p := range s.Producers {
				if p.ProducerID == s.ForegroundProducerID && p.Source.Host == source.Host && p.Source.NativeSessionID == source.NativeSessionID {
					return true
				}
			}
			return false
		}
	case "delete":
		method = http.MethodDelete
		body = viewer.TimelineMutationRequest{SchemaVersion: 3, OperationID: operationID, Actor: localViewerActor(), ExpectedVersion: current.Version}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return viewWorkResult{}, workInput("invalid management request")
	}
	finished := func(version int64) viewWorkResult {
		return viewWorkResult{WorkID: c.WorkID, Result: viewWorkLocalResults[c.Action], SessionVersion: version}
	}
	wantKind := viewer.TimelineOperationKind("session_patch")
	switch c.Action {
	case "recover":
		wantKind = "recover_lead"
	case "delete":
		wantKind = "session_delete"
	}
	answered := func(raw []byte) (int64, bool) {
		r, err := viewer.DecodeTimelineOperationResult(raw)
		if err == nil && r.OperationID == operationID && r.ViewerSessionID == c.WorkID && r.Kind == wantKind && r.SessionVersion > current.Version {
			return r.SessionVersion, true
		}
		return 0, false
	}
	rejected := func(code string) error {
		return &viewWorkError{Code: code, Next: "nothing was changed; read the current state with `hive view work show --work-id " + c.WorkID + "` and retry the command if it still applies"}
	}
	pending := &viewWorkError{Code: "delete_pending", Next: "the deletion committed and its file cleanup is pending; this command cannot finish it"}
	// The first attempt and, when its outcome is uncertain, one replay of the
	// exact same bytes and operation ID: the viewer answers a committed request
	// with its stored receipt and applies a lost one once.
	sawPending := false
	for attempt := 0; attempt < 2; attempt++ {
		response, status, code, err := l.send(method, path, raw)
		if err == nil {
			if version, ok := answered(response); ok {
				return finished(version), nil
			}
			continue
		}
		switch {
		case c.Action == "delete" && code == "delete_pending":
			sawPending = true
		case status >= 400 && status < 500 && code != "" && code != "session_deleted":
			return viewWorkResult{}, rejected(code)
		}
	}
	if sawPending {
		return viewWorkResult{}, pending
	}
	return l.settle(c, current.Version, holds, finished)
}

// send writes without a preflight: the health read at startup already confirmed
// the viewer. It returns the status and the machine code of a rejection.
func (l viewWorkLocal) send(method, path string, body []byte) ([]byte, int, string, error) {
	req, err := http.NewRequest(method, l.info.url()+path, bytes.NewReader(body))
	if err != nil {
		return nil, 0, "", errors.New("invalid request")
	}
	req.Header.Set("X-Hive-Token", l.info.Token)
	req.Header.Set(viewer.TimelineRevisionHeader, viewer.TimelineRevisionValue)
	req.Header.Set("Origin", l.info.url())
	req.Header.Set("X-Hive-Action-Scope", "session")
	req.Header.Set("Content-Type", "application/json")
	client := *viewClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, "", errWorkUncertain
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, viewer.TimelineMaxResponseBody+1))
	if err != nil || len(raw) > viewer.TimelineMaxResponseBody {
		return nil, resp.StatusCode, "", errWorkIncomplete
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var failure struct {
			Code string `json:"code"`
		}
		json.Unmarshal(raw, &failure)
		if !viewTimelineObjectID.MatchString(failure.Code) {
			failure.Code = ""
		}
		return nil, resp.StatusCode, failure.Code, errors.New("request rejected")
	}
	if resp.Header.Get(viewer.TimelineRevisionHeader) != viewer.TimelineRevisionValue {
		return nil, resp.StatusCode, "", errWorkIncomplete
	}
	return raw, resp.StatusCode, "", nil
}

// settle answers an uncertain outcome by re-reading the work.
func (l viewWorkLocal) settle(c viewWorkCommand, expected int64, holds func(viewer.TimelineSession) bool, finished func(int64) viewWorkResult) (viewWorkResult, error) {
	unknown := &viewWorkError{Code: "outcome_unknown", Next: "the response was lost and the work could not be re-read; run `hive view work show --work-id " + c.WorkID + "` before retrying"}
	s, code, err := l.read(c.WorkID)
	if c.Action == "delete" {
		switch {
		case err == nil:
			return viewWorkResult{}, &viewWorkError{Code: "not_applied", Next: "the work still exists at version " + strconv.FormatInt(s.Version, 10) + "; retry the command"}
		case code == "session_deleted":
			return viewWorkResult{}, &viewWorkError{Code: "work_gone", Next: "the work no longer exists; whether this request or another removed it cannot be told"}
		}
		return viewWorkResult{}, unknown
	}
	if err != nil {
		return viewWorkResult{}, unknown
	}
	if s.Version > expected && holds(s) {
		return finished(s.Version), nil
	}
	if s.Version == expected {
		return viewWorkResult{}, &viewWorkError{Code: "not_applied", Next: "the work is unchanged at version " + strconv.FormatInt(s.Version, 10) + "; retry the command"}
	}
	return viewWorkResult{}, &viewWorkError{Code: "not_applied", Next: "the work changed meanwhile (version " + strconv.FormatInt(s.Version, 10) + ") and does not hold this action; read it with `hive view work show --work-id " + c.WorkID + "` and retry if it still applies"}
}
