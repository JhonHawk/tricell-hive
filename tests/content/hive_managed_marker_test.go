package content

import (
	"os"
	"path/filepath"
	"testing"
)

// hiveManagedMarker is the file each Hive skill ships so that a project's
// skills registry check (audit-harness's skill_registry.py) can tell the skills
// Hive installed in a project from the third-party skills the registry lists,
// on any machine, without the manager's state.
const hiveManagedMarker = "{\"managed_by\": \"tricell-hive\"}\n"

func TestEverySkillShipsTheHiveManagedMarker(t *testing.T) {
	entries, err := filepath.Glob("../../content/skills/*/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no skills found under content/skills")
	}
	for _, entry := range entries {
		dir := filepath.Dir(entry)
		marker := filepath.Join(dir, "assets", "hive-managed.json")
		data, err := os.ReadFile(marker)
		if err != nil {
			t.Errorf("%s: missing assets/hive-managed.json, so a project's skills registry check would count this Hive skill as unlisted", filepath.ToSlash(dir[len("../../"):]))
			continue
		}
		if string(data) != hiveManagedMarker {
			t.Errorf("%s: assets/hive-managed.json is %q, want %q", filepath.ToSlash(dir[len("../../"):]), data, hiveManagedMarker)
		}
	}
}
