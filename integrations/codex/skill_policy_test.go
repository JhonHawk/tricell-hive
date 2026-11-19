package codex

import "testing"

func TestHiddenFromModelReadsOnlyTheFrontmatter(t *testing.T) {
	for text, want := range map[string]bool{
		"---\nname: a\ndisable-model-invocation: true\n---\nbody\n":        true,
		"---\r\nname: a\r\ndisable-model-invocation: true\r\n---\r\nb\r\n": true,
		"---\ndisable-model-invocation: \"true\" # hidden\n---\n":          true,
		"---\ndisable-model-invocation: false\n---\n":                      false,
		"---\nname: a\n---\ndisable-model-invocation: true\n":              false,
		"disable-model-invocation: true\n":                                 false,
		"---\n  disable-model-invocation: true\n---\n":                     false,
		"---\nx-disable-model-invocation: true\n---\n":                     false,
		"": false,
	} {
		if got := HiddenFromModel([]byte(text)); got != want {
			t.Errorf("HiddenFromModel(%q) = %v, want %v", text, got, want)
		}
	}
}

func TestSkillPolicyPath(t *testing.T) {
	if got := SkillPolicyPath("/s/leaf"); got != "/s/leaf/agents/openai.yaml" {
		t.Fatal(got)
	}
}
