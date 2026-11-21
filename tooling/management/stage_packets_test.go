package management

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const (
	packetQuestions = "# Questions\n\n" +
		"Read this. Sibling [commit](git-commit.md), anchored [when](git-commit.md#when), own [self](#self-heading),\n" +
		"nested [more](questions/unanswered.md), flow [entry](../../flow-build/SKILL.md), web [site](https://example.com/a).\n\n" +
		"```md\n[fenced](git-commit.md)\n```\n\n" +
		"[def]: git-commit.md\n[nested-def]: questions/unanswered.md\n\n" +
		"## Self heading\n\nBody.\n"
	packetGitCommit = "---\nname: ignored\n---\n# Git commit\n\nCommit rules.\n\n## When\n\nBefore a commit.\n"
	packetUnanswer  = "# Unanswered\n\nWhen a card returns unanswered.\n"
	packetChange    = "# Change folder\n\n## Change folder\n\nOwn section.\n"
	packetFlowSkill = "---\nname: %s\ndescription: A flow.\n---\nFlow body.\n"
)

const wantBuildPacket = "## hive-core/references/questions\n\n" +
	"# Questions\n\n" +
	"Read this. Sibling [commit](#hive-corereferencesgit-commit), anchored [when](../hive-core/references/git-commit.md#when), own [self](../hive-core/references/questions.md#self-heading),\n" +
	"nested [more](../hive-core/references/questions/unanswered.md), flow [entry](SKILL.md), web [site](https://example.com/a).\n\n" +
	"```md\n[fenced](git-commit.md)\n```\n\n" +
	"[def]: #hive-corereferencesgit-commit\n[nested-def]: ../hive-core/references/questions/unanswered.md\n\n" +
	"## Self heading\n\nBody.\n\n" +
	"## hive-core/references/git-commit\n\n" +
	"# Git commit\n\nCommit rules.\n\n## When\n\nBefore a commit.\n\n" +
	"<!-- hive packet end: stage · hive-core/references/questions,hive-core/references/git-commit -->\n"

func packetSource(o Options, rel string) string {
	return filepath.Join(o.Source, "content", "skills", filepath.FromSlash(rel))
}

func packetManifest(t *testing.T, o Options, flow, json string) {
	t.Helper()
	put(t, packetSource(o, flow+"/assets/stage.json"), json)
}

// packetOptions builds a source that holds hive-core with four leaves, a
// flow-build with one packet and a flow-verify with two.
func packetOptions(t *testing.T) Options {
	t.Helper()
	o := setup(t)
	o.Hosts = []string{"antigravity", "claude", "codex"}
	put(t, packetSource(o, "hive-core/SKILL.md"), "---\nname: hive-core\ndescription: Leaves.\ndisable-model-invocation: true\n---\nIndex.\n")
	put(t, packetSource(o, "hive-core/references/questions.md"), packetQuestions)
	put(t, packetSource(o, "hive-core/references/questions/unanswered.md"), packetUnanswer)
	put(t, packetSource(o, "hive-core/references/git-commit.md"), packetGitCommit)
	put(t, packetSource(o, "hive-core/references/change-folder.md"), packetChange)
	for _, flow := range []string{"flow-build", "flow-verify"} {
		put(t, packetSource(o, flow+"/SKILL.md"), fmt.Sprintf(packetFlowSkill, flow))
	}
	packetManifest(t, o, "flow-build", `{"stage": ["hive-core/references/questions.md", "hive-core/references/git-commit.md"]}`)
	packetManifest(t, o, "flow-verify", `{"stage-entry": ["hive-core/references/git-commit.md"], "stage-exit": ["hive-core/references/change-folder.md", "hive-core/references/questions.md"]}`)
	return o
}

func packetPath(o Options, flow, key string) string {
	return filepath.Join(o.Home, ".agents", "skills", flow, key+".md")
}

func packetRecords(t *testing.T, o Options) []string {
	t.Helper()
	var out []string
	for path := range stateFor(t, o).Records {
		if stagePacketFile.MatchString(filepath.Base(path)) {
			out = append(out, path)
		}
	}
	sort.Strings(out)
	return out
}

func packetChangedPaths(p Plan) []string {
	var out []string
	for _, ch := range p.Changes {
		if ch.Before == nil || ch.After == nil || !bytes.Equal(ch.Before.Managed, ch.After.Managed) {
			out = append(out, ch.Target.Path)
		}
	}
	sort.Strings(out)
	return out
}

