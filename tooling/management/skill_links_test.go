package management

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/agents"
	"tricell-hive/integrations/target"
)

const linkedRole = "---\nname: link-role\ndescription: Link role\nmodel_profile: execution\naccess_profile: observe\n---\n" +
	"Read [browser automation](skill:flow-build/references/browser-automation.md).\n\n" +
	"```\n[example](skill:flow-build/references/example.md)\n```\n\n" +
	"[ref]: skill:flow-build/references/browser-automation.md\n"

const linkedGlobal = "# Rules\nSee [guide](skill:flow-build/references/browser-automation.md).\n\n" +
	"```\n[example](skill:flow-build/references/example.md)\n```\n"

const linkRoleSource = "content/agents/design/link-role.md"

// linkSetup builds a synthetic release whose role and global block carry
// skill: links, so the rewrite is observable on every host.
func linkSetup(t *testing.T, hosts ...string) Options {
	t.Helper()
	o := setup(t)
	o.Hosts = hosts
	profiles, err := os.ReadFile(syntheticProfiles)
	if err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(o.Source, agents.ProfilesSource), string(profiles))
	put(t, filepath.Join(o.Source, linkRoleSource), linkedRole)
	put(t, filepath.Join(o.Source, GlobalSource), linkedGlobal)
	put(t, filepath.Join(o.Source, "content/skills/flow-build/SKILL.md"), "---\nname: flow-build\ndescription: Build.\n---\nBuild.\n")
	put(t, filepath.Join(o.Source, "content/skills/flow-build/references/browser-automation.md"), "Automate.\n")
	return o
}

func roleText(t *testing.T, path string) string {
	t.Helper()
	text := get(t, path)
	if !strings.HasSuffix(path, ".toml") {
		return text
	}
	for _, line := range strings.Split(text, "\n") {
		if v, ok := strings.CutPrefix(line, "developer_instructions = "); ok {
			var s string
			if err := json.Unmarshal([]byte(v), &s); err != nil {
				t.Fatal(err)
			}
			return s
		}
	}
	t.Fatalf("no developer_instructions in %s", path)
	return ""
}

func TestUserScopeRolesAndBlocksLinkInstalledSkillPaths(t *testing.T) {
	o := linkSetup(t, "claude", "codex", "grok", "pi", "opencode", "cursor")
	p := plan(t, "install", o)
	apply(t, p)
	want := "<" + filepath.Join(o.Home, ".agents", "skills", "flow-build", "references", "browser-automation.md") + ">"
	for host, path := range agentTargets(t, o, linkRoleSource) {
		body := roleText(t, path)
		if !strings.Contains(body, "]("+want+")") {
			t.Errorf("%s role lacks %s:\n%s", host, want, body)
		}
		if strings.Count(body, "](skill:") != 1 || !strings.Contains(body, "[ref]: skill:flow-build/") {
			t.Errorf("%s role must keep only the fenced link and the reference definition as skill:\n%s", host, body)
		}
	}
	for _, block := range []string{
		filepath.Join(o.Home, ".codex", "AGENTS.md"),
		filepath.Join(o.Home, ".claude", "CLAUDE.md"),
		filepath.Join(o.Home, ".pi", "agent", "AGENTS.md"),
		filepath.Join(o.Home, ".config", "opencode", "AGENTS.md"),
		filepath.Join(o.Home, ".cursor", "AGENTS.md"),
	} {
		text := get(t, block)
		if !strings.Contains(text, "]("+want+")") || strings.Count(text, "](skill:") != 1 {
			t.Errorf("block %s not rewritten outside fences:\n%s", block, text)
		}
	}
}

