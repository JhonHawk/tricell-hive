package management

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"tricell-hive/integrations/mdlinks"
)

func antigravityHome(o Options) string { return filepath.Join(o.Home, ".gemini", "config") }
func antigravityLink(o Options, skill string) string {
	return filepath.Join(antigravityHome(o), "skills", skill)
}

func antigravitySetup(t *testing.T) Options {
	t.Helper()
	o := setup(t)
	o.Hosts = []string{"antigravity"}
	put(t, filepath.Join(o.Source, "content/skills/flow-build/SKILL.md"), "---\nname: flow-build\ndescription: Build.\n---\nBuild.\n")
	return o
}

func linkTarget(t *testing.T, path string) string {
	t.Helper()
	dest, err := os.Readlink(path)
	if err != nil {
		t.Fatal(err)
	}
	return dest
}

func TestAntigravityPlanProposesRelativeSkillLinksAndRulePartsNotGeminiMd(t *testing.T) {
	o := antigravitySetup(t)
	p := plan(t, "install", o)
	links := 0
	for _, ch := range p.Changes {
		if ch.Target.Kind == "block" && ch.Target.Path != antigravityRule(o, 1) {
			t.Fatalf("unexpected guidance block: %s", ch.Target.Path)
		}
		if ch.Target.Kind == "symlink" && strings.HasPrefix(ch.Target.Path, antigravityHome(o)+string(filepath.Separator)) {
			links++
		}
	}
	if links != 2 {
		t.Fatalf("want one link per skill, got %d", links)
	}
	apply(t, p)
	for _, skill := range []string{"workspace-conventions", "flow-build"} {
		link := antigravityLink(o, skill)
		if dest := linkTarget(t, link); dest != "../../../.agents/skills/"+skill {
			t.Fatalf("link target %q", dest)
		}
		resolved, err := filepath.EvalSymlinks(link)
		if err != nil {
			t.Fatal(err)
		}
		want, _ := filepath.EvalSymlinks(filepath.Join(o.Home, ".agents", "skills", skill))
		if resolved != want {
			t.Fatalf("%s resolves to %s, want %s", skill, resolved, want)
		}
	}
	absent(t, filepath.Join(o.Home, ".gemini", "GEMINI.md"))
}

func TestAntigravityLifecycleInstallUpdateRemoveWithSharedConsumer(t *testing.T) {
	o := antigravitySetup(t)
	o.Hosts = []string{"codex", "antigravity"}
	apply(t, plan(t, "install", o))
	if len(stateFor(t, o).Records[sharedPath(o)].Consumers) != 2 {
		t.Fatal("Antigravity did not register as a shared-skill consumer")
	}
	if apply(t, plan(t, "install", o)) != "unchanged" {
		t.Fatal("reinstall was not idempotent")
	}
	put(t, filepath.Join(o.Source, SkillSource), "---\nname: workspace-conventions\ndescription: Organize records.\n---\nUpdated.\n")
	apply(t, plan(t, "install", o))
	if !strings.Contains(get(t, sharedPath(o)), "Updated.") {
		t.Fatal("shared skill was not updated")
	}

	partial := o
	partial.Hosts = []string{"antigravity"}
	apply(t, plan(t, "remove", partial))
	absent(t, antigravityLink(o, "workspace-conventions"))
	absent(t, antigravityLink(o, "flow-build"))
	if len(stateFor(t, o).Records[sharedPath(o)].Consumers) != 1 {
		t.Fatal("removing Antigravity dropped Codex's shared-skill consumer")
	}
	if !strings.Contains(get(t, filepath.Join(o.Home, ".codex", "AGENTS.md")), Begin) {
		t.Fatal("Codex block was removed")
	}
}

func TestAntigravityForeignSkillLinkKeepsItsTarget(t *testing.T) {
	o := antigravitySetup(t)
	skills := filepath.Join(antigravityHome(o), "skills")
	if err := os.MkdirAll(skills, 0700); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(skills, "firecrawl")
	const dest = "../../../.agents/skills/firecrawl"
	if err := os.Symlink(dest, foreign); err != nil {
		t.Fatal(err)
	}
	apply(t, plan(t, "install", o))
	if got := linkTarget(t, foreign); got != dest {
		t.Fatalf("install changed the foreign link to %q", got)
	}
	apply(t, plan(t, "install", o))
	apply(t, plan(t, "remove", o))
	if got := linkTarget(t, foreign); got != dest {
		t.Fatalf("remove changed the foreign link to %q", got)
	}
	absent(t, antigravityLink(o, "workspace-conventions"))
}

