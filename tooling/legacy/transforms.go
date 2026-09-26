package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func cleanHooks(path, host string, known []hookRecord) (Edit, bool, error) {
	b, mode, err := read(path)
	if err != nil || b == nil {
		return Edit{}, false, err
	}
	if err = uniqueJSON(b); err != nil {
		return Edit{}, false, fmt.Errorf("invalid or duplicate hook configuration keys: %s", path)
	}
	var doc map[string]json.RawMessage
	if err = json.Unmarshal(b, &doc); err != nil {
		return Edit{}, false, fmt.Errorf("invalid hook configuration: %s", path)
	}
	raw, ok := doc["hooks"]
	if !ok {
		return Edit{}, false, nil
	}
	var events map[string][]json.RawMessage
	if err = json.Unmarshal(raw, &events); err != nil {
		return Edit{}, false, fmt.Errorf("invalid hooks shape: %s", path)
	}
	changed := false
	for event, entries := range events {
		var next []json.RawMessage
		for _, entry := range entries {
			var obj map[string]json.RawMessage
			if err = json.Unmarshal(entry, &obj); err != nil {
				return Edit{}, false, fmt.Errorf("invalid hook event: %s", path)
			}
			var hooks []json.RawMessage
			if err = json.Unmarshal(obj["hooks"], &hooks); err != nil {
				return Edit{}, false, fmt.Errorf("invalid hook list: %s", path)
			}
			var keep []json.RawMessage
			local := false
			for _, h := range hooks {
				var command struct {
					Command string
					Type    string
				}
				if err = json.Unmarshal(h, &command); err != nil {
					return Edit{}, false, err
				}
				match := false
				for _, k := range known {
					if k.Host == host && k.Command == command.Command && k.Event == event && command.Type == "command" {
						match = true
						break
					}
				}
				if !match {
					for _, k := range known {
						name := filepath.Base(strings.Trim(strings.Split(k.Command, " ")[0], "\""))
						if len(name) > 3 && strings.Contains(command.Command, name) {
							return Edit{}, false, fmt.Errorf("unrecognized Hive hook registration: %s", path)
						}
					}
				}
				if match {
					local = true
					changed = true
				} else {
					keep = append(keep, h)
				}
			}
			if !local {
				next = append(next, entry)
			} else if len(keep) > 0 {
				obj["hooks"], _ = json.Marshal(keep)
				v, _ := json.Marshal(obj)
				next = append(next, v)
			}
		}
		if len(next) == 0 {
			delete(events, event)
		} else {
			events[event] = next
		}
	}
	if !changed {
		return Edit{}, false, nil
	}
	doc["hooks"], _ = json.Marshal(events)
	after, err := replaceTopValue(b, "hooks", doc["hooks"])
	return Edit{Path: path, Before: b, After: after, Mode: uint32(mode)}, true, err
}

type patchCatalog struct {
	Files map[string]struct{ SHA256Before, SHA256After string }
}

func patches() patchCatalog {
	var p patchCatalog
	if e := json.Unmarshal(patchJSON, &p); e != nil {
		panic(e)
	}
	return p
}

const patchPrefix = "npm/node_modules/pi-subagents/"

func knownPatch(rel, sha string) bool {
	if !strings.HasPrefix(rel, patchPrefix) {
		return false
	}
	p, ok := patches().Files[strings.TrimPrefix(rel, patchPrefix)]
	return ok && p.SHA256After == sha
}
func reversePatches(root string, claims map[string]bool) ([]Edit, error) {
	var out []Edit
	activePatch := false
	for rel, hashes := range patches().Files {
		b, _, err := read(filepath.Join(root, patchPrefix, rel))
		if err != nil {
			return nil, err
		}
		if b != nil && (digest(b) == hashes.SHA256After || bytes.Contains(b, []byte("PI_HIVE_AGENT"))) {
			activePatch = true
		}
	}
	for rel, hashes := range patches().Files {
		p := filepath.Join(root, patchPrefix, rel)
		b, mode, err := read(p)
		if err != nil {
			return nil, err
		}
		if b == nil {
			continue
		}
		h := digest(b)
		if h == hashes.SHA256Before {
			continue
		}
		if h != hashes.SHA256After {
			if claims[p] || activePatch {
				return nil, fmt.Errorf("modified Hive package patch: %s", p)
			}
			continue
		}
		restored, err := reverseUnified(b, rel)
		if err != nil {
			return nil, err
		}
		if digest(restored) != hashes.SHA256Before {
			return nil, fmt.Errorf("inverse patch hash mismatch: %s", p)
		}
		out = append(out, Edit{Path: p, Before: b, After: restored, Mode: uint32(mode), Hosts: []string{"pi"}})
	}
	return out, nil
}

