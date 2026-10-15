package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func flashSnapshotFixture(t *testing.T) (string, string, *guidanceVariant, map[string]any) {
	t.Helper()
	source, shadow, output := t.TempDir(), t.TempDir(), t.TempDir()
	config := filepath.Join(shadow, ".config", "opencode")
	skills := filepath.Join(config, "skills")
	files := map[string]string{
		filepath.Join(source, "tests/fixtures", flashSuite, "observer.js"): "// Independent observer\n",
		filepath.Join(source, "content/guidance/global.md"):                "# Source guidance\n",
		filepath.Join(config, "opencode.json"):                             `{"mcp":{"servers":{"engram":{"environment":{}}}}}`,
		filepath.Join(config, "hive/context-loader/index.js"):              "// Delivered snapshot module\n",
	}
	refs := []any{}
	for _, caseID := range []string{"harness-reference", "plan-reference"} {
		role, resource := flashRole(caseID)
		text := "\ufeff# " + role + "\r\nQuotes \" and \\ Unicode ñ __HIVE_CONTEXT_SNAPSHOT_DATA__\r\n"
		files[filepath.Join(source, "content/skills", resource)] = text
		files[filepath.Join(skills, resource)] = text
		files[filepath.Join(config, "agents", role+".md")] = "---\nmodel: github-copilot/gpt-6.1-sol#medium\n---\n# Keep role instructions\n"
		refs = append(refs, map[string]any{"role": role, "resource": resource, "sha256": digest([]byte(text))})
	}
	metadata := map[string]any{"schema": 2, "references": refs}
	b, _ := json.Marshal(metadata)
	files[filepath.Join(config, "hive/context-loader/context-loader.json")] = string(b)
	for path, text := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	g := &guidanceVariant{shadowHome: shadow, host: "opencode", report: guidanceVariantReport{Source: source, FlowBuildSkillPath: filepath.Join(skills, "flow-build/SKILL.md")}}
	return source, output, g, metadata
}

func TestFlashClassSnapshotMetadataRejectsBeforeActivation(t *testing.T) {
	cases := map[string]func(map[string]any, string, *guidanceVariant){
		"old schema":      func(m map[string]any, _ string, _ *guidanceVariant) { m["schema"] = 1 },
		"root capability": func(m map[string]any, _ string, g *guidanceVariant) { m["skillsDirectory"] = g.shadowHome },
		"unknown field":   func(m map[string]any, _ string, _ *guidanceVariant) { m["extra"] = true },
		"missing resource": func(m map[string]any, _ string, _ *guidanceVariant) {
			delete(object(list(m["references"])[0]), "resource")
		},
		"text capability": func(m map[string]any, _ string, _ *guidanceVariant) {
			object(list(m["references"])[0])["text"] = "untrusted"
		},
		"wrong role": func(m map[string]any, _ string, _ *guidanceVariant) {
			object(list(m["references"])[0])["role"] = "hive-build-backend"
		},
		"other role resource": func(m map[string]any, _ string, _ *guidanceVariant) {
			object(list(m["references"])[0])["resource"] = "flow-plan/references/plan-review.md"
		},
		"digest mismatch": func(m map[string]any, _ string, _ *guidanceVariant) {
			object(list(m["references"])[0])["sha256"] = strings.Repeat("0", 64)
		},
		"uppercase digest": func(m map[string]any, _ string, _ *guidanceVariant) {
			r := object(list(m["references"])[0])
			r["sha256"] = strings.ToUpper(str(r["sha256"]))
		},
		"wrong order": func(m map[string]any, _ string, _ *guidanceVariant) {
			r := list(m["references"])
			r[0], r[1] = r[1], r[0]
		},
		"missing entry": func(m map[string]any, _ string, _ *guidanceVariant) { m["references"] = list(m["references"])[:1] },
		"extra entry": func(m map[string]any, _ string, _ *guidanceVariant) {
			m["references"] = append(list(m["references"]), list(m["references"])[0])
		},
		"installed tamper": func(_ map[string]any, _ string, g *guidanceVariant) {
			_ = os.WriteFile(filepath.Join(filepath.Dir(filepath.Dir(g.report.FlowBuildSkillPath)), "harness-audit/SKILL.md"), []byte("tampered"), 0600)
		},
		"source tamper": func(_ map[string]any, source string, _ *guidanceVariant) {
			_ = os.WriteFile(filepath.Join(source, "content/skills/harness-audit/SKILL.md"), []byte("tampered"), 0600)
		},
		"same digest different source": func(m map[string]any, source string, g *guidanceVariant) {
			path := filepath.Join(filepath.Dir(filepath.Dir(g.report.FlowBuildSkillPath)), "harness-audit/SKILL.md")
			_ = os.WriteFile(path, []byte("tampered"), 0600)
			object(list(m["references"])[0])["sha256"] = digest([]byte("tampered"))
		},
		"invalid UTF-8": func(m map[string]any, source string, g *guidanceVariant) {
			data := []byte{0xff}
			_ = os.WriteFile(filepath.Join(source, "content/skills/harness-audit/SKILL.md"), data, 0600)
			_ = os.WriteFile(filepath.Join(filepath.Dir(filepath.Dir(g.report.FlowBuildSkillPath)), "harness-audit/SKILL.md"), data, 0600)
			object(list(m["references"])[0])["sha256"] = digest(data)
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			source, output, g, metadata := flashSnapshotFixture(t)
			config := filepath.Join(g.shadowHome, ".config/opencode")
			before, err := os.ReadFile(filepath.Join(config, "opencode.json"))
			if err != nil {
				t.Fatal(err)
			}
			mutate(metadata, source, g)
			data, _ := json.Marshal(metadata)
			if err := os.WriteFile(filepath.Join(config, "hive/context-loader/context-loader.json"), data, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := setupFlashClass(source, "plan-reference", "C", output, flashModels[0], g, "fixture"); err == nil {
				t.Fatal("invalid snapshot provenance accepted")
			}
			after, err := os.ReadFile(filepath.Join(config, "opencode.json"))
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("rejected metadata changed native configuration")
			}
			if _, err := os.Stat(filepath.Join(config, "hive/pilot-observer/index.js")); !os.IsNotExist(err) {
				t.Fatal("rejected metadata wrote observer before validation")
			}
		})
	}
}

