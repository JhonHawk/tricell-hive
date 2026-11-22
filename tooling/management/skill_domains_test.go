package management

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const domainsManifestPath = "hive-core/assets/domains.json"

// domainOptions builds a source whose hive-core declares two domains and also
// holds a folder it does not declare.
func domainOptions(t *testing.T) Options {
	t.Helper()
	o := setup(t)
	o.Hosts = []string{"claude", "codex"}
	put(t, packetSource(o, "hive-core/SKILL.md"), "---\nname: hive-core\ndescription: Leaves.\ndisable-model-invocation: true\n---\nIndex.\n")
	put(t, packetSource(o, domainsManifestPath), `{"domains": ["git", "review"]}`)
	put(t, packetSource(o, "hive-core/git/git-commit.md"), "# Git commit\n\n## When\n\nBefore a commit.\n")
	put(t, packetSource(o, "hive-core/review/blast-radius.md"), "# Blast radius\n\nSee [commit](../git/git-commit.md#when) and [same](skill:hive-core/git/git-commit.md).\n")
	put(t, packetSource(o, "hive-core/notes/draft.md"), "# Draft\n")
	return o
}

func installedCore(o Options, rel string) string {
	return filepath.Join(o.Home, ".agents", "skills", "hive-core", filepath.FromSlash(rel))
}

func storeRelease(t *testing.T, o Options, r Release) Options {
	t.Helper()
	sort.SliceStable(r.Files, func(i, j int) bool {
		if r.Files[i].Path == GlobalSource || r.Files[j].Path == GlobalSource {
			return r.Files[i].Path == GlobalSource
		}
		return r.Files[i].Path < r.Files[j].Path
	})
	r.ID = releaseID(r)
	if err := os.MkdirAll(filepath.Join(o.StateDir, "releases"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(o.StateDir, "releases", r.ID+".json"), encode(r), 0o600); err != nil {
		t.Fatal(err)
	}
	stored := o
	stored.Source = ""
	stored.ReleaseID = r.ID
	return stored
}

func TestDomainFoldersListedInTheManifestAreInstalledAndTheRestIsNot(t *testing.T) {
	o := domainOptions(t)
	apply(t, plan(t, "install", o))
	for _, f := range []string{"git/git-commit.md", "review/blast-radius.md", "assets/domains.json", "SKILL.md"} {
		if _, err := os.Lstat(installedCore(o, f)); err != nil {
			t.Errorf("hive-core lacks %s: %v", f, err)
		}
	}
	absent(t, installedCore(o, "notes"))
	absent(t, installedCore(o, "notes/draft.md"))
	for path := range stateFor(t, o).Records {
		if strings.Contains(path, "draft.md") {
			t.Errorf("state records the undeclared %s", path)
		}
	}
	if again := apply(t, plan(t, "install", o)); again != "unchanged" {
		t.Fatalf("second update = %q", again)
	}
}

func TestDomainFolderIsNotInstalledWithoutAManifest(t *testing.T) {
	o := domainOptions(t)
	if err := os.Remove(packetSource(o, domainsManifestPath)); err != nil {
		t.Fatal(err)
	}
	apply(t, plan(t, "install", o))
	absent(t, installedCore(o, "git"))
	absent(t, installedCore(o, "review"))
}

func TestDomainLinksAreValidated(t *testing.T) {
	for name, text := range map[string]string{
		"missing file":   "# B\n\n[x](../git/nope.md)\n",
		"missing anchor": "# B\n\n[x](../git/git-commit.md#nope)\n",
		"undeclared":     "# B\n\n[x](../notes/draft.md)\n",
	} {
		t.Run(name, func(t *testing.T) {
			o := domainOptions(t)
			put(t, packetSource(o, "hive-core/review/blast-radius.md"), text)
			_, err := BuildPlan("install", o)
			if err == nil || !strings.Contains(err.Error(), "hive-core/review/blast-radius.md") {
				t.Fatalf("broken link accepted or unnamed: %v", err)
			}
			absent(t, o.StateDir)
		})
	}
}

func TestDomainManifestBadFormsAreRejectedBeforeAnyWrite(t *testing.T) {
	cases := []struct{ name, manifest, want string }{
		{"not json", `{"domains": [`, "domains.json"},
		{"not an object", `["git"]`, "want a JSON object"},
		{"other key", `{"domains": ["git"], "extra": 1}`, `"extra"`},
		{"only another key", `{"folders": ["git"]}`, `"folders"`},
		{"key twice", `{"domains": ["git"], "domains": ["review"]}`, "twice"},
		{"not an array", `{"domains": "git"}`, "array"},
		{"null", `{"domains": null}`, "array"},
		{"name not a string", `{"domains": [3]}`, "array"},
		{"repeated name", `{"domains": ["git", "git"]}`, `"git"`},
		{"reserved references", `{"domains": ["references"]}`, `"references"`},
		{"reserved scripts", `{"domains": ["scripts"]}`, `"scripts"`},
		{"reserved assets", `{"domains": ["assets"]}`, `"assets"`},
		{"reserved playbooks", `{"domains": ["playbooks"]}`, `"playbooks"`},
		{"reserved agents", `{"domains": ["agents"]}`, `"agents"`},
		{"not kebab case", `{"domains": ["Git"]}`, `"Git"`},
		{"escapes the skill", `{"domains": ["../x"]}`, `"../x"`},
		{"nested name", `{"domains": ["git/sub"]}`, `"git/sub"`},
		{"missing folder", `{"domains": ["git", "review", "ghost"]}`, `"ghost"`},
		{"text after the object", `{"domains": ["git", "review"]} x`, "after"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := domainOptions(t)
			put(t, packetSource(o, domainsManifestPath), tc.manifest)
			_, err := BuildPlan("install", o)
			if err == nil {
				t.Fatal("release accepted")
			}
			for _, want := range []string{"hive-core/assets/domains.json", tc.want} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error %q does not name %q", err, want)
				}
			}
			absent(t, o.StateDir)
			absent(t, filepath.Join(o.Home, ".agents"))
		})
	}
}