func TestStagePacketOneIsGeneratedForEveryHostWithItsLinksRewritten(t *testing.T) {
	o := packetOptions(t)
	p := plan(t, "install", o)
	apply(t, p)

	if got := get(t, packetPath(o, "flow-build", "stage")); got != wantBuildPacket {
		t.Fatalf("packet = %q\nwant %q", got, wantBuildPacket)
	}
	r, ok := stateFor(t, o).Records[packetPath(o, "flow-build", "stage")]
	if !ok {
		t.Fatal("packet is not a recorded managed target")
	}
	if r.Target.Kind != "skill" || r.Target.Source != "content/skills/flow-build/stage.md" || !r.CreatedFile {
		t.Fatalf("record = %+v", r)
	}
	hosts := map[string]bool{}
	for _, c := range r.Consumers {
		hosts[c.Host] = true
	}
	for _, h := range o.Hosts {
		if !hosts[h] {
			t.Errorf("host %s does not consume the packet: %+v", h, r.Consumers)
		}
	}
	// Every host sees the packet through its skill folder: the shared store is
	// linked to Claude's and Antigravity's catalogues.
	for _, link := range []string{filepath.Join(o.Home, ".claude", "skills", "flow-build", "stage.md"), filepath.Join(o.Home, ".gemini", "config", "skills", "flow-build", "stage.md")} {
		if got := get(t, link); got != wantBuildPacket {
			t.Errorf("%s does not reach the packet", link)
		}
	}
	// SKILL.md and the manifest stay as released.
	if got := get(t, filepath.Join(o.Home, ".agents", "skills", "flow-build", "SKILL.md")); got != fmt.Sprintf(packetFlowSkill, "flow-build") {
		t.Fatalf("SKILL.md changed: %q", got)
	}
	if again := apply(t, plan(t, "install", o)); again != "unchanged" {
		t.Fatalf("second update = %q", again)
	}
}

func TestStagePacketTwoKeysGiveTwoPackets(t *testing.T) {
	o := packetOptions(t)
	apply(t, plan(t, "install", o))
	entry := get(t, packetPath(o, "flow-verify", "stage-entry"))
	exit := get(t, packetPath(o, "flow-verify", "stage-exit"))
	if !strings.HasPrefix(entry, "## hive-core/references/git-commit\n") || !strings.HasSuffix(entry, "<!-- hive packet end: stage-entry · hive-core/references/git-commit -->\n") {
		t.Fatalf("stage-entry = %q", entry)
	}
	if !strings.HasPrefix(exit, "## hive-core/references/change-folder\n") || strings.Index(exit, "## hive-core/references/questions\n") < 0 ||
		!strings.HasSuffix(exit, "<!-- hive packet end: stage-exit · hive-core/references/change-folder,hive-core/references/questions -->\n") {
		t.Fatalf("stage-exit = %q", exit)
	}
	if got := len(packetRecords(t, o)); got != 3 {
		t.Fatalf("%d packet records, want 3", got)
	}
}

func TestStagePacketRewritesWhenAMemberChangesAndOnlyThen(t *testing.T) {
	o := packetOptions(t)
	apply(t, plan(t, "install", o))
	put(t, packetSource(o, "hive-core/references/git-commit.md"), "# Git commit\n\nNew commit rules.\n\n## When\n\nBefore a commit.\n")
	p := plan(t, "install", o)
	want := []string{
		filepath.Join(o.Home, ".agents", "skills", "flow-build", "stage.md"),
		filepath.Join(o.Home, ".agents", "skills", "flow-verify", "stage-entry.md"),
		filepath.Join(o.Home, ".agents", "skills", "hive-core", "references", "git-commit.md"),
	}
	sort.Strings(want)
	if got := packetChangedPaths(p); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("changed = %v, want %v", got, want)
	}
	apply(t, p)
	if !strings.Contains(get(t, packetPath(o, "flow-build", "stage")), "New commit rules.") {
		t.Fatal("packet not rewritten")
	}
	if strings.Contains(get(t, packetPath(o, "flow-verify", "stage-exit")), "New commit rules.") {
		t.Fatal("a packet without the member changed")
	}
}

func TestStagePacketDeletedIsReinstalledAndEditedIsAnOwnershipError(t *testing.T) {
	o := packetOptions(t)
	apply(t, plan(t, "install", o))
	if err := os.Remove(packetPath(o, "flow-build", "stage")); err != nil {
		t.Fatal(err)
	}
	apply(t, plan(t, "install", o))
	if get(t, packetPath(o, "flow-build", "stage")) != wantBuildPacket {
		t.Fatal("deleted packet not reinstalled")
	}
	put(t, packetPath(o, "flow-build", "stage"), "edited\n")
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("edited packet overwritten without an ownership error")
	}
}

