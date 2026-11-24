package viewer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode"
	"unicode/utf8"
)

// Navigation is a request-local presentation projection, never stored with a session.
type Navigation struct {
	Version int               `json:"version"`
	Entries []NavigationEntry `json:"entries"`
}

type NavigationEntry struct {
	ViewerSessionID string `json:"viewer_session_id"`
	GroupKey        string `json:"group_key"`
	GroupName       string `json:"group_name"`
	OriginKind      string `json:"origin_kind"`
	OriginLabel     string `json:"origin_label"`
	ConfigStatus    string `json:"config_status"`
}

type explicitProject struct{ name, status string }

// readExplicitProject deliberately does not share the permissive legacy display
// name reader: invalid declarations must not manufacture a grouping identity.
// All diagnostics are fixed statuses; no file content or unknown fields escape.
func readExplicitProject(root string) explicitProject {
	invalid := explicitProject{status: "invalid"}
	rootInfo, err := os.Stat(root)
	if err != nil || !rootInfo.IsDir() {
		return explicitProject{status: "unavailable"}
	}
	dir := filepath.Join(root, ".engram")
	dirInfo, err := os.Lstat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return explicitProject{status: "missing"}
	}
	if err != nil || !dirInfo.IsDir() {
		return invalid
	}
	// Anchor both lookups to the opened directory. Verify its identity before
	// reading, so replacement by a symlink cannot silently change the source.
	d, err := os.OpenRoot(dir)
	if err != nil {
		return invalid
	}
	defer d.Close()
	openedDir, err := d.Stat(".")
	if err != nil || !os.SameFile(dirInfo, openedDir) {
		return invalid
	}
	dirAfter, err := os.Lstat(dir)
	if err != nil || !dirAfter.IsDir() || !os.SameFile(dirInfo, dirAfter) {
		return invalid
	}
	info, err := d.Lstat("config.json")
	if errors.Is(err, os.ErrNotExist) {
		return explicitProject{status: "missing"}
	}
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxEngramBytes {
		return invalid
	}
	f, err := d.OpenFile("config.json", os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return invalid
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(info, after) {
		return invalid
	}
	data, err := io.ReadAll(io.LimitReader(f, maxEngramBytes+1))
	if err != nil || len(data) > maxEngramBytes || !utf8.Valid(data) {
		return invalid
	}
	var cfg map[string]json.RawMessage
	dec := json.NewDecoder(bytes.NewReader(data))
	if dec.Decode(&cfg) != nil || cfg == nil {
		return invalid
	}
	if _, err := dec.Token(); err != io.EOF {
		return invalid
	}
	var name string
	if json.Unmarshal(cfg["project_name"], &name) != nil {
		return invalid
	}
	// Check controls before trimming: a control at an edge is still invalid.
	for _, r := range name {
		if unicode.IsControl(r) || r == '/' || r == '\\' {
			return invalid
		}
	}
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxNameRunes {
		return invalid
	}
	return explicitProject{name: name, status: "valid"}
}

func navigationKey(namespace, value string) string {
	hash := sha256.Sum256([]byte(value))
	return namespace + ":" + hex.EncodeToString(hash[:])
}

func groupedNavigation(sessions []TimelineAPIIndexEntry, registry Registry) *Navigation {
	navigation := &Navigation{Version: 1, Entries: make([]NavigationEntry, 0, len(sessions))}
	registered := map[string]bool{}
	for _, repository := range registry.Repositories {
		registered[repository.Root] = true
	}
	type observation struct {
		config      explicitProject
		kind, label string
		missing     bool
	}
	roots := map[string]observation{}
	fallbackNames := map[string]string{}
	projectNames := map[string]string{}
	conflictingNames := map[string]bool{}
	retired := map[string]bool{}
	for _, session := range sessions {
		root := session.Source.ProjectRoot
		missing := false
		if canonical, err := canonicalDir(root); err == nil {
			root = canonical
		} else {
			missing = errors.Is(err, os.ErrNotExist)
		}
		observed, ok := roots[root]
		if !ok {
			observed = observation{config: explicitProject{status: "unavailable"}, kind: "workspace", label: filepath.Base(root)}
			if registered[root] {
				observed.config = readExplicitProject(root)
			}
			if observed.config.status == "unavailable" {
				observed.kind = "unavailable"
			} else if info, err := os.Lstat(filepath.Join(root, ".git")); err == nil && (info.IsDir() || info.Mode().IsRegular()) {
				observed.kind = "repository"
			}
			observed.missing = missing && registered[root]
			roots[root] = observed
		}
		retired[session.ViewerSessionID] = observed.missing
		if observed.config.status == "valid" {
			if name, exists := projectNames[session.ProjectKey]; exists && name != observed.config.name {
				conflictingNames[session.ProjectKey] = true
			}
			projectNames[session.ProjectKey] = observed.config.name
		}
		entry := NavigationEntry{ViewerSessionID: session.ViewerSessionID, GroupKey: navigationKey("project", session.ProjectKey), GroupName: session.ProjectName, OriginKind: observed.kind, OriginLabel: observed.label, ConfigStatus: observed.config.status}
		if observed.config.status == "valid" {
			entry.GroupKey = navigationKey("explicit", observed.config.name)
			entry.GroupName = observed.config.name
		} else {
			// Historical sessions can retain different display names for one
			// project identity. Use the first fallback row in index order for
			// this request, without changing any persisted session name.
			name, exists := fallbackNames[session.ProjectKey]
			if !exists {
				name = session.ProjectName
				fallbackNames[session.ProjectKey] = name
			}
			entry.GroupName = name
		}
		navigation.Entries = append(navigation.Entries, entry)
	}
	// A removed registered source may share presentation with an available
	// source of the exact stored project identity. Never infer identity from
	// equal display names, and never pick between conflicting declarations.
	for i, session := range sessions {
		entry := &navigation.Entries[i]
		name, exists := projectNames[session.ProjectKey]
		if retired[session.ViewerSessionID] && entry.ConfigStatus == "unavailable" && exists && !conflictingNames[session.ProjectKey] {
			entry.GroupKey = navigationKey("explicit", name)
			entry.GroupName = name
			entry.ConfigStatus = "valid"
		}
	}
	return navigation
}
