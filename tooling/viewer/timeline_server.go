package viewer

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func timelineHTTPError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "storage_unavailable"
	field := ""
	var failure *TimelineError
	var contract *TimelineContractError
	switch {
	case errors.As(err, &failure):
		status = failure.Status
		code = failure.Code
	case errors.As(err, &contract):
		code = contract.Code
		field = contract.Field
		status = 400
		switch code {
		case "version_conflict", "stage_active", "session_closed", "stage_terminal", "stage_resume_required", "already_in_state", "entity_exists", "entity_conflict", "duplicate_plan_order":
			status = 409
		case "capacity", "session_limit", "limit":
			status = 413
		case "producer_not_accepted", "lead_required", "content_owner_required", "segment_assignee_required", "forbidden_actor", "local_viewer_required", "native_producer_required":
			status = 403
		case "unknown_event", "unknown_task", "unknown_wait", "unknown_ticket", "unknown_delivery", "unknown_candidate":
			status = 404
		case "current_predecessor_required", "stage_unavailable", "stage_binding_mismatch", "reentry_binding_required", "duplicate_identity", "duplicate_key", "duplicate_stage_order", "pending_request_required", "session_terminal", "immutable_audit", "invalid_transition", "foreign_project", "unknown_or_mixed_attempt", "candidate_mismatch", "attempt_already_known":
			status = 409
		}
	case errors.Is(err, ErrNotFound):
		status = 404
		code = "not_found"
	case errors.Is(err, ErrInvalidKey), errors.Is(err, ErrInvalidPath):
		status = 400
		code = "invalid_id"
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	body := map[string]any{"code": code}
	if failure != nil && failure.CleanupPending {
		body["cleanup_pending"] = true
		if failure.RemainingFiles != nil {
			body["remaining_files"] = *failure.RemainingFiles
		}
	}
	if field != "" {
		body["field"] = field
	}
	_ = json.NewEncoder(w).Encode(body)
}
func timelineReadBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, TimelineMaxBody))
	if err != nil {
		return nil, timelineFailure("body_too_large", 413)
	}
	return raw, nil
}
func (s *Server) timelineResponse(v TimelineSession, created *bool) TimelineSessionResponse {
	out := TimelineSessionResponse{SchemaVersion: 3, Mode: "timeline_api", Session: v, Created: created, ArtifactURLs: map[string]string{}, DocumentURLs: map[string]string{}}
	for _, a := range v.Artifacts {
		out.ArtifactURLs[a.ArtifactID] = "/files/" + v.ViewerSessionID + "/" + s.fileCap(v.ViewerSessionID) + "/artifacts/" + a.ArtifactID + timelineImageExtension(a.MediaType)
	}
	for _, d := range v.Documents {
		out.DocumentURLs[d.DocumentID] = "/api/timeline/sessions/" + v.ViewerSessionID + "/documents/" + d.DocumentID
	}
	return out
}
func (s *Server) timeline(w http.ResponseWriter, r *http.Request) {
	if !timelineRevisionMatches(w, r) {
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/timeline/sessions"), "/")
	if len(parts) == 1 && parts[0] == "" {
		if !allow(w, r, http.MethodPost) || !s.authorizeSessionWrite(w, r) {
			return
		}
		raw, err := timelineReadBody(w, r)
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		request, err := DecodeTimelineCreateSession(raw)
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		response, err := s.store.TimelineCreate(request)
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		writeJSON(w, s.timelineResponse(response.Session, response.Created))
		return
	}
	if len(parts) < 2 || parts[0] != "" || !keyPattern.MatchString(parts[1]) {
		timelineHTTPError(w, ErrNotFound)
		return
	}
	id := parts[1]
	if s.timelineDomain(w, r, id, parts[2:]) {
		return
	}
	if len(parts) == 2 {
		if r.Method == http.MethodGet {
			if !s.authorize(w, r, false) {
				return
			}
			v, err := s.store.TimelineGet(id)
			if err != nil {
				timelineHTTPError(w, err)
				return
			}
			writeJSON(w, s.timelineResponse(v, nil))
			return
		}
		if !allow(w, r, http.MethodPatch, http.MethodDelete) || !s.authorizeSessionWrite(w, r) {
			return
		}
		raw, err := timelineReadBody(w, r)
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		var result TimelineOperationResult
		if r.Method == http.MethodPatch {
			var request TimelineSessionPatchRequest
			request, err = DecodeTimelineSessionPatch(raw)
			if err == nil {
				result, err = s.store.TimelinePatch(id, request)
			}
		} else {
			var request TimelineMutationRequest
			request, err = DecodeTimelineMutation(raw)
			if err == nil {
				result, err = s.store.TimelineDelete(id, request)
			}
		}
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		writeJSON(w, result)
		return
	}
	if len(parts) == 4 && parts[2] == "documents" {
		if !allow(w, r, http.MethodGet) || !s.authorize(w, r, false) {
			return
		}
		response, err := s.store.TimelineDocument(id, parts[3])
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		writeJSON(w, response)
		return
	}
	if len(parts) == 3 && (parts[2] == "artifacts" || parts[2] == "documents") {
		if !allow(w, r, http.MethodPost) || !s.authorizeSessionWrite(w, r) {
			return
		}
		raw, content, err := timelineMultipart(w, r, parts[2] == "documents")
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		var result TimelineOperationResult
		if parts[2] == "documents" {
			var metadata TimelineDocumentMetadata
			metadata, err = DecodeTimelineDocumentMetadata(raw)
			if err == nil {
				result, err = s.store.TimelineAddDocument(id, metadata, content)
			}
		} else {
			var metadata TimelineArtifactMetadata
			metadata, err = DecodeTimelineArtifactMetadata(raw)
			if err == nil {
				result, err = s.store.TimelineAddArtifact(id, metadata, content)
			}
		}
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		writeJSON(w, result)
		return
	}
	if len(parts) == 3 && parts[2] == "events" {
		if !allow(w, r, http.MethodPost) || !s.authorizeSessionWrite(w, r) {
			return
		}
		raw, err := timelineReadBody(w, r)
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		request, err := DecodeTimelineAppendEvent(raw)
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		result, err := s.store.TimelineAppend(id, request)
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		writeJSON(w, result)
		return
	}
	if len(parts) == 4 && parts[2] == "events" {
		if !allow(w, r, http.MethodPatch, http.MethodDelete) || !s.authorizeSessionWrite(w, r) {
			return
		}
		raw, err := timelineReadBody(w, r)
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		var result TimelineOperationResult
		if r.Method == http.MethodPatch {
			var request TimelineEventPatchRequest
			request, err = DecodeTimelineEventPatch(raw)
			if err == nil {
				result, err = s.store.TimelinePatchEvent(id, parts[3], request)
			}
		} else {
			var request TimelineMutationRequest
			request, err = DecodeTimelineMutation(raw)
			if err == nil {
				result, err = s.store.TimelineWithdraw(id, parts[3], request)
			}
		}
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		writeJSON(w, result)
		return
	}
	if len(parts) == 5 && parts[2] == "events" && parts[4] == "document-snapshots" {
		if !allow(w, r, http.MethodPost) || !s.authorizeSessionWrite(w, r) {
			return
		}
		raw, err := timelineReadBody(w, r)
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		request, err := DecodeTimelineDocumentSnapshot(raw)
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		result, err := s.store.TimelineSnapshot(id, parts[3], request)
		if err != nil {
			timelineHTTPError(w, err)
			return
		}
		writeJSON(w, result)
		return
	}
	timelineHTTPError(w, ErrNotFound)
}
func timelineMultipart(w http.ResponseWriter, r *http.Request, document bool) ([]byte, []byte, error) {
	max := int64(maxCapture)
	if document {
		max = TimelineMaxDocumentBytes
	}
	r.Body = http.MaxBytesReader(w, r.Body, max+TimelineMaxBody)
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, nil, timelineFailure("invalid_multipart", 400)
	}
	var metadata, content []byte
	seen := map[string]bool{}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, timelineFailure("invalid_multipart", 400)
		}
		name := part.FormName()
		if seen[name] || (name != "metadata" && name != "file") {
			part.Close()
			return nil, nil, timelineFailure("invalid_multipart", 400)
		}
		seen[name] = true
		limit := max
		if name == "metadata" {
			limit = TimelineMaxBody
		}
		raw, readErr := io.ReadAll(io.LimitReader(part, limit+1))
		part.Close()
		if readErr != nil || int64(len(raw)) > limit {
			return nil, nil, timelineFailure("body_too_large", 413)
		}
		if name == "metadata" {
			metadata = raw
		} else {
			content = raw
		}
	}
	if !seen["metadata"] || !seen["file"] || len(bytes.TrimSpace(metadata)) == 0 {
		return nil, nil, timelineFailure("invalid_multipart", 400)
	}
	return metadata, content, nil
}

// Revision is checked before decoding or mutating; capability files bypass this route.
func timelineRevisionMatches(w http.ResponseWriter, r *http.Request) bool {
	if values := r.Header.Values(TimelineRevisionHeader); len(values) == 1 && values[0] == "2" {
		return true
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusConflict)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": "reload_required", "message": "Update Hive and reload the viewer to continue; preserve inputs and retry the unchanged operation ID."})
	return false
}