func TestStagePacketLeavesWhenItsKeyItsManifestOrItsFlowLeaves(t *testing.T) {
	steps := []struct {
		name   string
		change func(o Options)
		gone   []string
		kept   []string
	}{
		{"key", func(o Options) {
			packetManifest(t, o, "flow-verify", `{"stage-entry": ["hive-core/references/git-commit.md"]}`)
		}, []string{"flow-verify/stage-exit"}, []string{"flow-verify/stage-entry", "flow-build/stage"}},
		{"manifest", func(o Options) {
			if err := os.Remove(packetSource(o, "flow-verify/assets/stage.json")); err != nil {
				t.Fatal(err)
			}
		}, []string{"flow-verify/stage-entry", "flow-verify/stage-exit"}, []string{"flow-build/stage"}},
		{"flow", func(o Options) {
			if err := os.RemoveAll(packetSource(o, "flow-verify")); err != nil {
				t.Fatal(err)
			}
		}, []string{"flow-verify/stage-entry", "flow-verify/stage-exit"}, []string{"flow-build/stage"}},
	}
	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			o := packetOptions(t)
			apply(t, plan(t, "install", o))
			step.change(o)
			apply(t, plan(t, "install", o))
			for _, g := range step.gone {
				flow, key, _ := strings.Cut(g, "/")
				absent(t, packetPath(o, flow, key))
				if _, ok := stateFor(t, o).Records[packetPath(o, flow, key)]; ok {
					t.Errorf("record of %s kept", g)
				}
			}
			for _, k := range step.kept {
				flow, key, _ := strings.Cut(k, "/")
				if _, err := os.Lstat(packetPath(o, flow, key)); err != nil {
					t.Errorf("%s lost: %v", k, err)
				}
			}
		})
	}
}

func TestStagePacketGoesWithItsHostAndWithRemoval(t *testing.T) {
	o := packetOptions(t)
	apply(t, plan(t, "install", o))
	codexOnly := o
	codexOnly.Hosts = []string{"codex"}
	apply(t, plan(t, "remove", codexOnly))
	if _, err := os.Lstat(packetPath(o, "flow-build", "stage")); err != nil {
		t.Fatalf("packet lost while other hosts keep the skill: %v", err)
	}
	apply(t, plan(t, "remove", o))
	absent(t, filepath.Join(o.Home, ".agents", "skills", "flow-build"))
	if got := packetRecords(t, o); len(got) != 0 {
		t.Fatalf("records kept: %v", got)
	}
}

func TestStagePacketHandWrittenIsRejectedFromACheckoutAndAStoredRelease(t *testing.T) {
	for _, name := range []string{"stage.md", "stage-exit.md"} {
		o := packetOptions(t)
		put(t, packetSource(o, "flow-build/"+name), "# Mine\n")
		if _, err := BuildPlan("install", o); err == nil || !strings.Contains(err.Error(), "content/skills/flow-build/"+name) {
			t.Fatalf("%s: hand-written packet accepted or unnamed: %v", name, err)
		}
		absent(t, o.StateDir)
	}
	o := packetOptions(t)
	p := plan(t, "install", o)
	apply(t, p)
	rel := *p.Release
	rel.Files = append(append([]Payload{}, rel.Files...), Payload{Path: "content/skills/flow-build/stage.md", Data: []byte("# Mine\n"), Mode: 0o644})
	sort.SliceStable(rel.Files, func(i, j int) bool {
		if rel.Files[i].Path == GlobalSource || rel.Files[j].Path == GlobalSource {
			return rel.Files[i].Path == GlobalSource
		}
		return rel.Files[i].Path < rel.Files[j].Path
	})
	rel.ID = releaseID(rel)
	if err := os.WriteFile(filepath.Join(o.StateDir, "releases", rel.ID+".json"), encode(rel), 0o600); err != nil {
		t.Fatal(err)
	}
	stored := o
	stored.ReleaseID = rel.ID
	if _, err := BuildPlan("install", stored); err == nil || !strings.Contains(err.Error(), "content/skills/flow-build/stage.md") {
		t.Fatalf("stored release with a hand-written packet accepted: %v", err)
	}
}

