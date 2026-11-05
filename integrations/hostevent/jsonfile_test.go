package hostevent

import (
	"encoding/json"
	"testing"
)

func TestEntryShapesAndCommands(t *testing.T) {
	grouped := JSONHookFile{Word: "claude"}
	b, _ := json.Marshal(grouped.Entry("/opt/hive", "SessionEnd"))
	if string(b) != `{"hooks":[{"type":"command","command":"'/opt/hive' hook claude SessionEnd","timeout":5}]}` {
		t.Fatalf("grouped entry = %s", b)
	}
	flat := JSONHookFile{Word: "cursor", Flat: true}
	b, _ = json.Marshal(flat.Entry("/opt/hive", "stop"))
	if string(b) != `{"command":"'/opt/hive' hook cursor stop","timeout":5}` {
		t.Fatalf("flat entry = %s", b)
	}
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	if got := flat.Commands(m); len(got) != 1 || got[0] != "'/opt/hive' hook cursor stop" {
		t.Fatalf("flat commands = %v", got)
	}
	_ = json.Unmarshal([]byte(`{"matcher":"x","hooks":[{"type":"command","command":"a"},{"type":"command","command":"b"}]}`), &m)
	if got := grouped.Commands(m); len(got) != 2 {
		t.Fatalf("grouped commands = %v", got)
	}
	_ = json.Unmarshal([]byte(`{"hooks":"nope"}`), &m)
	if got := grouped.Commands(m); got != nil {
		t.Fatalf("a malformed group has no commands, got %v", got)
	}
}
