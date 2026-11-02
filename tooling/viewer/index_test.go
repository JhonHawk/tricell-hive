package viewer

import (
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

func projectOf(t *testing.T, e *env, repo string) Project {
	t.Helper()
	for _, p := range e.store.View().Projects {
		for _, tk := range p.Tickets {
			for _, s := range tk.Sessions {
				if s.Repository == repo {
					return p
				}
			}
		}
	}
	t.Fatalf("no project for %s", repo)
	return Project{}
}

func (e *env) calls(root string) (hive, remote int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.hiveCalls[root], e.remoteCalls[root]
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
			e := newEnv(t)
			repo := newRepo(t)
			if c.remote != "" {
				e.remotes[repo] = c.remote
			}
			if c.hive != "" {
				e.hive[repo] = c.hive
			}
			if c.engram != "" {
				os.MkdirAll(filepath.Join(repo, ".engram"), 0o755)
				os.WriteFile(filepath.Join(repo, ".engram", "config.json"), []byte(c.engram), 0o644)
			}

			if got, want := e.store.resolveProject(repo).name, c.want(repo); got != want {
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
			e := newEnv(t)
			repo := newRepo(t)
			os.MkdirAll(filepath.Join(repo, ".engram"), 0o755)
			setup(repo)

			if got := e.store.resolveProject(repo).name; got != filepath.Base(repo) {
				t.Fatalf("name = %q, want the folder name", got)
			}
		})
	}
	t.Run("control characters", func(t *testing.T) {
		e := newEnv(t)
		repo := newRepo(t)
		os.MkdirAll(filepath.Join(repo, ".engram"), 0o755)
		os.WriteFile(filepath.Join(repo, ".engram", "config.json"), []byte(`{"project_name":"a\u0000b\nc\u001b[31md"}`), 0o644)

		if got := e.store.resolveProject(repo).name; got != "abc[31md" {
			t.Fatalf("name = %q", got)
		}
	})
}
