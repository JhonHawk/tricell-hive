package viewer

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const (
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

// IngestError is a source that could not be ingested, or a stored copy that
// could not be read.
type IngestError struct {
	Code         string `json:"code,omitempty"`
	SessionGroup string `json:"session_group,omitempty"`
	Path         string `json:"path"`
	Reason       string `json:"reason"`
}

// RepositoryStatus reports one registered repository.
type RepositoryStatus struct {
	Root   string `json:"root"`
	Absent bool   `json:"absent"`
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