func TestFlashClassSnapshotExactBytesAndObserverOnlyArm(t *testing.T) {
	for _, arm := range []string{"B", "C"} {
		for _, caseID := range []string{"plan-reference", "harness-reference"} {
			t.Run(arm+"/"+caseID, func(t *testing.T) {
				source, output, g, _ := flashSnapshotFixture(t)
				if arm == "B" {
					_ = os.WriteFile(filepath.Join(g.shadowHome, ".config/opencode/hive/context-loader/context-loader.json"), []byte("invalid metadata unused by B"), 0600)
				}
				r, err := setupFlashClass(source, caseID, arm, output, flashModels[0], g, "fixture")
				if err != nil {
					t.Fatal(err)
				}
				_, resource := flashRole(caseID)
				data, err := os.ReadFile(filepath.Join(source, "content/skills", resource))
				if err != nil {
					t.Fatal(err)
				}
				if r.CanonicalHash != digest(data) || r.SourceReferenceHash != digest(data) || r.LoaderHash == "absent" || r.LoaderMetadataHash == "absent" || r.Coverage != "unverified" {
					t.Fatal("exact-byte hashes or independent coverage lost")
				}
				if len(r.ConfiguredPluginOrder) != map[string]int{"B": 1, "C": 2}[arm] || r.ConfiguredPluginOrder[len(r.ConfiguredPluginOrder)-1] != "hive.pilot-observer" {
					t.Fatal("plugin order changed")
				}
			})
		}
	}
}

func TestFlashClassSnapshotMalformedMetadata(t *testing.T) {
	for _, mode := range []string{"missing", "malformed", "trailing JSON", "duplicate key", "wrong field casing"} {
		t.Run(mode, func(t *testing.T) {
			source, output, g, metadata := flashSnapshotFixture(t)
			config := filepath.Join(g.shadowHome, ".config/opencode")
			path := filepath.Join(config, "hive/context-loader/context-loader.json")
			before, err := os.ReadFile(filepath.Join(config, "opencode.json"))
			if err != nil {
				t.Fatal(err)
			}
			data, _ := json.Marshal(metadata)
			switch mode {
			case "missing":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			case "malformed":
				data = []byte("{invalid")
			case "trailing JSON":
				data = append(data, []byte(" {}")...)
			case "duplicate key":
				data = []byte(strings.Replace(string(data), `"schema":2`, `"schema":1,"schema":2`, 1))
			case "wrong field casing":
				data = []byte(strings.Replace(string(data), `"schema":2`, `"Schema":2`, 1))
			}
			if mode != "missing" {
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := setupFlashClass(source, "plan-reference", "C", output, flashModels[0], g, "fixture"); err == nil {
				t.Fatal("ambiguous or malformed metadata accepted")
			}
			after, err := os.ReadFile(filepath.Join(config, "opencode.json"))
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("invalid metadata rewrote native configuration")
			}
			if _, err := os.Stat(filepath.Join(config, "hive/pilot-observer/index.js")); !os.IsNotExist(err) {
				t.Fatal("invalid metadata wrote observer")
			}
		})
	}
}
