package hostevent

import "testing"

func TestCommandRoundTripsAndQuotesThePath(t *testing.T) {
	for _, exe := range []string{"/opt/hive/bin/hive", "/Users/o'brien/bin/hive", "/path with space/hive"} {
		cmd := Command(exe, "claude", "SessionStart")
		got, word, event, ok := ParseCommand(cmd)
		if !ok || got != exe || word != "claude" || event != "SessionStart" {
			t.Fatalf("ParseCommand(%q) = %q %q %q %v", cmd, got, word, event, ok)
		}
		if !IsHiveCommand(cmd, "claude", "SessionStart", nil) {
			t.Fatalf("%q is not recognized", cmd)
		}
	}
}

func TestForeignCommandsAreNeverHives(t *testing.T) {
	for _, cmd := range []string{
		"herdr hook claude SessionStart",
		"'/usr/bin/rex' hook claude SessionStart",
		"'/opt/hive/bin/hive' hook claude SessionEnd",
		"'/opt/hive/bin/hive' hook codex SessionStart",
		"'/opt/hive/bin/hive' hook claude SessionStart --extra",
		"'relative/hive' hook claude SessionStart",
		"/opt/hive/bin/hive hook claude SessionStart",
		"",
	} {
		if IsHiveCommand(cmd, "claude", "SessionStart", nil) {
			t.Fatalf("%q must not be recognized", cmd)
		}
	}
	other := "'/usr/bin/other' hook claude SessionStart"
	if !IsHiveCommand(other, "claude", "SessionStart", map[string]bool{other: true}) {
		t.Fatal("a recorded command must be recognized")
	}
}
