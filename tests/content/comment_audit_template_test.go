package content

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	commentAuditTemplateDir = "../../content/skills/comment-audit/assets/ratchet"
	commentAuditParsersDir  = "tests/content/testdata/comment-audit/node_modules"
	commentAuditInstall     = "npm ci --ignore-scripts --prefix tests/content/testdata/comment-audit"
)

// commentAuditParserVersions are the versions package.json pins; an install
// left from an older pin must not pass as the pinned one.
var commentAuditParserVersions = map[string]string{"babel8": "8.1.2", "babel7": "7.29.14"}

// TestCommentAuditTemplate runs the ratchet template's own node:test suite
// once per pinned @babel/parser major version, since projects copy the
// template and run it with whichever of the two their Node supports.
func TestCommentAuditTemplate(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node unavailable")
	}
	parsers := commentAuditParsers(t)
	for _, version := range []string{"babel8", "babel7"} {
		t.Run(version, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range []string{"check-comments.mjs", "check-comments.test.mjs"} {
				data, err := os.ReadFile(filepath.Join(commentAuditTemplateDir, name))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			scope := filepath.Join(dir, "node_modules", "@babel")
			if err := os.MkdirAll(scope, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(parsers, version), filepath.Join(scope, "parser")); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("node", "--test", "check-comments.test.mjs")
			cmd.Dir = dir
			cmd.Env = withoutGitVariables(os.Environ())
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out
			if err := cmd.Run(); err != nil {
				t.Fatalf("node --test with %s failed: %v\n%s", version, err, out.String())
			}
			t.Logf("%s:\n%s", version, summaryLines(out.String()))
		})
	}
}

// commentAuditParsers finds the installed pinned parsers in this checkout or,
// from a worktree that never installed them, in the main checkout.
func commentAuditParsers(t *testing.T) string {
	t.Helper()
	candidates := []string{filepath.Join("..", "..", commentAuditParsersDir)}
	if common, err := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-common-dir").Output(); err == nil {
		main := filepath.Dir(strings.TrimSpace(string(common)))
		candidates = append(candidates, filepath.Join(main, commentAuditParsersDir))
	}
	for _, candidate := range candidates {
		abs, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if installedVersionsMatch(abs) {
			return abs
		}
	}
	t.Skipf("pinned parsers %v not installed; run %s from the repository root", commentAuditParserVersions, commentAuditInstall)
	return ""
}

func installedVersionsMatch(dir string) bool {
	for alias, want := range commentAuditParserVersions {
		data, err := os.ReadFile(filepath.Join(dir, alias, "package.json"))
		if err != nil {
			return false
		}
		var manifest struct {
			Version string `json:"version"`
		}
		if json.Unmarshal(data, &manifest) != nil || manifest.Version != want {
			return false
		}
	}
	return true
}

func withoutGitVariables(environ []string) []string {
	var kept []string
	for _, entry := range environ {
		if !strings.HasPrefix(entry, "GIT_") {
			kept = append(kept, entry)
		}
	}
	return kept
}

func summaryLines(output string) string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "ℹ ") {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
