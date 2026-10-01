package mdlinks

import "testing"

func TestRewriteSkillLinks(t *testing.T) {
	const dir = "/h/.agents/skills"
	cases := []struct{ name, in, want string }{
		{"inline", "See [x](skill:flow-build/references/a.md).", "See [x](</h/.agents/skills/flow-build/references/a.md>)."},
		{"angle destination", "[x](<skill:o/a.md>)", "[x](</h/.agents/skills/o/a.md>)"},
		{"title kept", `[x](skill:o/a.md "t")`, `[x](</h/.agents/skills/o/a.md> "t")`},
		{"two on a line", "[a](skill:o/a.md) and [b](skill:o/b.md)", "[a](</h/.agents/skills/o/a.md>) and [b](</h/.agents/skills/o/b.md>)"},
		{"escaped space", "[x](skill:o/a%20b.md)", "[x](</h/.agents/skills/o/a b.md>)"},
		{"backtick fence", "```\n[x](skill:o/a.md)\n```\n[y](skill:o/b.md)", "```\n[x](skill:o/a.md)\n```\n[y](</h/.agents/skills/o/b.md>)"},
		{"tilde fence", "~~~md\n[x](skill:o/a.md)\n~~~", "~~~md\n[x](skill:o/a.md)\n~~~"},
		{"reference definition", "[x]: skill:o/a.md", "[x]: skill:o/a.md"},
		{"other links", "[a](https://x.y) [b](references/c.md) [c](#h)", "[a](https://x.y) [b](references/c.md) [c](#h)"},
		{"no trailing newline", "[x](skill:o/a.md)\n", "[x](</h/.agents/skills/o/a.md>)\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := RewriteSkillLinks(c.in, dir); got != c.want {
				t.Fatalf("RewriteSkillLinks(%q)\n got %q\nwant %q", c.in, got, c.want)
			}
		})
	}
	if got := RewriteSkillLinks("[x](skill:o/a.md)", ".claude/skills"); got != "[x](<.claude/skills/o/a.md>)" {
		t.Fatalf("relative dir: %q", got)
	}
}

func TestOutsideFencesDropsFencedLinesAndDelimiters(t *testing.T) {
	in := "a\n```go\nb\n```\nc\n~~~\nd\n~~~\n"
	if got, want := OutsideFences(in), "a\nc\n\n"; got != want {
		t.Fatalf("OutsideFences = %q, want %q", got, want)
	}
}