func TestStagePacketBadManifestsAreRejectedBeforeAnyWrite(t *testing.T) {
	big := strings.Repeat("A line of rule text that adds up to the cap.\n", 600)
	many := strings.Repeat("x\n", 1100)
	cases := []struct {
		name, manifest, want string
		extra                map[string]string
	}{
		{"unknown key", `{"extra": ["hive-core/references/questions.md"]}`, `"extra"`, nil},
		{"key with digits", `{"stage2": ["hive-core/references/questions.md"]}`, `"stage2"`, nil},
		{"missing member", `{"stage": ["hive-core/references/nope.md"]}`, "hive-core/references/nope.md", nil},
		{"not json", `{"stage": [`, "assets/stage.json", nil},
		{"not an object", `["hive-core/references/questions.md"]`, "assets/stage.json", nil},
		{"value not an array", `{"stage": "hive-core/references/questions.md"}`, "assets/stage.json", nil},
		{"member not a string", `{"stage": [3]}`, "assets/stage.json", nil},
		{"empty list", `{"stage": []}`, `"stage"`, nil},
		{"no keys", `{}`, "assets/stage.json", nil},
		{"duplicate key", `{"stage": ["hive-core/references/questions.md"], "stage": ["hive-core/references/git-commit.md"]}`, `"stage"`, nil},
		{"duplicate member", `{"stage": ["hive-core/references/questions.md", "hive-core/references/questions.md"]}`, "hive-core/references/questions.md", nil},
		{"member not markdown", `{"stage": ["hive-core/scripts/run.py"]}`, "hive-core/scripts/run.py", map[string]string{"hive-core/scripts/run.py": "print(1)\n"}},
		{"member escapes the skills", `{"stage": ["../guidance/global.md"]}`, "../guidance/global.md", nil},
		{"over 24000 bytes", `{"stage": ["hive-core/references/big.md"]}`, "24000", map[string]string{"hive-core/references/big.md": "# Big\n\n" + big}},
		{"over 1000 lines", `{"stage": ["hive-core/references/many.md"]}`, "1000 lines", map[string]string{"hive-core/references/many.md": "# Many\n\n" + many}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := packetOptions(t)
			for rel, text := range tc.extra {
				put(t, packetSource(o, rel), text)
			}
			packetManifest(t, o, "flow-build", tc.manifest)
			_, err := BuildPlan("install", o)
			if err == nil {
				t.Fatal("release accepted")
			}
			for _, want := range []string{"flow-build", tc.want} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error %q does not name %q", err, want)
				}
			}
			absent(t, o.StateDir)
			absent(t, filepath.Join(o.Home, ".agents"))
		})
	}
}

func TestStagePacketCapsAreInclusive(t *testing.T) {
	o := packetOptions(t)
	// A packet of exactly 24,000 bytes passes: the cap is a maximum.
	head := "## hive-core/references/pad\n\n"
	end := "\n<!-- hive packet end: stage · hive-core/references/pad -->\n"
	pad := 24000 - len(head) - len(end)
	body := "# Pad\n\n" + strings.Repeat("x", pad-len("# Pad\n\n")-1) + "\n"
	put(t, packetSource(o, "hive-core/references/pad.md"), body)
	packetManifest(t, o, "flow-build", `{"stage": ["hive-core/references/pad.md"]}`)
	apply(t, plan(t, "install", o))
	if got := len(get(t, packetPath(o, "flow-build", "stage"))); got != 24000 {
		t.Fatalf("packet is %d bytes, the test wants exactly 24000", got)
	}
	put(t, packetSource(o, "hive-core/references/pad.md"), body+"y\n")
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("a packet one line over the byte cap was accepted")
	}
}

func TestStagePacketBrokenLinkIsCaughtInTheGeneratedPacket(t *testing.T) {
	// A member outside references/ is not validated as a file of its own, so a
	// broken link in it shows only once it is part of a packet.
	for _, link := range []string{"[x](missing.md)", "[x](#nope)"} {
		o := packetOptions(t)
		put(t, packetSource(o, "hive-core/scripts/notes.md"), "# Notes\n\nSee "+link+".\n")
		packetManifest(t, o, "flow-build", `{"stage": ["hive-core/scripts/notes.md"]}`)
		_, err := BuildPlan("install", o)
		if err == nil || !strings.Contains(err.Error(), "content/skills/flow-build/stage.md") || !strings.Contains(err.Error(), "instruction reference") {
			t.Fatalf("%s: broken link in the packet accepted or the packet unnamed: %v", link, err)
		}
	}
}

