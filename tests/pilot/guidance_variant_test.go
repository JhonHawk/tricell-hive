package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

func TestValidateGuidanceVariantFlags(t *testing.T) {
	for _, c := range []struct {
		name, host, source, arm string
		wantErr                 bool
	}{
		{"neither flag is the ordinary run", "codex", "", "", false},
		{"arm without source fails", "codex", "", "A", true},
		{"source without arm fails", "codex", "/src", "", true},
		{"bad arm value fails", "codex", "/src", "C", true},
		{"unsupported host fails", "claude", "/src", "A", true},
		{"codex arm A accepted", "codex", "/src", "A", false},
		{"grok arm B accepted", "grok", "/src", "B", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := validateGuidanceVariantFlags(c.host, c.source, c.arm)
			if (err != nil) != c.wantErr {
				t.Fatalf("host=%s source=%q arm=%q: err=%v wantErr=%v", c.host, c.source, c.arm, err, c.wantErr)
			}
		})
	}
}

func TestGuidanceReadPathsAreInsideShadowHomeForEachHost(t *testing.T) {
	shadow, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, host := range []string{"codex", "grok"} {
		agents, skill, err := guidanceReadPaths(host, shadow)
		if err != nil {
			t.Fatal(err)
		}
		if !under(agents, shadow) || !under(skill, shadow) {
			t.Fatalf("%s read paths escape the shadow home: %s %s", host, agents, skill)
		}
		if !strings.HasSuffix(skill, filepath.Join("skills", "flow-build", "SKILL.md")) {
			t.Fatalf("%s skill path is not flow-build's: %s", host, skill)
		}
	}
	if _, _, err := guidanceReadPaths("claude", shadow); err == nil {
		t.Fatal("expected an unsupported host to fail clearly")
	}
}