func TestAntigravityRejectsProjectScope(t *testing.T) {
	o := antigravitySetup(t)
	o.Scope = "project"
	o.Root = filepath.Join(o.Home, "project")
	os.MkdirAll(o.Root, 0700)
	if _, err := BuildPlan("install", o); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("project scope should fail with the unsupported-scope error, got %v", err)
	}
	absent(t, antigravityHome(o))
}

func TestAntigravityStatusInstalledRetainedAndNotInstalled(t *testing.T) {
	o := antigravitySetup(t)
	o.Hosts = []string{"codex", "antigravity"}
	apply(t, plan(t, "install", o))
	only := o
	only.Hosts = []string{"antigravity"}
	statusOf := func() map[string]string {
		entries, err := Status(only)
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]string{}
		for _, en := range entries {
			m[en.Path] = en.Status
		}
		return m
	}
	m := statusOf()
	if m[antigravityLink(o, "workspace-conventions")] != "installed" || m[sharedPath(o)] != "installed" {
		t.Fatalf("status after install: %+v", m)
	}
	apply(t, plan(t, "remove", only))
	m = statusOf()
	if m[antigravityLink(o, "workspace-conventions")] != "not_installed" || m[sharedPath(o)] != "retained_shared" {
		t.Fatalf("status after removal: %+v", m)
	}
}

func TestAntigravityHomeAbsentIsCreatedAndCleanedUp(t *testing.T) {
	o := antigravitySetup(t)
	absent(t, antigravityHome(o))
	apply(t, plan(t, "install", o))
	if info, err := os.Stat(antigravityHome(o)); err != nil || !info.IsDir() {
		t.Fatal("Antigravity home was not created")
	}
	apply(t, plan(t, "remove", o))
	absent(t, antigravityLink(o, "workspace-conventions"))
}

func TestAntigravityRecoveryAtWriteAndStateBoundaries(t *testing.T) {
	for _, stage := range []string{"prepared", "write:0", "write:1", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := antigravitySetup(t)
			p := plan(t, "install", o)
			e := Engine{failpoint: func(s string) error {
				if s == stage {
					return errors.New("injected")
				}
				return nil
			}}
			if _, err := e.Apply(p); err == nil {
				t.Fatal("did not fail")
			}
			if _, err := BuildPlan("install", o); err == nil {
				t.Fatal("pending operation ignored")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			absent(t, sharedPath(o))
			absent(t, antigravityLink(o, "workspace-conventions"))
			apply(t, plan(t, "install", o))
			if _, err := os.Lstat(antigravityLink(o, "workspace-conventions")); err != nil {
				t.Fatal("install after recovery did not write the link")
			}
		})
	}
}

func TestPlanSavedBeforeAntigravitySupportAsksToRegenerate(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	p.Config.AntigravityHome = ""
	p.ID = planID(p)
	if err := validatePlan(p, emptyState()); err == nil || !strings.Contains(err.Error(), "regenerate") {
		t.Fatalf("legacy plan accepted or wrong error: %v", err)
	}
}

// --- T2: the global guidance as numbered always-on rule parts ------------------

const ruleFrontmatterPrefix = "---\ntrigger: always_on\n---\n"

func antigravityRule(o Options, n int) string {
	return filepath.Join(antigravityHome(o), "rules", fmt.Sprintf("hive-%02d.md", n))
}