func TestProjectScopeLinksAreRootRelative(t *testing.T) {
	for host, dir := range map[string]string{"claude": ".claude/skills", "codex": ".agents/skills"} {
		o := linkSetup(t, host)
		o.Scope = "project"
		o.Root = filepath.Join(o.Home, "repo")
		os.MkdirAll(o.Root, 0700)
		apply(t, plan(t, "install", o))
		want := "](<" + dir + "/flow-build/references/browser-automation.md>)"
		paths := agentTargets(t, o, linkRoleSource)
		body := roleText(t, paths[host])
		if !strings.Contains(body, want) || strings.Contains(body, o.Home) {
			t.Errorf("%s project role should link %s with no home path:\n%s", host, want, body)
		}
		blockPath := filepath.Join(o.Root, "CLAUDE.md")
		if host == "codex" {
			blockPath = filepath.Join(o.Root, "AGENTS.md")
		}
		block := get(t, blockPath)
		if !strings.Contains(block, want) || strings.Contains(block, o.Home) {
			t.Errorf("%s project block should link %s with no home path:\n%s", host, want, block)
		}
	}
}

func TestSecondPlanOnSameRevisionProposesNoWrites(t *testing.T) {
	o := linkSetup(t, "claude", "codex", "grok")
	apply(t, plan(t, "install", o))
	if got := apply(t, plan(t, "install", o)); got != "unchanged" {
		t.Fatalf("second install = %q, want unchanged", got)
	}
}

// Only the block is shared here: Claude and Grok both manage ~/.claude/CLAUDE.md,
// while Claude's roles live in ~/.claude/agents and Grok's in <GrokHome>/agents.
// The linked role is present so the plan also renders roles with the forced dirs.
func TestPlanFailsWhenSharedBlockSkillsDirDiffersByHostWithRolesPresent(t *testing.T) {
	defer func(f func(string, target.Config) string) { skillsDirFor = f }(skillsDirFor)
	skillsDirFor = func(host string, c target.Config) string {
		return filepath.Join(c.Home, "skills-"+host)
	}
	o := linkSetup(t, "claude", "grok")
	if _, err := BuildPlan("install", o); err == nil || !strings.Contains(err.Error(), "differs by host") {
		t.Fatalf("hosts that disagree on the skills directory must fail the plan, got %v", err)
	}
	entries, _ := os.ReadDir(o.Home)
	if len(entries) != 0 {
		t.Fatalf("a failed plan wrote files: %v", entries)
	}
}

func TestSharedBlockDifferingByHostIsRejectedBeforeAnyWrite(t *testing.T) {
	defer func(f func(string, target.Config) string) { skillsDirFor = f }(skillsDirFor)
	skillsDirFor = func(host string, c target.Config) string {
		return filepath.Join(c.Home, "skills-"+host)
	}
	o := linkSetup(t, "claude", "grok")
	os.Remove(filepath.Join(o.Source, linkRoleSource))
	if _, err := BuildPlan("install", o); err == nil || !strings.Contains(err.Error(), "shared block rendering differs by host") {
		t.Fatalf("block shared by Claude and Grok must fail on differing dirs, got %v", err)
	}
}

func TestOlderReleaseStillValidatesAfterRewrite(t *testing.T) {
	o := linkSetup(t, "claude")
	first := plan(t, "install", o)
	apply(t, first)
	put(t, filepath.Join(o.Source, GlobalSource), linkedGlobal+"Newer.\n")
	apply(t, plan(t, "install", o))
	old := o
	old.ReleaseID = first.Release.ID
	if _, err := BuildPlan("install", old); err != nil {
		t.Fatalf("an older release must still plan (Version unchanged): %v", err)
	}
}

// RequiredHosts previews a shared block with every known consumer. Without the
// install Config the preview resolves the skills directory from an empty home,
// computes different bytes, and reports an unrelated host as required.
func TestRequiredHostsDoesNotAddHostsForUnchangedLinkedSharedBlock(t *testing.T) {
	o := linkSetup(t, "claude", "grok")
	apply(t, plan(t, "install", o))
	only := o
	only.Hosts = []string{"claude"}
	hosts, err := RequiredHosts(only)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 1 || hosts[0] != "claude" {
		t.Fatalf("unchanged shared block must not require other hosts, got %v", hosts)
	}
}