func newFixtureCheckout(t *testing.T, globalBody string) string {
	t.Helper()
	dir := t.TempDir()
	write := func(p, s string) {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(s), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write(management.GlobalSource, globalBody)
	write("content/skills/flow-build/SKILL.md", "---\nname: flow-build\ndescription: Implement authorized work.\n---\nPreserve evidence.\n")
	canonical, err := target.Canonical(dir)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}

func TestGuidanceVariantInstallsDifferentContentPerArmIntoDifferentShadowHomes(t *testing.T) {
	dirA := newFixtureCheckout(t, "# Rules A\nArm A body.\n")
	dirB := newFixtureCheckout(t, "# Rules B\nArm B body.\n")
	baseA, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	baseB, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	shadowA := filepath.Join(baseA, "shadow")
	shadowB := filepath.Join(baseB, "shadow")
	for _, home := range []string{shadowA, shadowB} {
		if err := os.MkdirAll(home, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := installGuidanceVariant(dirA, shadowA, "codex"); err != nil {
		t.Fatal(err)
	}
	if _, err := installGuidanceVariant(dirB, shadowB, "codex"); err != nil {
		t.Fatal(err)
	}
	agentsA, skillA, err := guidanceReadPaths("codex", shadowA)
	if err != nil {
		t.Fatal(err)
	}
	agentsB, skillB, err := guidanceReadPaths("codex", shadowB)
	if err != nil {
		t.Fatal(err)
	}
	if !under(agentsA, shadowA) || !under(agentsB, shadowB) {
		t.Fatal("installed guidance escaped its own shadow home")
	}
	hashA, hashB := hashOrAbsent(agentsA), hashOrAbsent(agentsB)
	if hashA == "absent" || hashB == "absent" {
		t.Fatalf("guidance was not installed: %s %s", hashA, hashB)
	}
	if hashA == hashB {
		t.Fatal("both arms installed identical guidance")
	}
	if hashOrAbsent(skillA) == "absent" || hashOrAbsent(skillB) == "absent" {
		t.Fatal("flow-build skill was not installed into the shadow home")
	}
}

func TestGuidanceVariantReportSerializesHashesIntoRunJSON(t *testing.T) {
	r := result{GuidanceVariant: &guidanceVariantReport{Source: "/src", Arm: "A", GuidanceBlockHash: "deadbeef", FlowBuildSkillHash: "cafef00d"}}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"GuidanceBlockHash":"deadbeef"`, `"FlowBuildSkillHash":"cafef00d"`, `"Arm":"A"`} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("missing %s in %s", want, b)
		}
	}
}

func TestPlanDestinationsUnderHomeGuardsEscape(t *testing.T) {
	home := filepath.Join(string(os.PathSeparator), "shadow", "home")
	good := management.Plan{Changes: []management.Change{{Target: target.Target{Path: filepath.Join(home, ".codex", "AGENTS.md")}}}}
	if err := planDestinationsUnderHome(good, home); err != nil {
		t.Fatal(err)
	}
	bad := management.Plan{Changes: []management.Change{{Target: target.Target{Path: filepath.Join(string(os.PathSeparator), "etc", "passwd")}}}}
	if err := planDestinationsUnderHome(bad, home); err == nil {
		t.Fatal("expected an escaping destination to be rejected")
	}
}

func TestScrubManagerEnvRemovesAndRestoresCodexAndGrokHome(t *testing.T) {
	t.Setenv("CODEX_HOME", "/real/codex")
	t.Setenv("GROK_HOME", "/real/grok")
	restore := scrubManagerEnv()
	if v := os.Getenv("CODEX_HOME"); v != "" {
		t.Fatalf("CODEX_HOME not scrubbed: %s", v)
	}
	if v := os.Getenv("GROK_HOME"); v != "" {
		t.Fatalf("GROK_HOME not scrubbed: %s", v)
	}
	restore()
	if v := os.Getenv("CODEX_HOME"); v != "/real/codex" {
		t.Fatalf("CODEX_HOME not restored: %s", v)
	}
	if v := os.Getenv("GROK_HOME"); v != "/real/grok" {
		t.Fatalf("GROK_HOME not restored: %s", v)
	}
}

func TestScrubManagerEnvLeavesUnsetVariablesUnset(t *testing.T) {
	t.Setenv("CODEX_HOME", "")
	os.Unsetenv("CODEX_HOME")
	t.Setenv("GROK_HOME", "")
	os.Unsetenv("GROK_HOME")
	restore := scrubManagerEnv()
	restore()
	if _, ok := os.LookupEnv("CODEX_HOME"); ok {
		t.Fatal("CODEX_HOME was invented by restore")
	}
	if _, ok := os.LookupEnv("GROK_HOME"); ok {
		t.Fatal("GROK_HOME was invented by restore")
	}
}

func TestGuidanceVariantEnvironmentPreservesEngramIsolationAndSetsHostHomes(t *testing.T) {
	for _, tc := range []struct{ host string }{{"codex"}, {"grok"}} {
		g := &guidanceVariant{shadowHome: "/shadow", host: tc.host, realGrokHome: "/real/grok"}
		in := []string{"HOME=/old-home", "CODEX_HOME=/old-codex", "GROK_HOME=/old-grok", "ENGRAM_DATA_DIR=/isolated/store", "PATH=/bin"}
		out := g.applyEnvironment(in)
		get := func(key string) (string, bool) {
			for _, e := range out {
				if k, v, ok := strings.Cut(e, "="); ok && k == key {
					return v, true
				}
			}
			return "", false
		}
		if v, ok := get("HOME"); !ok || v != "/shadow" {
			t.Fatalf("%s: HOME not shadowed: %v %v", tc.host, v, ok)
		}
		if v, ok := get("ENGRAM_DATA_DIR"); !ok || v != "/isolated/store" {
			t.Fatalf("%s: ENGRAM_DATA_DIR lost precedence: %v %v", tc.host, v, ok)
		}
		if v, ok := get("PATH"); !ok || v != "/bin" {
			t.Fatalf("%s: unrelated env entry lost", tc.host)
		}
		count := 0
		for _, e := range out {
			if strings.HasPrefix(e, "HOME=") {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("%s: duplicate HOME entries: %d", tc.host, count)
		}
		switch tc.host {
		case "codex":
			if v, ok := get("CODEX_HOME"); !ok || v != filepath.Join("/shadow", ".codex") {
				t.Fatalf("CODEX_HOME not shadowed: %v", v)
			}
			if _, ok := get("GROK_HOME"); ok {
				t.Fatal("a codex run must not carry GROK_HOME")
			}
		case "grok":
			if v, ok := get("GROK_HOME"); !ok || v != "/real/grok" {
				t.Fatalf("GROK_HOME not set to the real, unshadowed path: %v", v)
			}
			if _, ok := get("CODEX_HOME"); ok {
				t.Fatal("a grok run must not carry CODEX_HOME")
			}
		}
	}
}

func TestWriteShadowCodexConfigPreservesEngramServerDefinitionOnly(t *testing.T) {
	real := t.TempDir()
	realConfig := "model = \"gpt-6\"\n\n[mcp_servers.other]\ncommand = \"other-bin\"\n\n[mcp_servers.engram]\ncommand = \"npx\"\nargs = [\"-y\", \"@tricell/engram-mcp\"]\n\n[notice]\nflag = true\n"
	if err := os.WriteFile(filepath.Join(real, "config.toml"), []byte(realConfig), 0600); err != nil {
		t.Fatal(err)
	}
	shadow := t.TempDir()
	if err := writeShadowCodexConfig(shadow, real); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(shadow, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, want := range []string{"[mcp_servers.engram]", `command = "npx"`, `"-y"`, `"@tricell/engram-mcp"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in shadow config: %s", want, got)
		}
	}
	for _, unwanted := range []string{"other-bin", "mcp_servers.other", "notice", "gpt-6"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("shadow config copied unrelated real configuration (%q): %s", unwanted, got)
		}
	}
}

