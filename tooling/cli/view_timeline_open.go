package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"tricell-hive/tooling/viewer"
)

func viewOpenTimelineTarget(paths viewPaths, w io.Writer, deps viewDeps, sessionID, eventID string) error {
	if !viewTimelineSessionID.MatchString(sessionID) || (eventID != "" && !viewTimelineObjectID.MatchString(eventID)) {
		return errors.New("open requires a valid --session-id and optional --event-id")
	}
	info, err := viewTimelineRunning(paths)
	if err != nil {
		return err
	}
	raw, err := viewTimelineHTTP(info, http.MethodGet, "/api/timeline/sessions/"+sessionID, "", nil, viewer.TimelineMaxBody)
	if err != nil {
		return err
	}
	response, err := viewer.DecodeTimelineSessionResponse(raw)
	if err != nil || response.Session.ViewerSessionID != sessionID {
		return errors.New("cannot confirm the target viewer session")
	}
	if eventID != "" {
		found := false
		for _, event := range response.Session.Events {
			if event.EventID == eventID {
				found = true
				break
			}
		}
		if !found {
			return errors.New("the requested event does not belong to the target viewer session")
		}
	}
	code, err := viewAccessCode(info)
	if err != nil {
		return err
	}
	query := url.Values{"session_id": {sessionID}}
	if eventID != "" {
		query.Set("event_id", eventID)
	}
	target := info.url() + "/?" + query.Encode()
	if err := deps.open(target + "#code=" + code); err != nil {
		return errors.New("cannot open the target; retry `hive view open` with the same target")
	}
	fmt.Fprintf(w, "Opened %s\n", target)
	return nil
}
