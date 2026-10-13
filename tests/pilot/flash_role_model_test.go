package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFlashClassRoleMarkdownModelOverride(t *testing.T) {
	old := "github-copilot/gpt-6.1-sol#medium"
	for _, value := range []string{`"` + old + `"`, old, `'` + old + `'`, ""} {
		t.Run(value, func(t *testing.T) {
			config := t.TempDir()
			agents := filepath.Join(config, ".config/opencode/agents")
			if err := os.MkdirAll(agents, 0700); err != nil {
				t.Fatal(err)
			}
			body := "\n# Review\nKeep bytes  \nmodel: body value\n---\n"
			front := "---\nname: role\ndescription: Keep unchanged\n"
			line := ""
			if value != "" {
				line = "model: " + value + "\n"
			}
			raw := front + line + "tools: [read]\n---\n" + body
			target := filepath.Join(agents, "hive-review-plan.md")
			other := filepath.Join(agents, "hive-review-harness.md")
			if err := os.WriteFile(target, []byte(raw), 0640); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(other, []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
			before, after, original, err := overrideFlashRoleModel(config, "hive-review-plan", flashModels[0])
			if err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			expected := strings.Replace(raw, old, flashModels[0], 1)
			if value == "" {
				expected = front + "tools: [read]\nmodel: \"" + flashModels[0] + "\"\n---\n" + body
			}
			if string(got) != expected {
				t.Error("role model not replaced with all other bytes preserved")
			}
			wantOriginal := old
			if value == "" {
				wantOriginal = ""
			}
			if before != digest([]byte(raw)) || after != digest(got) || before == after || original != wantOriginal {
				t.Error("role provenance missing or incorrect")
			}
			untouched, err := os.ReadFile(other)
			if err != nil || string(untouched) != raw {
				t.Fatal("another role changed")
			}
			info, _ := os.Stat(target)
			if info.Mode().Perm() != 0640 {
				t.Error("role file permissions changed")
			}
		})
	}
}

func TestFlashClassRoleModelRequiresFrontmatter(t *testing.T) {
	for _, raw := range []string{"# No frontmatter\nmodel: body\n", "---\nname: role\n"} {
		config := t.TempDir()
		os.MkdirAll(filepath.Join(config, ".config/opencode/agents"), 0700)
		target := filepath.Join(config, ".config/opencode/agents/hive-review-plan.md")
		os.WriteFile(target, []byte(raw), 0600)
		if _, _, _, err := overrideFlashRoleModel(config, "hive-review-plan", flashModels[0]); err == nil {
			t.Error("missing or incomplete frontmatter accepted")
		}
		got, _ := os.ReadFile(target)
		if string(got) != raw {
			t.Error("invalid role modified")
		}
	}
}

func TestFlashClassAuthExportStandalone(t *testing.T) {
	g := &guidanceVariant{host: "opencode", shadowHome: t.TempDir()}
	export, imp := openCodeAuthCommands(context.Background(), g, []string{"HOME=/fake-home"}, "opencode-go")
	if strings.Join(export.Args, " ") != "opencode auth export opencode-go --standalone" {
		t.Errorf("export can start managed service: %v", export.Args)
	}
	if strings.Join(imp.Args, " ") != "opencode auth import --standalone" {
		t.Fatal("import command changed")
	}
}

func TestFlashClassRoleModelPreservesCRLFAndComments(t *testing.T) {
	config := t.TempDir()
	os.MkdirAll(filepath.Join(config, ".config/opencode/agents"), 0700)
	raw := "---\r\nmodel:  \"old/provider#medium\"  # Keep comment\r\nother: value\r\n---\r\nBody with CRLF  \r\n"
	target := filepath.Join(config, ".config/opencode/agents/hive-review-plan.md")
	os.WriteFile(target, []byte(raw), 0600)
	_, _, original, err := overrideFlashRoleModel(config, "hive-review-plan", flashModels[0])
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(target)
	if original != "old/provider#medium" || string(got) != strings.Replace(raw, "old/provider#medium", flashModels[0], 1) {
		t.Error("line endings, comment or other bytes changed")
	}
}

func TestFlashClassRoleModelRejectsAlternativeAndDuplicateKeys(t *testing.T) {
	forms := []string{`"model": another/provider`, `'model': another/provider`, `model : another/provider`, `model    : another/provider`, "model\t: another/provider", `"model" : another/provider`, `'model' : another/provider`, `"\u006dodel": another/provider`, "  model: another/provider"}
	cases := map[string]string{"duplicate plain": "model: old/provider\nmodel: another/provider\n"}
	for _, form := range forms {
		cases["alternative only "+form] = form + "\n"
		cases["plain first "+form] = "model: old/provider\n" + form + "\n"
		cases["plain last "+form] = form + "\nmodel: old/provider\n"
	}
	for name, declarations := range cases {
		t.Run(name, func(t *testing.T) {
			config := t.TempDir()
			agents := filepath.Join(config, ".config/opencode/agents")
			if err := os.MkdirAll(agents, 0700); err != nil {
				t.Fatal(err)
			}
			raw := "---\nname: role\n" + declarations + "description: Keep unchanged\n---\n# Keep body bytes\n"
			target := filepath.Join(agents, "hive-review-plan.md")
			if err := os.WriteFile(target, []byte(raw), 0640); err != nil {
				t.Fatal(err)
			}
			if _, _, _, err := overrideFlashRoleModel(config, "hive-review-plan", flashModels[0]); err == nil {
				t.Error("alternative or duplicate model key accepted")
			}
			got, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != raw {
				t.Error("rejected role was written")
			}
			info, err := os.Stat(target)
			if err != nil || info.Mode().Perm() != 0640 {
				t.Error("rejected role permissions changed")
			}
		})
	}
}
