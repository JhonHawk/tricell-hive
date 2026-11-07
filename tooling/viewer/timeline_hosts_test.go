package viewer

import "testing"

func TestTimelineSourceAcceptsEveryHostAndOnlyThose(t *testing.T) {
	for _, host := range []TimelineHost{"codex", "claude-code", "grok", "pi", "opencode", "cursor", "antigravity"} {
		if err := timelineSource(TimelineSource{ProjectRoot: "/work/app", Host: host, NativeSessionID: "native-1"}); err != nil {
			t.Errorf("%s: %v", host, err)
		}
	}
	err := timelineSource(TimelineSource{ProjectRoot: "/work/app", Host: "gemini", NativeSessionID: "native-1"})
	if TimelineContractCode(err) != "unsupported_host" {
		t.Fatalf("an unknown host: %v", err)
	}
}

func TestTimelineProtocolListsAntigravityAsASourceHost(t *testing.T) {
	brief, err := TimelineProtocolBrief("delivery")
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range brief.Operations {
		for _, host := range command.Enums["source.host"] {
			if host == "antigravity" {
				return
			}
		}
	}
	t.Fatal("the protocol brief's source.host enum lacks antigravity")
}