func TestDomainFolderThatHoldsNoFileIsRejected(t *testing.T) {
	o := domainOptions(t)
	put(t, packetSource(o, domainsManifestPath), `{"domains": ["git", "review", "empty"]}`)
	if err := os.MkdirAll(packetSource(o, "hive-core/empty"), 0o700); err != nil {
		t.Fatal(err)
	}
	_, err := BuildPlan("install", o)
	if err == nil || !strings.Contains(err.Error(), "hive-core/assets/domains.json") || !strings.Contains(err.Error(), `"empty"`) {
		t.Fatalf("empty domain accepted or unnamed: %v", err)
	}
}

func TestStoredReleaseWithADomainInstalls(t *testing.T) {
	o := domainOptions(t)
	p := plan(t, "install", o)
	rel := *p.Release
	fresh := setup(t)
	fresh.Hosts = o.Hosts
	stored := storeRelease(t, fresh, rel)
	apply(t, plan(t, "install", stored))
	if got := get(t, installedCore(fresh, "git/git-commit.md")); !strings.Contains(got, "Git commit") {
		t.Fatalf("domain file not installed from the stored release: %q", got)
	}
	absent(t, installedCore(fresh, "notes"))
}

func TestStoredReleaseIsCheckedAgainstItsOwnFiles(t *testing.T) {
	base := func(t *testing.T) (Options, Release) {
		o := domainOptions(t)
		p := plan(t, "install", o)
		return o, *p.Release
	}
	mutate := map[string]func(r *Release){
		"file in an unlisted folder": func(r *Release) {
			r.Files = append(r.Files, Payload{Path: "content/skills/hive-core/notes/draft.md", Data: []byte("# Draft\n"), Mode: 0o644})
		},
		"file in agents": func(r *Release) {
			r.Files = append(r.Files, Payload{Path: "content/skills/hive-core/agents/mine.yaml", Data: []byte("a: 1\n"), Mode: 0o644})
		},
		"listed folder missing": func(r *Release) {
			for i := range r.Files {
				if r.Files[i].Path == "content/skills/hive-core/assets/domains.json" {
					r.Files[i].Data = []byte(`{"domains": ["git", "review", "ghost"]}`)
				}
			}
		},
		"manifest malformed": func(r *Release) {
			for i := range r.Files {
				if r.Files[i].Path == "content/skills/hive-core/assets/domains.json" {
					r.Files[i].Data = []byte(`{"domains": ["references"]}`)
				}
			}
		},
		"folder without a manifest": func(r *Release) {
			var kept []Payload
			for _, f := range r.Files {
				if f.Path != "content/skills/hive-core/assets/domains.json" {
					kept = append(kept, f)
				}
			}
			r.Files = kept
		},
	}
	for name, mut := range mutate {
		t.Run(name, func(t *testing.T) {
			o, rel := base(t)
			rel.Files = append([]Payload{}, rel.Files...)
			if name == "folder without a manifest" {
				rel.Files = append(rel.Files, Payload{Path: "content/skills/hive-core/extra/e.md", Data: []byte("# E\n"), Mode: 0o644})
			}
			mut(&rel)
			stored := storeRelease(t, o, rel)
			_, err := BuildPlan("install", stored)
			if err == nil || !strings.Contains(err.Error(), "hive-core/assets/domains.json") {
				t.Fatalf("stored release accepted or manifest unnamed: %v", err)
			}
		})
	}
}