// Reverse only the embedded patch and validate every context line. The final
// pristine hash is checked by the caller; no external patch program executes.
func reverseUnified(b []byte, relative string) ([]byte, error) {
	lines := strings.SplitAfter(patchText, "\n")
	src := strings.SplitAfter(string(b), "\n")
	if src[len(src)-1] == "" {
		src = src[:len(src)-1]
	}
	var out []string
	cursor := 0
	active := false
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "+++ ") {
			active = strings.TrimSpace(strings.TrimPrefix(line, "+++ b/")) == relative
			continue
		}
		if !active || !strings.HasPrefix(line, "@@ ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("bad embedded patch")
		}
		position := strings.Split(strings.TrimPrefix(fields[2], "+"), ",")[0]
		n, e := strconv.Atoi(position)
		if e != nil || n < 1 || n-1 < cursor || n-1 > len(src) {
			return nil, fmt.Errorf("bad patch position")
		}
		out = append(out, src[cursor:n-1]...)
		cursor = n - 1
		for i++; i < len(lines); i++ {
			l := lines[i]
			if l == "" || strings.HasPrefix(l, "@@ ") || strings.HasPrefix(l, "diff ") || strings.HasPrefix(l, "--- ") {
				i--
				break
			}
			if l[0] == '\\' {
				return nil, fmt.Errorf("unsupported patch newline")
			}
			switch l[0] {
			case ' ', '+':
				if cursor >= len(src) || src[cursor] != l[1:] {
					return nil, fmt.Errorf("patch context mismatch for %s", relative)
				}
				cursor++
				if l[0] == ' ' {
					out = append(out, l[1:])
				}
			case '-':
				out = append(out, l[1:])
			default:
				return nil, fmt.Errorf("bad patch line")
			}
		}
	}
	out = append(out, src[cursor:]...)
	return []byte(strings.Join(out, "")), nil
}

// Reject duplicate keys before editing JSON so a shadowed user value cannot be lost.
func uniqueJSON(b []byte) error {
	d := json.NewDecoder(bytes.NewReader(b))
	var value func() error
	value = func() error {
		t, e := d.Token()
		if e != nil {
			return e
		}
		switch t {
		case json.Delim('{'):
			seen := map[string]bool{}
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return e
				}
				s, ok := k.(string)
				if !ok || seen[s] {
					return fmt.Errorf("duplicate key")
				}
				seen[s] = true
				if e = value(); e != nil {
					return e
				}
			}
			_, e = d.Token()
			return e
		case json.Delim('['):
			for d.More() {
				if e = value(); e != nil {
					return e
				}
			}
			_, e = d.Token()
			return e
		}
		return nil
	}
	return value()
}

// topSpan uses the decoder's offsets to locate one value without matching text
// inside unrelated strings or changing any surrounding formatting.
func topSpan(b []byte, key string) (start, end int, err error) {
	d := json.NewDecoder(bytes.NewReader(b))
	t, e := d.Token()
	if e != nil || t != json.Delim('{') {
		return 0, 0, fmt.Errorf("expected object")
	}
	for d.More() {
		k, e := d.Token()
		if e != nil {
			return 0, 0, e
		}
		at := int(d.InputOffset())
		for at < len(b) && (b[at] == ' ' || b[at] == '\n' || b[at] == '\r' || b[at] == '\t' || b[at] == ':') {
			at++
		}
		var raw json.RawMessage
		if e = d.Decode(&raw); e != nil {
			return 0, 0, e
		}
		if k == key {
			return at, int(d.InputOffset()), nil
		}
	}
	return 0, 0, fmt.Errorf("key missing")
}
func replaceTopValue(b []byte, key string, value []byte) ([]byte, error) {
	start, end, e := topSpan(b, key)
	if e != nil {
		return nil, e
	}
	out := append([]byte{}, b[:start]...)
	out = append(out, value...)
	out = append(out, b[end:]...)
	return out, nil
}

