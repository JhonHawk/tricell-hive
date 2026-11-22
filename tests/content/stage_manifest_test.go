package content

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var stageKeyPattern = regexp.MustCompile(`^stage(-[a-z]+)?$`)

// stageFlows are the flows that read a stage packet at entry.
var stageFlows = []string{
	"hive-mode", "flow-plan", "flow-build", "flow-verify", "flow-close",
	"git-workflow", "flow-deploy", "unattended-delegation",
}

// readStageManifest returns the manifest of a flow, keyed by packet key, with
// each member as a path under content/skills.
func readStageManifest(t *testing.T, flow string) map[string][]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "content", "skills", flow, "assets", "stage.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := map[string][]string{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("%s/assets/stage.json: %v", flow, err)
	}
	return manifest
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestStageManifests(t *testing.T) {
	skills := filepath.Join(repoRoot(t), "content", "skills")
	found, err := filepath.Glob(filepath.Join(skills, "*", "assets", "stage.json"))
	if err != nil {
		t.Fatal(err)
	}
	var withManifest []string
	for _, path := range found {
		withManifest = append(withManifest, filepath.Base(filepath.Dir(filepath.Dir(path))))
	}
	sort.Strings(withManifest)
	want := append([]string{}, stageFlows...)
	sort.Strings(want)
	if strings.Join(withManifest, ",") != strings.Join(want, ",") {
		t.Fatalf("flows with a stage manifest: %v, want %v", withManifest, want)
	}
	for _, flow := range stageFlows {
		manifest := readStageManifest(t, flow)
		if len(manifest) == 0 {
			t.Errorf("%s: empty stage manifest", flow)
		}
		skill, err := os.ReadFile(filepath.Join(skills, flow, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		for key, members := range manifest {
			if !stageKeyPattern.MatchString(key) {
				t.Errorf("%s: key %q does not match %s", flow, key, stageKeyPattern)
			}
			total := 0
			for _, member := range members {
				data, err := os.ReadFile(filepath.Join(skills, filepath.FromSlash(member)))
				if err != nil {
					t.Errorf("%s/%s: member %s: %v", flow, key, member, err)
					continue
				}
				total += len(data)
			}
			t.Logf("%s/%s: %d members, %d B", flow, key, len(members), total)
			// The flow names its packet at entry, with the fallback for a missing
			// or cut packet, so an agent never depends on the generated file alone.
			if !strings.Contains(string(skill), "read `"+key+".md` ") {
				t.Errorf("%s: SKILL.md does not tell the agent to read `%s.md`", flow, key)
			}
		}
		for _, phrase := range []string{"read the linked leaves instead", "packet end line, read the rest", "names is missing, read that leaf", "says it was cut, read the linked leaves", "needs no new read until a compaction, after which read"} {
			if !strings.Contains(string(skill), phrase) {
				t.Errorf("%s: SKILL.md lacks the packet fallback %q", flow, phrase)
			}
		}
	}
}