func TestRecordedStateWithADomainPathIsAccepted(t *testing.T) {
	o := domainOptions(t)
	apply(t, plan(t, "install", o))
	found := false
	for path, r := range stateFor(t, o).Records {
		if strings.HasSuffix(path, filepath.FromSlash("hive-core/git/git-commit.md")) && r.Target.Source == "content/skills/hive-core/git/git-commit.md" {
			found = true
		}
	}
	if !found {
		t.Fatal("state does not record the domain file")
	}
	for _, source := range []string{"content/skills/hive-core/git/git-commit.md", "content/skills/hive-core/principles/a/b.md", "content/skills/x/y-z/f.json"} {
		if !validSource(source) {
			t.Errorf("validSource rejects %s", source)
		}
	}
	for _, source := range []string{"content/skills/hive-core/Git/a.md", "content/skills/hive-core/../a.md", "content/skills/hive-core/git/.hidden.md", "content/skills/hive-core/git/a.exe", "content/skills/hive-core/a.md"} {
		if validSource(source) {
			t.Errorf("validSource accepts %s", source)
		}
	}
}

func TestUpdateRetiresReferencesAndPrincipleSkillsAndInstallsDomains(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude", "codex"}
	core := filepath.Join(o.Source, "content", "skills", "hive-core")
	put(t, filepath.Join(core, "SKILL.md"), "---\nname: hive-core\ndescription: Rule leaves.\ndisable-model-invocation: true\n---\nIndex.\n")
	put(t, filepath.Join(core, "references", "aa.md"), "# AA\n\nRead this.\n")
	principle := filepath.Join(o.Source, "content", "skills", "hive-principle-x")
	put(t, filepath.Join(principle, "SKILL.md"), leafFrontmatter("hive-principle-x"))
	apply(t, plan(t, "install", o))
	policy := filepath.Join(o.Home, ".agents", "skills", "hive-principle-x", "agents", "openai.yaml")
	if _, err := os.Lstat(policy); err != nil {
		t.Fatalf("prior install lacks the principle skill policy: %v", err)
	}
	oldLeafFile := installedCore(o, "references/aa.md")
	if _, err := os.Lstat(oldLeafFile); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(principle); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(core, "references")); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(core, "assets", "domains.json"), `{"domains": ["git"]}`)
	put(t, filepath.Join(core, "git", "aa.md"), "# AA\n\nRead this.\n")
	apply(t, plan(t, "install", o))

	absent(t, oldLeafFile)
	absent(t, installedCore(o, "references"))
	absent(t, filepath.Join(o.Home, ".agents", "skills", "hive-principle-x"))
	absent(t, policy)
	if got := get(t, installedCore(o, "git/aa.md")); !strings.Contains(got, "Read this.") {
		t.Fatalf("domain file not installed: %q", got)
	}
	for path := range stateFor(t, o).Records {
		if strings.Contains(path, "hive-principle-x") || strings.Contains(path, filepath.FromSlash("hive-core/references")) {
			t.Errorf("state still records %s", path)
		}
	}
	if again := apply(t, plan(t, "install", o)); again != "unchanged" {
		t.Fatalf("second update = %q", again)
	}
}
