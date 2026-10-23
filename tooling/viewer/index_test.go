package viewer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestIndexProjectKeyIsTheNormalizedRemoteOrTheRootPath(t *testing.T) {
	e := newEnv(t)
	withRemote, withoutRemote := newRepo(t), newRepo(t)
	e.remotes[withRemote] = "git@github.com:Owner/Repo.git"
	e.place(withRemote, "a", nil)
	e.place(withoutRemote, "b", nil)
	e.register(withRemote)
	e.register(withoutRemote)
	e.start()
	e.sync()
	if p := projectOf(t, e, withRemote); p.Key != "github.com/Owner/Repo" {
		t.Fatalf("remote key = %q", p.Key)
	}
	if p := projectOf(t, e, withoutRemote); p.Key != withoutRemote {
		t.Fatalf("root key = %q, want %q", p.Key, withoutRemote)
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
			e.place(repo, "a", nil)
			e.register(repo)
			e.start()
			e.sync()
			if got, want := projectOf(t, e, repo).Name, c.want(repo); got != want {
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
			e.place(repo, "a", nil)
			e.register(repo)
			e.start()
			e.sync()
			if got := projectOf(t, e, repo).Name; got != filepath.Base(repo) {
				t.Fatalf("name = %q, want the folder name", got)
			}
		})
	}
	t.Run("control characters", func(t *testing.T) {
		e := newEnv(t)
		repo := newRepo(t)
		os.MkdirAll(filepath.Join(repo, ".engram"), 0o755)
		os.WriteFile(filepath.Join(repo, ".engram", "config.json"), []byte(`{"project_name":"a\u0000b\nc\u001b[31md"}`), 0o644)
		e.place(repo, "a", nil)
		e.register(repo)
		e.start()
		e.sync()
		if got := projectOf(t, e, repo).Name; got != "abc[31md" {
			t.Fatalf("name = %q", got)
		}
	})
}

func TestIndexWorktreesOfOneRemoteShareTheProject(t *testing.T) {
	e := newEnv(t)
	main, worktree := newRepo(t), newRepo(t)
	e.remotes[main] = "git@github.com:o/r.git"
	e.remotes[worktree] = "https://github.com/o/r"
	e.hive[main] = "Shared Name"
	e.place(main, "a", func(m map[string]any) { m["ticket"] = "gh-1" })
	e.place(worktree, "b", func(m map[string]any) { m["ticket"] = "gh-1" })
	e.register(main)
	e.register(worktree)
	e.start()
	e.sync()
	view := e.store.View()
	if len(view.Projects) != 1 {
		t.Fatalf("projects = %d, want 1: %+v", len(view.Projects), view.Projects)
	}
	p := view.Projects[0]
	if p.Key != "github.com/o/r" || len(p.Tickets) != 1 || len(p.Tickets[0].Sessions) != 2 {
		t.Fatalf("project = %+v", p)
	}
}

func TestIndexGroupsByProjectThenTicketThenSessionNewestFirst(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.hive[repo] = "Proj"
	e.place(repo, "s1", func(m map[string]any) { m["ticket"] = "gh-67-skill" })
	e.advance(time.Minute)
	e.place(repo, "s2", func(m map[string]any) { m["ticket"] = "gh-67-skill" })
	e.advance(time.Minute)
	e.place(repo, "s3", func(m map[string]any) { m["ticket"] = "gh-75-viewer" })
	e.advance(time.Minute)
	e.place(repo, "s4", func(m map[string]any) { delete(m, "ticket") })
	e.register(repo)
	e.start()
	// Ingest one at a time so ingested_at differs.
	e.sync()

	view := e.store.View()
	if len(view.Projects) != 1 || view.Projects[0].Name != "Proj" {
		t.Fatalf("projects = %+v", view.Projects)
	}
	tickets := view.Projects[0].Tickets
	var names []string
	for _, tk := range tickets {
		names = append(names, tk.Name)
	}
	if strings.Join(names, ",") != "gh-67-skill,gh-75-viewer," {
		t.Fatalf("ticket order = %q (named tickets alphabetically, no-ticket group last)", strings.Join(names, ","))
	}
	if len(tickets[0].Sessions) != 2 || len(tickets[1].Sessions) != 1 || len(tickets[2].Sessions) != 1 {
		t.Fatalf("tickets = %+v", tickets)
	}
	if tickets[2].Sessions[0].Ticket != "" {
		t.Fatal("a manifest without ticket must land in the no-ticket group")
	}
}

