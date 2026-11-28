package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"tricell-hive/tooling/viewer"
)

// workContinuation describes a pending intent for a message that names it: a
// line that identifies it, then the shell command that continues or archives
// it, each on its own line so copying one line runs one command. A command
// carries the intent's own identity; when that identity is not the caller's,
// a line before it says so, since only the user or that session may run it.
func workContinuation(i *viewWorkIntent, caller viewWorkCommand) []string {
	project, key := i.Source.ProjectRoot, ""
	var open viewWorkOpenContext
	var publication workPublicationIntent
	switch i.Action {
	case "open":
		if json.Unmarshal(i.Context, &open) == nil {
			project, key = open.Binding.Project, open.Binding.Key
		}
	case "publish":
		if json.Unmarshal(i.Context, &publication) == nil {
			project, key = publication.Command.Project, publication.Command.Key
		}
	}
	subject := "work " + i.WorkID
	if i.WorkID == "" {
		subject = "key " + key
	}
	lines := []string{fmt.Sprintf("%s: %s, %s · %s session %s · %s · %s", i.ID, i.Action, i.State, i.Source.Host, i.Source.NativeSessionID, project, subject)}
	base := []string{"hive", "view", "work"}
	identity := []string{"--host", string(i.Source.Host), "--native-session-id", i.Source.NativeSessionID}
	if caller.Options.StateDir != "" {
		identity = append(identity, "--state-dir", caller.Options.StateDir)
	}
	if caller.Options.Home != "" {
		identity = append(identity, "--home", caller.Options.Home)
	}
	rejected := false
	for _, s := range i.Steps {
		rejected = rejected || s.Rejection != ""
	}
	var command []string
	switch {
	case rejected && i.WorkID != "":
		command = append(append(append(base, "reconcile"), identity...), "--work-id", i.WorkID, "--pending", i.ID, "--action", "archive-rejected", "--reason", "<why it is archived>")
	case rejected:
		lines = append(lines, "  rejected before it created a work: no command continues or archives it")
	case i.Action == "open" && open.Binding.Key != "":
		var request viewer.TimelineCreateSessionRequest
		_ = json.Unmarshal(open.Request, &request)
		command = append(append(append(base, "open"), identity...), "--project", open.Binding.Project)
		if open.Binding.Change != "" {
			command = append(command, "--change", open.Binding.Change)
		}
		// Continuing resends the frozen request: title and summary only have
		// to pass validation, so they are kept on one line.
		command = append(command, "--key", open.Binding.Key, "--title", workOneLine(request.Title, "Resume"), "--summary", workOneLine(request.Summary, "Resume"))
	case i.Action == "publish" && i.WorkID == "" && publication.Command.Change != "":
		c := publication.Command
		command = append(append(append(base, "publish"), identity...), "--project", c.Project, "--change", c.Change, "--key", c.Key, "--title", workOneLine(c.Title, "Resume"), "--summary", workOneLine(c.Summary, "Resume"), "--stage", c.Stage, "--reason", workOneLine(c.Reason, "Resume"))
		if c.DocumentKey != "" {
			command = append(command, "--document-key", c.DocumentKey)
		}
		if c.Status != "" {
			command = append(command, "--status", c.Status)
		}
	case i.WorkID != "":
		command = append(append(append(base, "resume"), identity...), "--work-id", i.WorkID)
	default:
		lines = append(lines, "  no command continues it")
	}
	if command != nil {
		if string(i.Source.Host) != caller.Host || i.Source.NativeSessionID != caller.NativeSessionID {
			lines = append(lines, "  another session's: give it to the user, do not run it from this session")
		}
		quoted := make([]string, len(command))
		for n, arg := range command {
			quoted[n] = shellQuote(arg)
		}
		lines = append(lines, "  "+strings.Join(quoted, " "))
	}
	if i.State == "needs_reconciliation" {
		lines = append(lines, "  its outcome is unknown, so it cannot be discarded")
	}
	return lines
}

// workOneLine joins text that only has to pass validation into one line.
func workOneLine(s, fallback string) string {
	if one := strings.Join(strings.Fields(s), " "); one != "" {
		return one
	}
	return fallback
}

// workCapacityNext names each pending intent that holds the journal and how it
// continues.
func workCapacityNext(pending []*viewWorkIntent, caller viewWorkCommand) string {
	lines := []string{"pending intents hold the publisher journal; continue or archive one before starting another:"}
	for _, i := range pending {
		lines = append(lines, workContinuation(i, caller)...)
	}
	return strings.Join(lines, "\n")
}
