package viewer

import "net/http"

func (s *Server) timelineDomain(w http.ResponseWriter, r *http.Request, id string, parts []string) bool {
	kind, object, method := "", "", http.MethodPost
	if len(parts) == 1 {
		switch parts[0] {
		case "join-requests":
			kind = "join_request"
		case "recover-lead":
			kind = "recover_lead"
		case "tickets":
			kind = "ticket_create"
		case "deliveries":
			kind = "delivery_create"
		case "candidates":
			kind = "candidate_create"
		case "tasks":
			kind = "task_create"
		case "waits":
			kind = "wait_create"
		}
	}
	if len(parts) == 2 {
		switch parts[0] + "/" + parts[1] {
		case "producers/accept":
			kind = "producer_accept"
		case "producers/transfer":
			kind = "foreground_transfer"
		}
		if parts[0] == "tickets" || parts[0] == "deliveries" {
			kind = "ticket_patch"
			if parts[0] == "deliveries" {
				kind = "delivery_patch"
			}
			object = parts[1]
			method = http.MethodPatch
		}
		if parts[0] == "tasks" {
			kind = "task_patch"
			object = parts[1]
			method = http.MethodPatch
		}
	}
	if len(parts) == 3 {
		object = parts[1]
		switch parts[0] + "/" + parts[2] {
		case "tasks/transitions":
			kind = "task_transition"
		case "waits/attempts":
			kind = "wait_attempt_start"
		case "waits/observations":
			kind = "wait_observe"
		}
	}
	if kind == "" {
		return false
	}
	if !allow(w, r, method) || !s.authorizeSessionWrite(w, r) {
		return true
	}
	if object != "" {
		if err := timelineID(object, "object_id"); err != nil {
			timelineHTTPError(w, err)
			return true
		}
	}
	raw, err := timelineReadBody(w, r)
	if err != nil {
		timelineHTTPError(w, err)
		return true
	}
	request, err := DecodeTimelineV3Operation(kind, raw)
	if err != nil {
		timelineHTTPError(w, err)
		return true
	}
	result, err := s.store.TimelineV3Operation(id, kind, object, request)
	if err != nil {
		timelineHTTPError(w, err)
		return true
	}
	writeJSON(w, result)
	return true
}