func TestIndexSessionsInATicketAreNewestFirst(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.register(repo)
	e.start()
	for _, name := range []string{"old", "mid", "new"} {
		e.place(repo, name, func(m map[string]any) { m["ticket"] = "gh-1" })
		e.sync()
		e.advance(time.Minute)
	}
	var ids []string
	for _, s := range e.store.View().Projects[0].Tickets[0].Sessions {
		ids = append(ids, s.SessionID)
	}
	if strings.Join(ids, ",") != "sess-new,sess-mid,sess-old" {
		t.Fatalf("order = %v", ids)
	}
}

func TestIndexListsRegisteredRepositoriesEvenWithoutSessions(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.register(repo)
	e.start()
	e.sync()
	view := e.store.View()
	if len(view.Projects) != 0 || len(view.Repositories) != 1 || view.Repositories[0].Root != repo || view.Repositories[0].Absent {
		t.Fatalf("view = %+v", view)
	}
	b, _ := json.Marshal(view)
	for _, field := range []string{`"projects":[]`, `"errors":[]`, `"tombstones":[]`} {
		if !strings.Contains(string(b), field) {
			t.Fatalf("empty collections must encode as []: missing %s in %s", field, b)
		}
	}
}

func TestIndexIsRebuiltFromMetaWhenTheFileIsMissingOrCorrupt(t *testing.T) {
	for name, damage := range map[string]func(path string){
		"missing": func(path string) { os.Remove(path) },
		"corrupt": func(path string) { os.WriteFile(path, []byte("{not json"), 0o600) },
		"truncated": func(path string) {
			data, _ := os.ReadFile(path)
			os.WriteFile(path, data[:len(data)/2], 0o600)
		},
	} {
		t.Run(name, func(t *testing.T) {
			e := newEnv(t)
			repo := newRepo(t)
			e.hive[repo] = "Proj"
			e.place(repo, "a", func(m map[string]any) { m["ticket"] = "gh-9" })
			e.place(repo, "b", nil)
			e.register(repo)
			e.start()
			e.sync()
			keys := e.sessions()
			if err := e.store.Pin(keys[0].Key, true); err != nil {
				t.Fatal(err)
			}
			before := e.store.View()

			damage(filepath.Join(e.state, "viewer", "index.json"))
			e.store = e.open()
			e.start()
			after := e.store.View()
			if len(e.sessions()) != 2 {
				t.Fatalf("rebuilt sessions = %d", len(e.sessions()))
			}
			bj, _ := json.Marshal(before.Projects)
			aj, _ := json.Marshal(after.Projects)
			if string(bj) != string(aj) {
				t.Fatalf("rebuilt projects differ:\nbefore %s\nafter  %s", bj, aj)
			}
			// The next pass finds everything up to date: no duplicate copies.
			if res := e.sync(); res.Added != 0 || res.Replaced != 0 {
				t.Fatalf("rebuild caused re-ingestion: %+v", res)
			}
		})
	}
}

func TestIndexRebuildReportsAnOrphanCopyWithoutDeletingIt(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "a", nil)
	e.register(repo)
	e.start()
	e.sync()
	key := e.onlySession().Key
	os.WriteFile(filepath.Join(e.sessionDir(key), "meta.json"), []byte("{broken"), 0o600)
	e.store = e.open()
	e.start()
	view := e.store.View()
	if len(e.sessions()) != 0 || len(view.Errors) != 1 || !strings.Contains(view.Errors[0].Path, key) {
		t.Fatalf("view = %+v", view)
	}
	if _, err := os.Stat(e.sessionDir(key)); err != nil {
		t.Fatal("an unreadable copy must be reported, not deleted")
	}
}

func TestIndexFileHoldsSessionsTombstonesAndErrors(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "a", nil)
	e.place(repo, "b", nil)
	os.MkdirAll(filepath.Join(repo, "_support", "workspace", "bad"), 0o755)
	os.WriteFile(filepath.Join(repo, "_support", "workspace", "bad", "walkthrough.json"), []byte("{"), 0o644)
	e.register(repo)
	e.start()
	e.sync()
	e.store.Delete(e.sessions()[0].Key)
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(readFile(t, filepath.Join(e.state, "viewer", "index.json")), &doc); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"version", "sessions", "tombstones", "errors"} {
		if _, ok := doc[k]; !ok {
			t.Fatalf("index.json lacks %q: %v", k, doc)
		}
	}
	var sessions, tombs, errs []json.RawMessage
	json.Unmarshal(doc["sessions"], &sessions)
	json.Unmarshal(doc["tombstones"], &tombs)
	json.Unmarshal(doc["errors"], &errs)
	if len(sessions) != 1 || len(tombs) != 1 || len(errs) != 1 {
		t.Fatalf("sessions/tombstones/errors = %d/%d/%d", len(sessions), len(tombs), len(errs))
	}
	if info, _ := os.Stat(filepath.Join(e.state, "viewer", "index.json")); info.Mode().Perm() != 0o600 {
		t.Fatalf("index.json mode = %v", info.Mode().Perm())
	}
}

