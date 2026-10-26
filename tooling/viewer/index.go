package viewer

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	indexVersion   = 1
	maxEngramBytes = 64 << 10
	maxNameRunes   = 200
)

// HiveProjectReader returns the `Project:` value of the `## Hive` section of
// the repository at root, and whether there is one. tooling/cli injects an
// implementation built on parseHiveSection; nil means "no section".
type HiveProjectReader func(root string) (string, bool)

// GitRemoteReader returns the URL of the repository's primary remote (for
// example `origin`). tooling/cli injects one built on its git runner; nil, or
// an error, means "no remote".
type GitRemoteReader func(root string) (string, error)

// Session is one stored copy of a walkthrough, as the index describes it.
type Session struct {
	// Key is the 64-hex SHA-256 of "<source folder>\x00<session id>".
	Key string `json:"key"`
	// Source is the folder the copy came from (absolute).
	Source       string `json:"source"`
	SourceStatus string `json:"source_status"`
	SourceReason string `json:"source_reason,omitempty"`
	// Repository is the root of the registered repository that holds Source.
	Repository  string    `json:"repository"`
	ContentHash string    `json:"content_hash"`
	IngestedAt  time.Time `json:"ingested_at"`
	Pinned      bool      `json:"pinned"`
	Archived    bool      `json:"archived"`

	SessionID     string `json:"session_id"`
	SessionStatus string `json:"session_status"`
	Host          string `json:"host"`
	Ticket        string `json:"ticket"`
	Title         string `json:"title"`
	Kind          string `json:"kind"`
	CreatedAt     string `json:"created_at"`

	// ProjectKey and ProjectName are the grouping resolved for Repository.
	ProjectKey  string `json:"project_key"`
	ProjectName string `json:"project_name"`
}

// Tombstone remembers a deleted or purged copy so the next scan does not
// ingest the same content again. Source is an addition to the {key,
// content_hash} of design DS4: it lets a scan tell that the source is gone.
type Tombstone struct {
	Key         string    `json:"key"`
	ContentHash string    `json:"content_hash"`
	Source      string    `json:"source"`
	DeletedAt   time.Time `json:"deleted_at"`
}

// IngestError is a source that could not be ingested, or a stored copy that
// could not be read.
type IngestError struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// RepositoryStatus reports one registered repository.
type RepositoryStatus struct {
	Root   string `json:"root"`
	Absent bool   `json:"absent"`
}

// indexData is the content of index.json. A published value is never
// modified: writers clone it, change the clone and publish that.
type indexData struct {
	Version      int                `json:"version"`
	Sessions     []Session          `json:"sessions"`
	Tombstones   []Tombstone        `json:"tombstones"`
	Errors       []IngestError      `json:"errors"`
	Repositories []RepositoryStatus `json:"repositories"`
}

func newIndexData() *indexData {
	return &indexData{
		Version:      indexVersion,
		Sessions:     []Session{},
		Tombstones:   []Tombstone{},
		Errors:       []IngestError{},
		Repositories: []RepositoryStatus{},
	}
}

func (d *indexData) clone() *indexData {
	c := newIndexData()
	c.Sessions = append(c.Sessions, d.Sessions...)
	c.Tombstones = append(c.Tombstones, d.Tombstones...)
	c.Errors = append(c.Errors, d.Errors...)
	c.Repositories = append(c.Repositories, d.Repositories...)
	return c
}

func (d *indexData) session(key string) (int, bool) {
	for i, s := range d.Sessions {
		if s.Key == key {
			return i, true
		}
	}
	return 0, false
}

func (d *indexData) tombstone(key string) (int, bool) {
	for i, t := range d.Tombstones {
		if t.Key == key {
			return i, true
		}
	}
	return 0, false
}

func (d *indexData) dropTombstone(key string) {
	if i, ok := d.tombstone(key); ok {
		d.Tombstones = append(d.Tombstones[:i], d.Tombstones[i+1:]...)
	}
}

// sorted orders the slices so index.json is deterministic.
func (d *indexData) sorted() {
	sort.Slice(d.Sessions, func(i, j int) bool { return d.Sessions[i].Key < d.Sessions[j].Key })
	sort.Slice(d.Tombstones, func(i, j int) bool { return d.Tombstones[i].Key < d.Tombstones[j].Key })
	sort.Slice(d.Errors, func(i, j int) bool { return d.Errors[i].Path < d.Errors[j].Path })
}

// View is the grouped read model behind GET /api/index: projects, their
// tickets, their normal sessions, excluded copies, ingest errors and repositories.
type View struct {
	ExcludedSessions []Session          `json:"excluded_sessions"`
	TTLSeconds       int64              `json:"ttl_seconds"`
	Projects         []Project          `json:"projects"`
	Repositories     []RepositoryStatus `json:"repositories"`
	Errors           []IngestError      `json:"errors"`
	Tombstones       []Tombstone        `json:"tombstones"`
}