func TestWriteShadowCodexConfigFailsClearlyWithoutEngramServer(t *testing.T) {
	real := t.TempDir()
	if err := os.WriteFile(filepath.Join(real, "config.toml"), []byte("model = \"gpt-6\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeShadowCodexConfig(t.TempDir(), real); err == nil {
		t.Fatal("expected a missing Engram server definition to fail clearly")
	}
}

func TestLinkCodexAuthSymlinksRatherThanCopies(t *testing.T) {
	realHome := t.TempDir()
	if err := os.WriteFile(filepath.Join(realHome, "auth.json"), []byte(`{"token":"secret-value"}`), 0600); err != nil {
		t.Fatal(err)
	}
	shadowHome := t.TempDir()
	if err := linkCodexAuth(shadowHome, realHome); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(shadowHome, "auth.json")
	info, err := os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected a symlink, not a copy")
	}
	dest, err := os.Readlink(link)
	if err != nil || dest != filepath.Join(realHome, "auth.json") {
		t.Fatalf("wrong symlink target: %s", dest)
	}
}

func TestLinkCodexAuthFailsClearlyWhenRealAuthMissing(t *testing.T) {
	if err := linkCodexAuth(t.TempDir(), t.TempDir()); err == nil {
		t.Fatal("expected a clear error when the real auth.json is absent")
	}
}

func TestVerifyAndCleanupAuthSymlinkStillSymlinkIsRemovedWithoutWarning(t *testing.T) {
	real := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(real, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	shadow := filepath.Join(t.TempDir(), "auth.json")
	if err := os.Symlink(real, shadow); err != nil {
		t.Fatal(err)
	}
	stillSymlink, warning := verifyAndCleanupAuthSymlink(shadow)
	if !stillSymlink || warning != "" {
		t.Fatalf("unexpected result: stillSymlink=%v warning=%q", stillSymlink, warning)
	}
	if _, err := os.Lstat(shadow); !os.IsNotExist(err) {
		t.Fatal("symlink was not removed")
	}
}

func TestVerifyAndCleanupAuthSymlinkRegularFileIsDeletedWithWarningAndNeverRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(path, []byte(`{"token":"renewed-secret"}`), 0600); err != nil {
		t.Fatal(err)
	}
	stillSymlink, warning := verifyAndCleanupAuthSymlink(path)
	if stillSymlink || warning == "" {
		t.Fatalf("expected a replacement warning, got stillSymlink=%v warning=%q", stillSymlink, warning)
	}
	if strings.Contains(warning, "renewed-secret") {
		t.Fatal("warning must never carry the file's content")
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatal("the replaced auth file was not deleted")
	}
}

func TestVerifyAndCleanupAuthSymlinkAbsentIsQuiet(t *testing.T) {
	stillSymlink, warning := verifyAndCleanupAuthSymlink(filepath.Join(t.TempDir(), "auth.json"))
	if stillSymlink || warning != "" {
		t.Fatalf("an already-absent path must be quiet: stillSymlink=%v warning=%q", stillSymlink, warning)
	}
}
