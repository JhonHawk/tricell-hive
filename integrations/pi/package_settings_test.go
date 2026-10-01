package pi

import "testing"

func TestClassifySubagents(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		in     string
		status Status
		source string
		err    bool
	}{
		{"empty object", `{}`, Absent, "", false},
		{"packages missing", `{"theme":"dark"}`, Absent, "", false},
		{"empty list", `{"packages":[]}`, Absent, "", false},
		{"other package only", `{"packages":["npm:other@1.0.0"]}`, Absent, "", false},
		{"pinned 0.67.0", `{"packages":["npm:pi-subagents@0.67.0"]}`, Present, "npm:pi-subagents@0.67.0", false},
		{"pinned 0.74.0", `{"packages":["npm:pi-subagents@0.74.0"]}`, Present, "npm:pi-subagents@0.74.0", false},
		{"unpinned", `{"packages":["npm:pi-subagents"]}`, Present, "npm:pi-subagents", false},
		{"object without extensions", `{"packages":[{"source":"npm:pi-subagents@0.74.0"}]}`, Present, "npm:pi-subagents@0.74.0", false},
		{"empty extensions", `{"packages":[{"source":"npm:pi-subagents@0.74.0","extensions":[]}]}`, Conflict, "npm:pi-subagents@0.74.0", false},
		{"non-empty extensions filter", `{"packages":[{"source":"npm:pi-subagents","extensions":["prompts/*"]}]}`, Conflict, "npm:pi-subagents", false},
		{"invalid json", `{`, Absent, "", true},
		{"packages not a list", `{"packages":"npm:pi-subagents"}`, Absent, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ClassifySubagents([]byte(tc.in))
			if tc.err {
				if err == nil {
					t.Fatalf("ClassifySubagents(%s): want error", tc.name)
				}
				return
			}
			if err != nil {
				t.Fatalf("ClassifySubagents(%s): %v", tc.name, err)
			}
			if got.Status != tc.status || got.Source != tc.source {
				t.Fatalf("ClassifySubagents(%s) = {%v %q}, want {%v %q}", tc.name, got.Status, got.Source, tc.status, tc.source)
			}
		})
	}
}
