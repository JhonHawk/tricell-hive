package viewer

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIndexNormalizeRemote(t *testing.T) {
	cases := map[string]string{
		"git@github.com:JhonHawk/tricell-hive.git":                  "github.com/JhonHawk/tricell-hive",
		"https://github.com/JhonHawk/tricell-hive.git":              "github.com/JhonHawk/tricell-hive",
		"https://github.com/JhonHawk/tricell-hive":                  "github.com/JhonHawk/tricell-hive",
		"https://user:s3cret@github.com/JhonHawk/tricell-hive.git/": "github.com/JhonHawk/tricell-hive",
		"HTTPS://GitHub.com/o/r.git":                                "github.com/o/r",
		"ssh://git@gitlab.example.com:2222/group/sub/repo.git":      "gitlab.example.com/group/sub/repo",
		"git://host.example/o/r.git":                                "host.example/o/r",
		"https://host.example:8443/o/r":                             "host.example:8443/o/r",
		"  git@host.example:o/r.git\n":                              "host.example/o/r",
		"host.example:o/r":                                          "host.example/o/r",
		"":                                                          "",
		"/srv/git/local.git":                                        "",
		"../relative":                                               "",
		"https://github.com":                                        "",
		"https://github.com/":                                       "",
		"not a url":                                                 "",
	}
	for raw, want := range cases {
		if got := NormalizeRemote(raw); got != want {
			t.Errorf("NormalizeRemote(%q) = %q, want %q", raw, got, want)
		}
		if strings.Contains(NormalizeRemote(raw), "s3cret") {
			t.Errorf("credentials leaked for %q", raw)
		}
	}
}

// resolveStore is a Store with injected project readers, enough to resolve names.
func resolveStore(remotes, hive map[string]string) *Store {
	return &Store{
		remote: func(root string) (string, error) {
			if r, ok := remotes[root]; ok {
				return r, nil
			}
			return "", errors.New("no remote")
		},
		hive: func(root string) (string, bool) {
			n, ok := hive[root]
			return n, ok
		},
	}
}

func TestIndexDisplayNamePrecedence(t *testing.T) {
	cases := []struct {
		name   string
		remote string
		hive   string
		engram string // raw .engram/config.json, "" for none
		want   func(repo string) string
	}{
		{"hive section wins", "git@github.com:o/repo-name.git", "Hive Name", `{"project_name":"engram-name"}`, func(string) string { return "Hive Name" }},
		{"engram when no hive", "git@github.com:o/repo-name.git", "", `{"project_name":"engram-name"}`, func(string) string { return "engram-name" }},
		{"remote segment when no engram", "git@github.com:o/repo-name.git", "", "", func(string) string { return "repo-name" }},
		{"folder name without remote", "", "", "", func(repo string) string { return filepath.Base(repo) }},
		{"engram without remote", "", "", `{"project_name":"only-engram"}`, func(string) string { return "only-engram" }},
		{"empty engram name falls through", "git@github.com:o/repo-name.git", "", `{"project_name":"  "}`, func(string) string { return "repo-name" }},
		{"engram wrong type falls through", "", "", `{"project_name":7}`, func(repo string) string { return filepath.Base(repo) }},
		{"engram invalid json falls through", "", "", `{nope`, func(repo string) string { return filepath.Base(repo) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := newRepo(t)
			remotes, hive := map[string]string{}, map[string]string{}
			if c.remote != "" {
				remotes[repo] = c.remote
			}
			if c.hive != "" {
				hive[repo] = c.hive
			}
			store := resolveStore(remotes, hive)
			if c.engram != "" {
				os.MkdirAll(filepath.Join(repo, ".engram"), 0o755)
				os.WriteFile(filepath.Join(repo, ".engram", "config.json"), []byte(c.engram), 0o644)
			}

			if got, want := store.resolveProject(repo).name, c.want(repo); got != want {
				t.Fatalf("name = %q, want %q", got, want)
			}
		})
	}
}

func TestIndexEngramConfigIsBoundedAndSanitized(t *testing.T) {
	cases := map[string]func(repo string){
		"over 64 KiB": func(repo string) {
			big := `{"project_name":"huge","pad":"` + strings.Repeat("x", 64<<10) + `"}`
			os.WriteFile(filepath.Join(repo, ".engram", "config.json"), []byte(big), 0o644)
		},
		"symlink": func(repo string) {
			real := filepath.Join(t.TempDir(), "config.json")
			os.WriteFile(real, []byte(`{"project_name":"linked"}`), 0o644)
			os.Symlink(real, filepath.Join(repo, ".engram", "config.json"))
		},
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			store := resolveStore(nil, nil)
			repo := newRepo(t)
			os.MkdirAll(filepath.Join(repo, ".engram"), 0o755)
			setup(repo)

			if got := store.resolveProject(repo).name; got != filepath.Base(repo) {
				t.Fatalf("name = %q, want the folder name", got)
			}
		})
	}
	t.Run("control characters", func(t *testing.T) {
		store := resolveStore(nil, nil)
		repo := newRepo(t)
		os.MkdirAll(filepath.Join(repo, ".engram"), 0o755)
		os.WriteFile(filepath.Join(repo, ".engram", "config.json"), []byte(`{"project_name":"a\u0000b\nc\u001b[31md"}`), 0o644)

		if got := store.resolveProject(repo).name; got != "abc[31md" {
			t.Fatalf("name = %q", got)
		}
	})
}
