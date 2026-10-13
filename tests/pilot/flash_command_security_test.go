package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestFlashCommandCredentialShapesAreWithheld(t *testing.T) {
	cases := map[string]string{
		"short user flag":    "curl -u demo:EXAMPLE_VALUE https://example.invalid",
		"attached user flag": "curl -udemo:EXAMPLE_VALUE https://example.invalid",
		"long user flag":     "curl --user 'demo:EXAMPLE_VALUE' https://example.invalid",
		"short pair flag":    "tool -x=a:b",
		"generic pair flag":  "tool --x=a:b",
		"spaced pair flag":   "tool --x a:b",
		"URL userinfo":       "curl https://demo:EXAMPLE_VALUE@example.invalid",
		"URL username":       "curl https://demo@example.invalid",
		"authorization":      "tool -H 'Authorization: EXAMPLE_VALUE'",
		"bearer":             "tool -H 'Bearer EXAMPLE_VALUE'",
		"basic":              "tool -H 'Basic EXAMPLE_VALUE'",
		"key assignment":     "key=EXAMPLE_VALUE tool",
		"token flag":         "tool --TOKEN EXAMPLE_VALUE",
		"secret assignment":  "secret = 'EXAMPLE_VALUE'",
		"password flag":      "tool --PASSWORD=EXAMPLE_VALUE",
		"passwd assignment":  "PASSWD=EXAMPLE_VALUE tool",
		"credential flag":    "tool --credential-file fixture",
		"auth assignment":    "Auth = 'EXAMPLE_VALUE'",
		"env key":            "env ACCESS_KEY=EXAMPLE_VALUE tool",
		"private key":        "execute('-----BEGIN PRIVATE KEY-----\\nEXAMPLE_VALUE\\n-----END PRIVATE KEY-----')",
		"hex token":          "tool 0123456789abcdef0123456789abcdef",
		"base64 token":       "tool aBcD1234EfGh5678IjKl9012+/ab",
		"URL safe token":     "tool aBcD1234_EfGh5678-IjKl9012",
	}
	for name, command := range cases {
		t.Run(name, func(t *testing.T) {
			if flashOperationCommand(command) != "" {
				t.Error("credential-shaped text accepted")
			}
			model := map[string]any{"providerID": "opencode-go", "id": "claude-haiku-5-5", "variant": "high"}
			for _, tool := range []string{"shell", "execute"} {
				key := "command"
				if tool == "execute" {
					key = "code"
				}
				doc := map[string]any{"info": map[string]any{"id": "root", "agent": "build", "model": model, "outcome": "succeeded", "time": map[string]any{"idle": 1}}, "messages": []any{map[string]any{"id": "final", "type": "assistant", "agent": "build", "model": model, "finish": "stop", "time": map[string]any{"completed": 1}, "content": []any{map[string]any{"type": "tool", "name": tool, "state": map[string]any{"status": "completed", "input": map[string]any{key: command}}}}}}}
				raw, _ := json.Marshal(doc)
				c, err := flashChildFromExport(raw, "", "build", flashModels[0])
				if err != nil {
					t.Fatal("native projection failed")
				}
				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				if err := saveFlashProjection(root, &c); err != nil {
					root.Close()
					t.Fatal(err)
				}
				evidence, err := root.ReadFile(c.EvidenceFile)
				root.Close()
				if err != nil {
					t.Fatal(err)
				}
				var projected struct {
					Operations []struct {
						Command  string `json:"command"`
						Withheld bool   `json:"command_withheld"`
					}
				}
				if err := json.Unmarshal(evidence, &projected); err != nil {
					t.Fatal(err)
				}
				if len(projected.Operations) != 1 || projected.Operations[0].Command != "" || !projected.Operations[0].Withheld {
					t.Error("persisted projection did not withhold text")
				}
				if strings.Contains(string(evidence), "EXAMPLE_VALUE") {
					t.Error("synthetic credential persisted")
				}
			}
		})
	}
}

func TestFlashCommandOrdinaryReadOnlyTextIsPreserved(t *testing.T) {
	for _, command := range []string{"ls", "cat /fixture/plan.md", "rg needle /fixture", "git status"} {
		if flashOperationCommand(command) != command {
			t.Error("ordinary read-only command withheld")
		}
	}
}
