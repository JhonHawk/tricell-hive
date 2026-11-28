package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// workContinuation describes a pending intent for a message that names it: a
// line that identifies it, then the shell command that continues or archives
// it, each on its own line so copying one line runs one command. A command
// carries the intent's own identity and the resolved state directory; when
// that identity is not the caller's, a line before it says so, since only the
// user or that session may run it. Stored free text stays out of commands:
// continuing resends the frozen request, so title, summary and reason only
// have to pass validation.
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
	lines := []string{fmt.Sprintf("%s: %s, %s · %s session %s · %s · %s", workDisplay(i.ID), workDisplay(i.Action), workDisplay(i.State), workDisplay(string(i.Source.Host)), workDisplay(i.Source.NativeSessionID), workDisplay(project), workDisplay(subject))}
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
	var command, after []string
	switch {
	case rejected && i.WorkID != "":
		command = append(append(append(base, "reconcile"), identity...), "--work-id", i.WorkID, "--pending", i.ID, "--action", "archive-rejected")
		after = append(after, "  add --reason TEXT to that command before running it: archiving cannot be undone")
	case rejected:
		lines = append(lines, "  rejected before it created a work: no command continues or archives it")
	case i.Action == "open" && open.Binding.Key != "":
		command = append(append(append(base, "open"), identity...), "--project", open.Binding.Project)
		if open.Binding.Change != "" {
			command = append(command, "--change", open.Binding.Change)
		}
		command = append(command, "--key", open.Binding.Key, "--title", "Resume", "--summary", "Resume")
	case i.Action == "publish" && i.WorkID == "" && publication.Command.Change != "":
		c := publication.Command
		command = append(append(append(base, "publish"), identity...), "--project", c.Project, "--change", c.Change, "--key", c.Key, "--title", "Resume", "--summary", "Resume", "--stage", c.Stage, "--reason", "Resume")
		if c.DocumentKey != "" {
			command = append(command, "--document-key", c.DocumentKey)
		}
		if c.Status != "" {
			command = append(command, "--status", c.Status)
		}
		for _, d := range c.Documents {
			command = append(command, "--document", d)
		}
		if c.ExplicitDocuments {
			command = append(command, "--explicit-documents")
		}
	case i.WorkID != "":
		command = append(append(append(base, "resume"), identity...), "--work-id", i.WorkID)
	default:
		lines = append(lines, "  no command continues it")
	}
	for _, arg := range command {
		if strings.ContainsFunc(arg, workLineBreaking) {
			// A quoted control or separator character would split the command
			// over lines, or reorder it on screen.
			lines, command, after = append(lines, "  a path or key holds a control character: no command is printed"), nil, nil
			break
		}
	}
	if command != nil {
		if string(i.Source.Host) != caller.Host || i.Source.NativeSessionID != caller.NativeSessionID {
			lines = append(lines, "  another session's: give it to the user, do not run it from this session")
		}
		quoted := make([]string, len(command))
		for n, arg := range command {
			quoted[n] = shellQuote(arg)
		}
		lines = append(append(lines, "  "+strings.Join(quoted, " ")), after...)
	}
	if i.State == "needs_reconciliation" {
		lines = append(lines, "  its outcome is unknown, so it cannot be discarded")
	}
	return lines
}

// workLineBreaking is a character that can start a new line or reorder text
// when shown: a control character, a format character such as a bidi
// override, or a line or paragraph separator.
func workLineBreaking(r rune) bool {
	return unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zl, unicode.Zp)
}

// workDisplay shows a stored value on the description line, quoted when it
// holds such a character, so it cannot start a line of its own.
func workDisplay(s string) string {
	if strings.ContainsFunc(s, workLineBreaking) {
		return strconv.QuoteToASCII(s)
	}
	return s
}

// workIntentNext is the next action for one pending intent: the lead, then the
// intent's description and its continuation.
func workIntentNext(i *viewWorkIntent, caller viewWorkCommand, lead string) string {
	return lead + ":\n" + strings.Join(workContinuation(i, caller), "\n")
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