// Project groups the sessions of one repository identity (the normalized Git
// remote, or the root path when there is none).
type Project struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	// Absent is true when every repository of the project is missing from
	// disk; its copies are kept.
	Absent  bool     `json:"absent"`
	Tickets []Ticket `json:"tickets"`
}

// Ticket groups the sessions that worked one ticket. An empty Name is the
// "no ticket" group and sorts last.
type Ticket struct {
	Name     string    `json:"ticket"`
	Sessions []Session `json:"sessions"`
}

// View returns the current grouped index. It never waits for an ingestion in
// progress.
func (s *Store) View() View {
	s.viewMu.RLock()
	d := s.idx
	s.viewMu.RUnlock()
	return buildView(d, s.ttl)
}

func buildView(d *indexData, ttl time.Duration) View {
	v := View{
		ExcludedSessions: []Session{},
		TTLSeconds:       int64(ttl / time.Second),
		Projects:         []Project{},
		Repositories:     append([]RepositoryStatus{}, d.Repositories...),
		Errors:           append([]IngestError{}, d.Errors...),
		Tombstones:       append([]Tombstone{}, d.Tombstones...),
	}
	present := map[string]bool{}
	for _, r := range d.Repositories {
		present[r.Root] = !r.Absent
	}
	sessions := append([]Session(nil), d.Sessions...)
	sort.SliceStable(sessions, func(i, j int) bool {
		if !sessions[i].IngestedAt.Equal(sessions[j].IngestedAt) {
			return sessions[i].IngestedAt.After(sessions[j].IngestedAt)
		}
		return sessions[i].Key < sessions[j].Key
	})
	projects := map[string]*Project{}
	tickets := map[string]map[string]*Ticket{}
	var order []string
	for _, sess := range sessions {
		if sess.SourceStatus == "tracked" {
			v.ExcludedSessions = append(v.ExcludedSessions, sess)
			continue
		}
		p, ok := projects[sess.ProjectKey]
		if !ok {
			p = &Project{Key: sess.ProjectKey, Name: sess.ProjectName, Absent: true, Tickets: []Ticket{}}
			projects[sess.ProjectKey] = p
			tickets[sess.ProjectKey] = map[string]*Ticket{}
			order = append(order, sess.ProjectKey)
		}
		if present[sess.Repository] {
			p.Absent = false
		}
		t, ok := tickets[sess.ProjectKey][sess.Ticket]
		if !ok {
			t = &Ticket{Name: sess.Ticket, Sessions: []Session{}}
			tickets[sess.ProjectKey][sess.Ticket] = t
		}
		t.Sessions = append(t.Sessions, sess)
	}
	for _, key := range order {
		p := projects[key]
		for _, t := range tickets[key] {
			p.Tickets = append(p.Tickets, *t)
		}
		sort.Slice(p.Tickets, func(i, j int) bool {
			a, b := p.Tickets[i].Name, p.Tickets[j].Name
			if (a == "") != (b == "") {
				return b == ""
			}
			return a < b
		})
		v.Projects = append(v.Projects, *p)
	}
	sort.Slice(v.Projects, func(i, j int) bool {
		a, b := strings.ToLower(v.Projects[i].Name), strings.ToLower(v.Projects[j].Name)
		if a != b {
			return a < b
		}
		return v.Projects[i].Key < v.Projects[j].Key
	})
	return v
}

// NormalizeRemote reduces a Git remote URL to `host/owner/repo`: no scheme,
// credentials, `.git` suffix or trailing slash, and a lowercase host. The
// SSH port is dropped (it does not identify the repository); an http(s) port
// is kept. It returns "" for an empty value, a local path, or anything
// without both a host and a path.
func NormalizeRemote(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.ContainsAny(raw, " \t\r\n") {
		return ""
	}
	var host, path string
	if scheme, rest, ok := strings.Cut(raw, "://"); ok {
		u, err := url.Parse(strings.ToLower(scheme) + "://" + rest)
		if err != nil || u.Hostname() == "" {
			return ""
		}
		host = strings.ToLower(u.Hostname())
		if (u.Scheme == "http" || u.Scheme == "https") && u.Port() != "" {
			host += ":" + u.Port()
		}
		path = u.Path
	} else {
		// scp-like: [user@]host:path
		if strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, ".") {
			return ""
		}
		hostPart, rest, ok := strings.Cut(raw, ":")
		if !ok || strings.Contains(hostPart, "/") {
			return ""
		}
		if at := strings.LastIndex(hostPart, "@"); at >= 0 {
			hostPart = hostPart[at+1:]
		}
		host, path = strings.ToLower(hostPart), rest
	}
	path = strings.Trim(path, "/")
	path = strings.TrimSuffix(path, ".git")
	path = strings.Trim(path, "/")
	if host == "" || path == "" {
		return ""
	}
	return host + "/" + path
}

type projectInfo struct{ key, name string }