func cleanPiConfig(path, relative string, managed json.RawMessage, root string) ([]Edit, error) {
	var owned map[string]json.RawMessage
	if err := json.Unmarshal(managed, &owned); err != nil {
		return nil, fmt.Errorf("invalid Pi managed configuration: %s", path)
	}
	allowed := map[string][]string{"settings.json": {"packages", "shellPath"}, "mcp.json": {"mcpServers", "settings"}, "web-search.json": {"provider", "openaiSearchProviders", "workflow"}, "extensions/subagent/config.json": {"forceTopLevelAsync"}}
	for k := range owned {
		valid := false
		for _, x := range allowed[relative] {
			valid = valid || x == k
		}
		if !valid {
			return nil, fmt.Errorf("unsupported Pi configuration ownership: %s", path)
		}
	}
	b, mode, err := read(path)
	if err != nil || b == nil {
		return nil, err
	}
	if err = uniqueJSON(b); err != nil {
		return nil, fmt.Errorf("invalid Pi configuration: %s", path)
	}
	var doc map[string]json.RawMessage
	if err = json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("invalid Pi configuration object: %s", path)
	}
	// Package pins, provider preferences and shell paths remain user's tooling.
	// Explicit references to removed runtime code cannot remain active unnoticed.
	for _, marker := range []string{"hive-hooks", "hive/reviewer-guard", filepath.Join(root, "src"), filepath.Join(root, "global", "hooks")} {
		if bytes.Contains(b, []byte(marker)) {
			return nil, fmt.Errorf("Pi configuration references legacy runtime; resolve before migration: %s", path)
		}
	}
	if relative != "extensions/subagent/config.json" || len(owned) == 0 {
		return nil, nil
	}
	var was bool
	if err = json.Unmarshal(owned["forceTopLevelAsync"], &was); err != nil || !was {
		return nil, fmt.Errorf("unsupported Pi async configuration ownership: %s", path)
	}
	value, ok := doc["forceTopLevelAsync"]
	if !ok || !bytes.Equal(bytes.TrimSpace(value), []byte("true")) {
		return nil, nil
	}
	after, err := removeTopKey(b, "forceTopLevelAsync")
	if err != nil {
		return nil, err
	}
	return []Edit{{Path: path, Before: b, After: after, Mode: uint32(mode), Hosts: []string{"pi"}}}, nil
}
func removeTopKey(b []byte, key string) ([]byte, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	if _, e := d.Token(); e != nil {
		return nil, e
	}
	for d.More() {
		start := int(d.InputOffset())
		k, e := d.Token()
		if e != nil {
			return nil, e
		}
		var raw json.RawMessage
		if e = d.Decode(&raw); e != nil {
			return nil, e
		}
		end := int(d.InputOffset())
		if k != key {
			continue
		}
		for start < len(b) && (b[start] == ' ' || b[start] == '\n' || b[start] == '\r' || b[start] == '\t') {
			start++
		}
		if b[start] == ',' {
			out := append([]byte{}, b[:start]...)
			return append(out, b[end:]...), nil
		}
		tail := end
		for tail < len(b) && (b[tail] == ' ' || b[tail] == '\n' || b[tail] == '\r' || b[tail] == '\t') {
			tail++
		}
		if tail < len(b) && b[tail] == ',' {
			end = tail + 1
		}
		out := append([]byte{}, b[:start]...)
		out = append(out, b[end:]...)
		return out, nil
	}
	return nil, fmt.Errorf("key missing")
}
