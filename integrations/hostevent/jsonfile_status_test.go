package hostevent

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEntryCarriesStatusMessageOnlyWhenConfigured(t *testing.T) {
	named := JSONHookFile{Word: "codex", StatusMessages: map[string]string{"SessionStart": "Hive: report session start to the viewer"}}
	b, _ := json.Marshal(named.Entry("/opt/hive", "SessionStart"))
	if string(b) != `{"hooks":[{"type":"command","command":"'/opt/hive' hook codex SessionStart","timeout":5,"statusMessage":"Hive: report session start to the viewer"}]}` {
		t.Fatalf("named entry = %s", b)
	}
	b, _ = json.Marshal(named.Entry("/opt/hive", "SessionEnd"))
	if strings.Contains(string(b), "statusMessage") {
		t.Fatalf("an event without a text must not write the field: %s", b)
	}
	b, _ = json.Marshal(JSONHookFile{Word: "claude"}.Entry("/opt/hive", "SessionEnd"))
	if strings.Contains(string(b), "statusMessage") {
		t.Fatalf("a file without texts must not write the field: %s", b)
	}
}