func TestStagePacketLinksResolveInTheInstalledStore(t *testing.T) {
	o := packetOptions(t)
	apply(t, plan(t, "install", o))
	packet := get(t, packetPath(o, "flow-build", "stage"))
	dir := filepath.Dir(packetPath(o, "flow-build", "stage"))
	for _, m := range regexp.MustCompile(`\]\(([^)#]+)(?:#[^)]*)?\)`).FindAllStringSubmatch(packet, -1) {
		if strings.HasPrefix(m[1], "http") || m[1] == "git-commit.md" {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, m[1])); err != nil {
			t.Errorf("link %s does not resolve from the installed packet: %v", m[1], err)
		}
	}
}

func TestStagePacketStatusReportsDrift(t *testing.T) {
	o := packetOptions(t)
	apply(t, plan(t, "install", o))
	statusOf := func() string {
		entries, err := Status(o)
		if err != nil {
			t.Fatal(err)
		}
		for _, en := range entries {
			if en.Path == packetPath(o, "flow-build", "stage") && en.Host == "codex" {
				return en.Status
			}
		}
		t.Fatal("packet missing from status")
		return ""
	}
	if got := statusOf(); got != "installed" {
		t.Fatalf("status = %q", got)
	}
	put(t, packetPath(o, "flow-build", "stage"), "edited\n")
	if got := statusOf(); got != "drift" {
		t.Fatalf("edited status = %q", got)
	}
}

func TestStagePacketRecordedShapeIsValidated(t *testing.T) {
	o := packetOptions(t)
	apply(t, plan(t, "install", o))
	path := packetPath(o, "flow-build", "stage")
	base := stateFor(t, o).Records[path]
	mutate := map[string]func(r *Record){
		"another kind":      func(r *Record) { r.Target.Kind = "hook" },
		"another skill":     func(r *Record) { r.Target.Source = "content/skills/flow-verify/stage.md" },
		"another file name": func(r *Record) { r.Target.Source = "content/skills/flow-build/stage-exit.md" },
		"not a packet key":  func(r *Record) { r.Target.Source = "content/skills/flow-build/stage2.md" },
	}
	for name, f := range mutate {
		s := stateFor(t, o)
		bad := base
		f(&bad)
		s.Records[path] = bad
		if err := normalizeState(&s); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	s := stateFor(t, o)
	if err := normalizeState(&s); err != nil {
		t.Fatalf("the state a release writes is refused: %v", err)
	}
}

// preE2ValidSource is the check of validSource before packets existed: its
// generated sources are hooks and policy files, and a skill resource lives
// under references, scripts, assets or playbooks.
func preE2ValidSource(source string) bool {
	if source == GlobalSource || source == "integrations/opencode/context-loader.js" {
		return true
	}
	if _, ok := generatedHooks()[source]; ok || generatedSkillPolicy(source) {
		return true
	}
	if regexp.MustCompile(`^content/agents/[a-z0-9]+(-[a-z0-9]+)*/[a-z0-9]+(-[a-z0-9]+)*\.md$`).MatchString(source) {
		return true
	}
	return regexp.MustCompile(`^content/skills/[a-z0-9]+(-[a-z0-9]+)*/(SKILL\.md|(references|scripts|assets|playbooks)/([A-Za-z0-9][A-Za-z0-9._-]*/)*(\.gitignore|[A-Za-z0-9][A-Za-z0-9._-]*\.(md|html|py|sh|json|ts|mjs|astro|mdx|css|yaml|yml|tsv)))$`).MatchString(source)
}

func TestStateAfterPacketsAreRetiredReadsWithTheSourceCheckOfBeforePackets(t *testing.T) {
	o := packetOptions(t)
	apply(t, plan(t, "install", o))
	// A binary from before packets rejects the records of a packet.
	if len(packetRecords(t, o)) == 0 {
		t.Fatal("no packet recorded")
	}
	for _, path := range packetRecords(t, o) {
		if preE2ValidSource(stateFor(t, o).Records[path].Target.Source) {
			t.Fatalf("%s would be accepted by the earlier check: the test proves nothing", path)
		}
	}
	// Retiring them, by dropping every manifest, leaves a state it reads.
	for _, flow := range []string{"flow-build", "flow-verify"} {
		if err := os.Remove(packetSource(o, flow+"/assets/stage.json")); err != nil {
			t.Fatal(err)
		}
	}
	apply(t, plan(t, "install", o))
	if got := packetRecords(t, o); len(got) != 0 {
		t.Fatalf("packet records kept: %v", got)
	}
	for path, r := range stateFor(t, o).Records {
		if !preE2ValidSource(r.Target.Source) {
			t.Errorf("%s records source %q that a binary from before packets refuses", path, r.Target.Source)
		}
	}
}