func TestIndexTombstonesSurviveARestartWhenTheIndexIsIntact(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "a", nil)
	e.register(repo)
	e.start()
	e.sync()
	e.store.Delete(e.onlySession().Key)
	e.store = e.open()
	e.start()
	if res := e.sync(); res.Added != 0 {
		t.Fatalf("a deleted session came back after a restart: %+v", res)
	}
}

func TestIndexWorkspaceRootWithoutGitIsScannedAndGroupedByPath(t *testing.T) {
	e := newEnv(t)
	root, work := newWorkspace(t)
	e.hive[root] = "demo-ws" // what the cli reader finds in <root>/AGENTS.md
	if res, err := RegisterRepository(e.state, work); err != nil || res.Root != root {
		t.Fatalf("register = %+v, %v", res, err)
	}
	e.start()
	res := e.sync() // the remote reader fails outside Git: no ingest error
	if res.Added != 1 || res.Errors != 0 || len(e.store.View().Errors) != 0 {
		t.Fatalf("result = %+v, errors %v", res, e.store.View().Errors)
	}
	view := e.store.View()
	if len(view.Projects) != 1 || view.Projects[0].Name != "demo-ws" || view.Projects[0].Key != root {
		t.Fatalf("projects = %+v, want demo-ws keyed by %s", view.Projects, root)
	}
	if s := e.onlySession(); s.Source != work || s.Repository != root {
		t.Fatalf("session = %+v", s)
	}
}

func TestIndexWorkspaceInsideGitUsesTheRemoteAsKey(t *testing.T) {
	e := newEnv(t)
	root, work := newWorkspace(t)
	os.MkdirAll(filepath.Join(root, ".git"), 0o755) // the workspace root is itself a Git toplevel
	e.remotes[root] = "git@github.com:o/ws.git"
	e.register(work)
	e.start()
	e.sync()
	if p := e.store.View().Projects[0]; p.Key != "github.com/o/ws" || p.Name != "ws" {
		t.Fatalf("project = %+v", p)
	}
}

func TestIndexWorkspaceInsideAnUnrelatedGitRepositoryHasNoRemote(t *testing.T) {
	e := newEnv(t)
	outer := newRepo(t) // a versioned parent with a remote
	root := filepath.Join(outer, "ws")
	work := filepath.Join(root, "_support", "workspace", "w")
	copyTree(t, filepath.Join(fixturesRoot, "valid", "final"), work)
	// A reader that, like git run inside root, would answer with the
	// enclosing repository's remote.
	e.remotes[root] = "git@github.com:o/outer.git"
	e.hive[root] = "ws-name"
	e.register(work)
	e.start()
	e.sync()
	p := e.store.View().Projects[0]
	if p.Key != root || p.Name != "ws-name" {
		t.Fatalf("project = %+v; the key must be the root path %s, not the outer repository's remote", p, root)
	}
}

func (e *env) calls(root string) (hive, remote int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.hiveCalls[root], e.remoteCalls[root]
}

func TestIndexProjectInfoIsCachedAndRefreshedOnAnInterval(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.remotes[repo] = "git@github.com:o/r.git"
	e.hive[repo] = "Cached"
	e.place(repo, "a", nil)
	e.register(repo)
	e.start()
	for i := 0; i < 4; i++ {
		e.sync()
	}
	if h, r := e.calls(repo); h != 1 || r != 1 {
		t.Fatalf("readers ran %d/%d times over 5 passes, want 1/1", h, r)
	}
	// A change is picked up once the refresh interval passes.
	e.hive[repo] = "Renamed"
	e.advance(30 * time.Second)
	e.sync()
	if got := projectOf(t, e, repo).Name; got != "Cached" {
		t.Fatalf("name changed before the refresh interval: %q", got)
	}
	e.advance(31 * time.Second)
	e.touch(filepath.Join(repo, "_support", "workspace", "a"))
	e.sync()
	if got := projectOf(t, e, repo).Name; got != "Renamed" {
		t.Fatalf("name = %q after the refresh interval", got)
	}
	if h, r := e.calls(repo); h != 2 || r != 2 {
		t.Fatalf("readers ran %d/%d times, want 2/2", h, r)
	}
}

func TestIndexProjectInfoIsRefreshedWhenTheRegistryChanges(t *testing.T) {
	e := newEnv(t)
	first, second := newRepo(t), newRepo(t)
	e.hive[first] = "One"
	e.place(first, "a", nil)
	e.register(first)
	e.start()
	e.sync()
	e.hive[first] = "One renamed"
	e.register(second)
	e.sync()
	if got := projectOf(t, e, first).Name; got != "One renamed" {
		t.Fatalf("registering a repository must refresh the cache, name = %q", got)
	}
}