// ruleGuidance builds a synthetic guidance of sections ## Section 01.. whose
// bodies are about size bytes each, so tests control the part count.
func ruleGuidance(sections, size int) string {
	var b strings.Builder
	b.WriteString("# Synthetic guidance\n\nPreamble line.\n\n")
	for i := 1; i <= sections; i++ {
		fmt.Fprintf(&b, "## Section %02d\n\n", i)
		line := fmt.Sprintf("Guidance sentence of section %02d padded to sixty bytes....\n", i)
		for n := 0; n < size/len(line); n++ {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// ruleFiles maps each hive-NN.md file of the rules folder to its bytes.
func ruleFiles(t *testing.T, o Options) map[string]string {
	t.Helper()
	out := map[string]string{}
	entries, err := os.ReadDir(filepath.Join(antigravityHome(o), "rules"))
	if os.IsNotExist(err) {
		return out
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		out[e.Name()] = get(t, filepath.Join(antigravityHome(o), "rules", e.Name()))
	}
	return out
}

func hiveRuleNames(files map[string]string) []string {
	var names []string
	for name := range files {
		if strings.HasPrefix(name, "hive-") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// ruleBody returns the text between the Hive block markers of a rule file.
func ruleBody(t *testing.T, file string) string {
	t.Helper()
	a := strings.Index(file, Begin+"\n")
	b := strings.Index(file, "\n"+End)
	if a < 0 || b < a {
		t.Fatalf("rule file has no Hive block:\n%.200s", file)
	}
	return file[a+len(Begin)+1 : b]
}

func checkRuleParts(t *testing.T, files map[string]string, rewritten string) []string {
	t.Helper()
	names := hiveRuleNames(files)
	if len(names) == 0 {
		t.Fatal("no rule parts written")
	}
	var bodies []string
	for i, name := range names {
		if want := fmt.Sprintf("hive-%02d.md", i+1); name != want {
			t.Fatalf("part %d is %s, want %s", i+1, name, want)
		}
		f := files[name]
		if !strings.HasPrefix(f, ruleFrontmatterPrefix) {
			t.Fatalf("%s does not start with the always_on frontmatter", name)
		}
		if len(f) > 24000 {
			t.Fatalf("%s is %d bytes, over 24000", name, len(f))
		}
		bodies = append(bodies, ruleBody(t, f))
	}
	if joined := strings.Join(bodies, "\n\n"); joined != strings.TrimRight(rewritten, "\n") {
		t.Fatalf("joined managed texts differ from the rewritten guidance (%d vs %d bytes)", len(joined), len(strings.TrimRight(rewritten, "\n")))
	}
	return names
}

func rewrittenGuidance(o Options, source string) string {
	c, _, _ := normalize(o)
	return mdlinks.RewriteSkillLinks(source, skillsDirFor("antigravity", c))
}

func TestAntigravityRulePartsFromRealGuidanceFitAndReproduceIt(t *testing.T) {
	o := antigravitySetup(t)
	real, err := os.ReadFile("../../content/guidance/global.md")
	if err != nil {
		t.Fatal(err)
	}
	// The guidance links to the skills it names; the release needs them present.
	if err := os.RemoveAll(filepath.Join(o.Source, "content", "skills")); err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(filepath.Join(o.Source, "content", "skills"), os.DirFS("../../content/skills")); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(o.Source, GlobalSource), string(real))
	apply(t, plan(t, "install", o))
	files := ruleFiles(t, o)
	names := checkRuleParts(t, files, rewrittenGuidance(o, string(real)))
	for _, n := range names {
		t.Logf("%s: %d bytes", n, len(files[n]))
	}
	absent(t, filepath.Join(o.Home, ".gemini", "GEMINI.md"))
	if strings.Contains(files[names[0]], "\n## Hive\n") && !strings.Contains(string(real), "```markdown\n## Hive") {
		t.Fatal("unexpected heading")
	}
	for _, u := range splitGuidanceUnits(string(real)) {
		if u.heading == "## Hive" {
			t.Fatal("a ## inside a fenced code block was treated as a heading")
		}
	}
	// Status reports every part as installed.
	entries, err := Status(o)
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, en := range entries {
		if en.Kind == "block" && en.Status == "installed" {
			seen++
		}
	}
	if seen != len(names) {
		t.Fatalf("status reports %d installed parts, want %d", seen, len(names))
	}
}

func TestAntigravityOtherHostsKeepOneBlock(t *testing.T) {
	o := antigravitySetup(t)
	o.Hosts = []string{"codex", "claude", "antigravity"}
	put(t, filepath.Join(o.Source, GlobalSource), ruleGuidance(6, 5000))
	apply(t, plan(t, "install", o))
	want := ruleGuidance(6, 5000)
	for _, f := range []string{filepath.Join(o.Home, ".codex", "AGENTS.md"), filepath.Join(o.Home, ".claude", "CLAUDE.md")} {
		if body := ruleBody(t, get(t, f)); body != strings.TrimRight(want, "\n") {
			t.Fatalf("%s does not hold the whole guidance", f)
		}
	}
	if len(hiveRuleNames(ruleFiles(t, o))) < 2 {
		t.Fatal("Antigravity did not get several parts")
	}
}

func TestAntigravityFencedHeadingDoesNotCut(t *testing.T) {
	guidance := "# G\n\n## One\n\ntext\n\n```markdown\n## Hive\n\n- Project: x\n```\n\n~~~\n## Tilde\n~~~\n\n## Two\n\ntext\n"
	var headings []string
	for _, u := range splitGuidanceUnits(guidance) {
		headings = append(headings, u.heading)
	}
	want := []string{"the text before the first ## heading", "## One", "## Two"}
	if strings.Join(headings, "|") != strings.Join(want, "|") {
		t.Fatalf("units %q, want %q", headings, want)
	}
	// Force the cut to fall right at the fenced heading: no part may start inside the fence.
	big := "# G\n\n## One\n\n" + strings.Repeat("a line of text for the first section ........\n", 330) + "\n```markdown\n## Hive\n```\n\n## Two\n\ntext\n"
	parts, err := splitAntigravityGuidance(big)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range parts {
		if strings.HasPrefix(p, "## Hive") {
			t.Fatal("a part starts at the fenced heading")
		}
	}
	if strings.Join(parts, "") != big {
		t.Fatal("parts do not concatenate to the guidance")
	}
}

func TestAntigravityOversizedSectionFailsThePlanNamingIt(t *testing.T) {
	o := antigravitySetup(t)
	put(t, filepath.Join(o.Source, GlobalSource), ruleGuidance(2, 2000)+"## The huge section\n\n"+strings.Repeat("x", 23000)+"\n")
	_, err := BuildPlan("install", o)
	if err == nil || !strings.Contains(err.Error(), "The huge section") {
		t.Fatalf("want an error naming the section, got %v", err)
	}
	absent(t, antigravityHome(o))
}

func TestAntigravityPartCountFollowsGuidanceAndUpdateRemovesOnlyHiveSurplus(t *testing.T) {
	o := antigravitySetup(t)
	foreign := filepath.Join(antigravityHome(o), "rules", "other.md")
	put(t, foreign, "---\ntrigger: always_on\n---\nMy own rule.\n")
	const foreignBytes = "---\ntrigger: always_on\n---\nMy own rule.\n"
	checkForeign := func(stage string) {
		t.Helper()
		if got := get(t, foreign); got != foreignBytes {
			t.Fatalf("%s changed the foreign rule: %q", stage, got)
		}
	}
	install := func(guidance string) map[string]string {
		put(t, filepath.Join(o.Source, GlobalSource), guidance)
		apply(t, plan(t, "install", o))
		files := ruleFiles(t, o)
		checkRuleParts(t, files, rewrittenGuidance(o, guidance))
		checkForeign("install")
		return files
	}
	small := ruleGuidance(3, 5000)
	grown := ruleGuidance(14, 5000)
	one := install(small)
	if n := len(hiveRuleNames(one)); n != 1 {
		t.Fatalf("small guidance gave %d parts", n)
	}
	many := install(grown)
	nMany := len(hiveRuleNames(many))
	if nMany < 3 {
		t.Fatalf("grown guidance gave %d parts", nMany)
	}
	if apply(t, plan(t, "install", o)) != "unchanged" {
		t.Fatal("update with the same guidance was not idempotent")
	}
	shrunk := install(ruleGuidance(7, 5000))
	names := hiveRuleNames(shrunk)
	if len(names) >= nMany {
		t.Fatalf("shrunk guidance still gives %d parts", len(names))
	}
	for i := len(names) + 1; i <= nMany; i++ {
		absent(t, antigravityRule(o, i))
	}
	for name := range stateRuleParts(t, o) {
		if _, ok := shrunk[name]; !ok {
			t.Fatalf("state records %s but the file is gone", name)
		}
	}
	apply(t, plan(t, "remove", o))
	if left := hiveRuleNames(ruleFiles(t, o)); len(left) != 0 {
		t.Fatalf("remove left %v", left)
	}
	checkForeign("remove")
}

func TestAntigravitySurplusPartWithUserTextKeepsTheText(t *testing.T) {
	o := antigravitySetup(t)
	put(t, filepath.Join(o.Source, GlobalSource), ruleGuidance(14, 5000))
	apply(t, plan(t, "install", o))
	n := len(hiveRuleNames(ruleFiles(t, o)))
	last := antigravityRule(o, n)
	put(t, last, get(t, last)+"user note\n")
	put(t, filepath.Join(o.Source, GlobalSource), ruleGuidance(3, 5000))
	apply(t, plan(t, "install", o))
	if got := get(t, last); got != "user note\n" {
		t.Fatalf("surplus part with user text became %q", got)
	}
}

func stateRuleParts(t *testing.T, o Options) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	for path := range stateFor(t, o).Records {
		if strings.HasPrefix(filepath.Base(path), "hive-") {
			out[filepath.Base(path)] = true
		}
	}
	return out
}

func TestAntigravityGuidanceShrinksFromManyPartsToOne(t *testing.T) {
	o := antigravitySetup(t)
	put(t, filepath.Join(o.Source, GlobalSource), ruleGuidance(14, 5000))
	apply(t, plan(t, "install", o))
	if len(hiveRuleNames(ruleFiles(t, o))) < 3 {
		t.Fatal("setup did not produce several parts")
	}
	small := ruleGuidance(2, 3000)
	put(t, filepath.Join(o.Source, GlobalSource), small)
	apply(t, plan(t, "install", o))
	files := ruleFiles(t, o)
	if names := checkRuleParts(t, files, rewrittenGuidance(o, small)); len(names) != 1 {
		t.Fatalf("want one part, got %v", names)
	}
	if got := stateRuleParts(t, o); len(got) != 1 {
		t.Fatalf("state keeps %v", got)
	}
}

func TestAntigravityVoiceLivesOnlyInPartOne(t *testing.T) {
	o := antigravitySetup(t)
	if err := os.CopyFS(filepath.Join(o.Source, "content", "voices"), os.DirFS("../../content/voices")); err != nil {
		t.Fatal(err)
	}
	guidance := ruleGuidance(14, 3000)
	put(t, filepath.Join(o.Source, GlobalSource), guidance)
	foreign := filepath.Join(antigravityHome(o), "rules", "other.md")
	put(t, foreign, "mine\n")
	apply(t, plan(t, "install", o))
	before := ruleFiles(t, o)
	names := hiveRuleNames(before)
	if len(names) < 2 {
		t.Fatalf("want several parts, got %v", names)
	}
	longest := VoiceSetting{ID: "mentor", Address: "name", Name: strings.Repeat("𝒜", 40), Intensity: "marked"}
	apply(t, voicePlan(t, "set", o, longest))
	after := ruleFiles(t, o)
	for i, name := range names {
		changed := before[name] != after[name]
		if (i == 0) != changed {
			t.Fatalf("%s changed=%v after voice set", name, changed)
		}
		if i > 0 && strings.Contains(after[name], VoiceBegin) {
			t.Fatalf("%s carries a voice block", name)
		}
		if len(after[name]) > 24000 {
			t.Fatalf("%s is %d bytes", name, len(after[name]))
		}
	}
	if !strings.Contains(after[names[0]], VoiceBegin) {
		t.Fatal("part 1 has no voice block")
	}
	// Installing again with the voice active keeps the same bytes.
	if apply(t, plan(t, "install", o)) != "unchanged" {
		t.Fatal("update with the voice active changed files")
	}
	// A grown guidance adds parts without a second voice block.
	put(t, filepath.Join(o.Source, GlobalSource), ruleGuidance(20, 3000))
	apply(t, plan(t, "install", o))
	grown := ruleFiles(t, o)
	voices := 0
	for _, name := range hiveRuleNames(grown) {
		voices += strings.Count(grown[name], VoiceBegin)
		if len(grown[name]) > 24000 {
			t.Fatalf("%s is %d bytes", name, len(grown[name]))
		}
	}
	if voices != 1 || !strings.Contains(grown["hive-01.md"], VoiceBegin) {
		t.Fatalf("want one voice block in part 1, got %d", voices)
	}
	apply(t, voicePlan(t, "off", o, VoiceSetting{}))
	off := ruleFiles(t, o)
	if strings.Contains(off["hive-01.md"], VoiceBegin) {
		t.Fatal("voice off left the block")
	}
	if get(t, foreign) != "mine\n" {
		t.Fatal("foreign rule changed")
	}
}

func TestAntigravityVoiceReserveCoversTheShippedVoices(t *testing.T) {
	voices, err := ListVoices("../..")
	if err != nil || len(voices) == 0 {
		t.Fatalf("voices: %v %v", voices, err)
	}
	for _, v := range voices {
		body, _, err := RenderVoice("../..", VoiceSetting{ID: v.ID, Address: "name", Name: strings.Repeat("𝒜", 40), Intensity: "marked"})
		if err != nil {
			t.Fatal(err)
		}
		if size := len(managedBlock(body, nil, voiceMarkers)); size > antigravityVoiceReserve {
			t.Fatalf("voice %s renders %d bytes, over the %d-byte reserve", v.ID, size, antigravityVoiceReserve)
		}
	}
}

func TestAntigravityUnownedRuleFileStopsInstall(t *testing.T) {
	o := antigravitySetup(t)
	mine := antigravityRule(o, 1)
	put(t, mine, "my own rule\n")
	if _, err := BuildPlan("install", o); err == nil || !strings.Contains(err.Error(), "unowned Antigravity rule file") {
		t.Fatalf("want unowned rule error, got %v", err)
	}
	if get(t, mine) != "my own rule\n" {
		t.Fatal("foreign hive-01.md changed")
	}
}

func TestAntigravityRulePartsRecoverAtWriteBoundaries(t *testing.T) {
	for _, stage := range []string{"write:0", "write:1", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := antigravitySetup(t)
			put(t, filepath.Join(o.Source, GlobalSource), ruleGuidance(14, 3000))
			p := plan(t, "install", o)
			e := Engine{failpoint: func(s string) error {
				if s == stage {
					return errors.New("injected")
				}
				return nil
			}}
			if _, err := e.Apply(p); err == nil {
				t.Fatal("did not fail")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if names := hiveRuleNames(ruleFiles(t, o)); len(names) != 0 {
				t.Fatalf("recovery left %v", names)
			}
			apply(t, plan(t, "install", o))
			checkRuleParts(t, ruleFiles(t, o), rewrittenGuidance(o, ruleGuidance(14, 3000)))
		})
	}
}

func TestAntigravityRolesInstallOnePerRoleAndRemoveDeletesThem(t *testing.T) {
	o := antigravitySetup(t)
	modelsSource(t, o)
	roles := []string{plainRole, "deep-role", inheritRole}
	apply(t, plan(t, "install", o))
	for _, role := range roles {
		path := filepath.Join(antigravityHome(o), "agents", role+".md")
		text := get(t, path)
		for _, want := range []string{"name: \"" + role + "\"\n", "model: \"inherit\"\n", "Use evidence.\n"} {
			if !strings.Contains(text, want) {
				t.Fatalf("%s lacks %q:\n%s", path, want, text)
			}
		}
		for _, banned := range []string{"tools", "readonly", "effort", "thinking"} {
			if strings.Contains(text, banned) {
				t.Fatalf("%s must not mention %q:\n%s", path, banned, text)
			}
		}
	}
	entries, err := os.ReadDir(filepath.Join(antigravityHome(o), "agents"))
	if err != nil || len(entries) != len(roles) {
		t.Fatalf("agents folder = %d entries, %v; want %d", len(entries), err, len(roles))
	}
	// A foreign agent file is not Hive's and survives remove.
	put(t, filepath.Join(antigravityHome(o), "agents", "mine.md"), "mine\n")
	apply(t, plan(t, "remove", o))
	for _, role := range roles {
		absent(t, filepath.Join(antigravityHome(o), "agents", role+".md"))
	}
	if get(t, filepath.Join(antigravityHome(o), "agents", "mine.md")) != "mine\n" {
		t.Fatal("remove changed a foreign agent file")
	}
}

func TestAntigravityModelOverrideAcceptsATierAndRefusesAnEffort(t *testing.T) {
	o := antigravitySetup(t)
	modelsSource(t, o)
	apply(t, plan(t, "install", o))
	path := filepath.Join(antigravityHome(o), "agents", plainRole+".md")
	apply(t, buildModels(t, o, "antigravity", map[string]ModelOverride{plainRole: {Model: "flash_lite"}}))
	if got := installedField(get(t, path), "model"); got != "flash_lite" {
		t.Fatalf("model = %q, want flash_lite", got)
	}
	_, err := BuildModelsPlan(o, "antigravity", map[string]ModelOverride{plainRole: {Model: "pro", Effort: "high"}})
	if err == nil || !strings.Contains(err.Error(), "antigravity does not support effort") {
		t.Fatalf("BuildModelsPlan error = %v, want the no-effort reason", err)
	}
	apply(t, buildModels(t, o, "antigravity", nil))
	if got := installedField(get(t, path), "model"); got != "inherit" {
		t.Fatalf("model after reset = %q, want inherit", got)
	}
}
