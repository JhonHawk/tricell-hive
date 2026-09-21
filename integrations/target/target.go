// Package target describes filesystem destinations, not host model behavior.
package target

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Scope        string `json:"scope"`
	Home         string `json:"home"`
	Root         string `json:"root,omitempty"`
	CodexHome    string `json:"codex_home"`
	ClaudeHome   string `json:"claude_home"`
	PiHome       string `json:"pi_home,omitempty"`
	GrokHome     string `json:"grok_home,omitempty"`
	OpenCodeHome string `json:"opencode_home,omitempty"`
	Synthetic    bool   `json:"synthetic,omitempty"`
}

type Target struct {
	Path       string
	Kind       string
	Host       string
	Scope      string
	Context    string
	LinkTarget string `json:",omitempty"`
}

// ExpandHostHomes initializes and canonicalizes the home paths for hosts whose
// native locations are environment-overridable. A synthetic home deliberately
// ignores the current process environment so plans remain isolated and stable.
func ExpandHostHomes(c Config, synthetic bool) (Config, error) {
	c.Synthetic = synthetic
	if c.Home == "" {
		return c, fmt.Errorf("home is required")
	}
	home, err := Canonical(c.Home)
	if err != nil {
		return c, err
	}
	c.Home = home

	resolve := func(current, envName, fallback string) (string, error) {
		path := current
		if path == "" && !synthetic && envName != "" {
			path = os.Getenv(envName)
		}
		if path == "" {
			path = fallback
		}
		return Canonical(path)
	}
	c.PiHome, err = resolve(c.PiHome, "PI_CODING_AGENT_DIR", filepath.Join(home, ".pi", "agent"))
	if err != nil {
		return c, fmt.Errorf("resolve Pi home: %w", err)
	}
	c.GrokHome, err = resolve(c.GrokHome, "GROK_HOME", filepath.Join(home, ".grok"))
	if err != nil {
		return c, fmt.Errorf("resolve Grok home: %w", err)
	}
	opencodeDefault := filepath.Join(home, ".config", "opencode")
	if c.OpenCodeHome == "" && !synthetic {
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			if !filepath.IsAbs(xdg) {
				return c, fmt.Errorf("XDG_CONFIG_HOME must be an absolute path")
			}
			opencodeDefault = filepath.Join(xdg, "opencode")
		}
	}
	c.OpenCodeHome, err = resolve(c.OpenCodeHome, "", opencodeDefault)
	if err != nil {
		return c, fmt.Errorf("resolve OpenCode home: %w", err)
	}
	return c, nil
}

// Safe rejects links and non-directory ancestors. Call again before mutation.
func Safe(path string) error {
	path = filepath.Clean(path)
	for p := path; ; p = filepath.Dir(p) {
		info, err := os.Lstat(p)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink conflict: %s", p)
			}
			if p != path && !info.IsDir() {
				return fmt.Errorf("non-directory ancestor: %s", p)
			}
			if p == path && !info.IsDir() && !info.Mode().IsRegular() {
				return fmt.Errorf("non-regular target: %s", p)
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		if p == filepath.Dir(p) {
			break
		}
	}
	return nil
}

// Canonical resolves the existing prefix of an explicitly selected root.
func Canonical(path string) (string, error) {
	p, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var suffix []string
	for {
		_, err = os.Lstat(p)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		suffix = append(suffix, filepath.Base(p))
		p = filepath.Dir(p)
	}
	p, err = filepath.EvalSymlinks(p)
	if err != nil {
		return "", err
	}
	for i := len(suffix) - 1; i >= 0; i-- {
		p = filepath.Join(p, suffix[i])
	}
	return p, nil
}
