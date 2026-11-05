package management

import (
	"bytes"
	"fmt"

	"tricell-hive/integrations/hostevent"
	"tricell-hive/integrations/target"
)

// inspectJSONHooks reports which events of f hold one of Hive's entries in the
// file as it is. skipped is non-empty when the file cannot be read as the
// expected shape, and then present is nil. A missing or blank file has none.
func inspectJSONHooks(cur []byte, f hostevent.JSONHookFile, rec *HostHooksRecord) (present map[string]bool, skipped string) {
	present = map[string]bool{}
	if len(bytes.TrimSpace(cur)) == 0 {
		return present, ""
	}
	root, err := parseJSONSpans(cur)
	if err != nil {
		return nil, err.Error()
	}
	if root.kind != '{' {
		return nil, "the top level is not a JSON object"
	}
	hooksM, err := root.member("hooks")
	if err != nil {
		return nil, err.Error() + " in the top level object"
	}
	if hooksM == nil {
		return present, ""
	}
	if hooksM.val.kind != '{' {
		return nil, "\"hooks\" is not an object"
	}
	var recorded map[string]bool
	if rec != nil {
		recorded = recordedSet(rec.Commands)
	}
	for _, event := range f.Events {
		em, err := hooksM.val.member(event)
		if err != nil {
			return nil, err.Error() + " in \"hooks\""
		}
		if em == nil {
			continue
		}
		if em.val.kind != '[' {
			return nil, fmt.Sprintf("\"hooks.%s\" is not an array", event)
		}
		present[event] = len(ownEntries(cur, em.val, f, event, recorded)) > 0
	}
	return present, ""
}

// hostInstalled reports whether Hive has installed anything for host in this
// home, so a fresh home reports no hook rows.
func hostInstalled(state State, c target.Config, host string) bool {
	for _, r := range state.Records {
		if hasConsumer(r.Consumers, Consumer{Host: host, Scope: c.Scope, Context: c.Home}) {
			return true
		}
	}
	return false
}

// hostHookRows are the Status rows of the Hive entries in the shared JSON hook
// files of the selected hosts: one row per event, installed, missing, or skipped
// with the reason the file cannot be read. The rows say what the files hold
// now, not what the state recorded.
func hostHookRows(c target.Config, hosts []string, state State, pending bool) []StatusEntry {
	if c.Scope != "user" {
		return nil
	}
	var out []StatusEntry
	for _, h := range hosts {
		spec, ok := hostHookSpec(h, c)
		if !ok || !hostInstalled(state, c, h) {
			continue
		}
		var present map[string]bool
		var reason string
		cur, err := readHookFile(spec.Path)
		if err != nil {
			reason = skipReason(err, spec.Path)
		} else {
			present, reason = inspectJSONHooks(cur.Data, spec, priorHostHooks(state, spec.Path))
		}
		for _, event := range spec.Events {
			en := StatusEntry{Path: spec.Path, Host: h, Kind: "hook_entry", Event: event, Status: "missing"}
			switch {
			case reason != "":
				en.Status, en.Reason = "skipped", reason
			case present[event]:
				en.Status = "installed"
			}
			if pending {
				en.Status = "recovery_required"
			}
			out = append(out, en)
		}
	}
	return out
}