// resolveProject applies the AC12 precedence for the repository at root.
// Key: normalized remote (only for a Git toplevel), else the root path. Name: the `Project:` value,
// else Engram's project_name, else the last segment of the remote, else the
// folder name.
func (s *Store) resolveProject(root string) projectInfo {
	remote := ""
	// Only a Git toplevel has a remote of its own: a root that merely lives
	// inside another repository (a non-Git workspace in a versioned home)
	// must not borrow that repository's remote.
	if _, gitErr := os.Lstat(filepath.Join(root, ".git")); s.remote != nil && gitErr == nil {
		if raw, err := s.remote(root); err == nil {
			remote = NormalizeRemote(raw)
		}
	}
	info := projectInfo{key: root}
	if remote != "" {
		info.key = remote
	}
	if s.hive != nil {
		if name, ok := s.hive(root); ok {
			if name = cleanName(name); name != "" {
				info.name = name
				return info
			}
		}
	}
	if name := readEngramName(root); name != "" {
		info.name = name
		return info
	}
	if remote != "" {
		info.name = remote[strings.LastIndex(remote, "/")+1:]
		return info
	}
	info.name = filepath.Base(root)
	return info
}

// cleanName drops control characters and bounds the length: the name comes
// from files the repository owner controls and ends up in the UI.
func cleanName(name string) string {
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(name)
	if r := []rune(name); len(r) > maxNameRunes {
		name = string(r[:maxNameRunes])
	}
	return name
}

// readEngramName reads only `project_name` from <root>/.engram/config.json:
// a regular file of at most 64 KiB, reached without following links.
func readEngramName(root string) string {
	dir := filepath.Join(root, ".engram")
	if info, err := os.Lstat(dir); err != nil || !info.IsDir() {
		return ""
	}
	path := filepath.Join(dir, "config.json")
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxEngramBytes {
		return ""
	}
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxEngramBytes+1))
	if err != nil || len(data) > maxEngramBytes {
		return ""
	}
	var cfg struct {
		ProjectName any `json:"project_name"`
	}
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&cfg); err != nil {
		return ""
	}
	name, _ := cfg.ProjectName.(string)
	return cleanName(name)
}

// repoRootOf maps a source folder `<root>/_support/workspace/<name>` to its
// repository root.
func repoRootOf(source string) string {
	ws := filepath.Dir(source)
	support := filepath.Dir(ws)
	if filepath.Base(ws) == "workspace" && filepath.Base(support) == "_support" {
		return filepath.Dir(support)
	}
	return source
}

// projectRefresh is how long a resolved project is trusted before the
// readers run again.
const projectRefresh = 60 * time.Second

type cachedProject struct {
	info projectInfo
	at   time.Time
}

func (s *Store) clearProjects() {
	s.projMu.Lock()
	s.projects, s.projRepos = nil, ""
	s.projMu.Unlock()
}

// noteRepositories drops the cached projects when the set of registered
// roots changed since the last scan.
func (s *Store) noteRepositories(reg Registry) {
	roots := make([]string, 0, len(reg.Repositories))
	for _, r := range reg.Repositories {
		roots = append(roots, r.Root)
	}
	sort.Strings(roots)
	joined := strings.Join(roots, "\x00")
	s.projMu.Lock()
	if joined != s.projRepos {
		s.projects, s.projRepos = nil, joined
	}
	s.projMu.Unlock()
}

// projectFor returns the project of root, resolved at most once per
// projectRefresh.
func (s *Store) projectFor(root string) projectInfo {
	now := s.now()
	s.projMu.Lock()
	if c, ok := s.projects[root]; ok && now.Sub(c.at) < projectRefresh && !now.Before(c.at) {
		s.projMu.Unlock()
		return c.info
	}
	s.projMu.Unlock()
	info := s.resolveProject(root)
	s.projMu.Lock()
	if s.projects == nil {
		s.projects = map[string]cachedProject{}
	}
	s.projects[root] = cachedProject{info: info, at: now}
	s.projMu.Unlock()
	return info
}

// enrich refreshes project information without replacing a previously resolved
// remote identity with a path fallback when the repository cannot be read.
func (s *Store) enrich(d *indexData) {
	known := map[string]projectInfo{}
	for _, sess := range d.Sessions {
		root := repoRootOf(sess.Source)
		if sess.Repository == root && sess.ProjectKey != "" && sess.ProjectKey != root && NormalizeRemote("https://"+sess.ProjectKey) == sess.ProjectKey {
			known[root] = projectInfo{key: sess.ProjectKey, name: sess.ProjectName}
		}
	}
	for i := range d.Sessions {
		root := repoRootOf(d.Sessions[i].Source)
		info := s.projectFor(root)
		if previous, ok := known[root]; ok && info.key == root {
			info = previous
		}
		d.Sessions[i].Repository = root
		d.Sessions[i].ProjectKey = info.key
		d.Sessions[i].ProjectName = info.name
	}
}
