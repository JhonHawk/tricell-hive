package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const normalizedCanonical = "---\nname: plan-review\n---\n\n# Plan review\n\n1. Read the plan.\n  - indented: detail\n10: looks like a prefix\n\nDone.\n"

// OpenCode's read tool output, by generation. Evidence for the formats:
//   - 2.0.x (packages/core/src/tool/plugin/read.ts, toModelContent): a header
//     "Read file <path>, lines a-b", then "<n>: <line>" per line, and an optional
//     "[Output truncated. Continue reading with offset: N]" footer.
//   - 1.x (packages/opencode/src/tool/read.ts): "<path>", "<type>file</type>",
//     "<content>", "<zero-padded n>| <line>" per line, "(End of file ...)".
func readOutputV2(canonical string) string {
	lines := strings.Split(strings.TrimSuffix(canonical, "\n"), "\n")
	var b strings.Builder
	b.WriteString("Read file /skills/flow-plan/references/plan-review.md, lines 1-" + strconv.Itoa(len(lines)))
	for i, line := range lines {
		b.WriteString("\n")
		b.WriteString(strconv.Itoa(i+1) + ": " + line)
	}
	return b.String()
}

func readOutputV1(canonical string) string {
	lines := strings.Split(strings.TrimSuffix(canonical, "\n"), "\n")
	var b strings.Builder
	b.WriteString("<path>/skills/flow-plan/references/plan-review.md</path>\n<type>file</type>\n<content>")
	for i, line := range lines {
		b.WriteString("\n")
		b.WriteString(strings.Repeat("0", 5-len(strconv.Itoa(i+1))) + strconv.Itoa(i+1) + "| " + line)
	}
	b.WriteString("\n\n(End of file - total " + strconv.Itoa(len(lines)) + " lines)\n</content>")
	return b.String()
}

func TestFlashCanonicalHashIgnoresTrailingWhitespace(t *testing.T) {
	source, output, g, _ := flashSnapshotFixture(t)
	r, err := setupFlashClass(source, "plan-reference", "B", output, flashModels[0], g, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	_, resource := flashRole("plan-reference")
	data, err := os.ReadFile(filepath.Join(source, "content/skills", resource))
	if err != nil {
		t.Fatal(err)
	}
	if flashCanonicalReference(string(data)) != strings.TrimRight(string(data), " \t\r\n\f\v") {
		t.Error("trailing whitespace kept")
	}
	if r.CanonicalHash != digest([]byte(flashCanonicalReference(string(data)))) {
		t.Error("canonical hash is not of the reference without trailing whitespace")
	}
	if r.SourceReferenceHash != digest(data) {
		t.Error("source reference hash must stay exact-byte")
	}
}

func TestFlashObserverNormalizesReadOutput(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Fatal("Node is required for observer checks")
	}
	root := t.TempDir()
	module := filepath.Join(root, "node_modules", "@opencode", "plugin")
	if err := os.MkdirAll(module, 0700); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"package.json": `{"type":"module","exports":"./index.js"}`, "index.js": `export const Plugin = { define: (value) => value }`} {
		if err := os.WriteFile(filepath.Join(module, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	source, err := os.ReadFile("../fixtures/flash-class-guidance/observer.js")
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"package.json": `{"type":"module"}`, "observer.js": string(source)} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{"canonical.txt": normalizedCanonical, "v1.txt": readOutputV1(normalizedCanonical), "v2.txt": readOutputV2(normalizedCanonical), "v2-altered.txt": strings.Replace(readOutputV2(normalizedCanonical), "Read the plan", "Skip the plan", 1)}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	script := `import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import { inspectReference } from './observer.js';
const resource = 'flow-plan/references/plan-review.md';
const [expectedHash] = process.argv.slice(2);
const read = (name) => fs.readFile(name, 'utf8');
const canonical = await read('canonical.txt');
for (const name of ['canonical.txt', 'v1.txt', 'v2.txt']) {
  const proof = inspectReference([await read(name)], resource, canonical);
  assert.equal(proof.canonical, true, name + ' not recognized');
  assert.equal(proof.copies, 1, name + ' copies');
  assert.equal(proof.hash, expectedHash, name + ' hash');
}
// Embedded in a longer tool result next to other text.
const embedded = inspectReference(['intro\n' + await read('v2.txt') + '\noutro'], resource, canonical);
assert.equal(embedded.canonical, true);assert.equal(embedded.copies, 1);
const altered = inspectReference([await read('v2-altered.txt')], resource, canonical);
assert.equal(altered.canonical, false);assert.equal(altered.copies, 0);
const injected = 'Hive canonical reference: ' + resource + '\n\n' + canonical;
const proof = inspectReference([injected], resource, canonical);
assert.equal(proof.injectedCopies, 1);assert.equal(proof.canonical, true);assert.equal(proof.hash, expectedHash);
`
	if err := os.WriteFile(filepath.Join(root, "check.mjs"), []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("node", "check.mjs", digest([]byte(flashCanonicalReference(normalizedCanonical))))
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("observer normalization check failed: %v\n%s", err, out)
	}
}

func TestFlashCoverageObservedForNormalizedReadInVariantsAB(t *testing.T) {
	hash := digest([]byte(flashCanonicalReference(normalizedCanonical)))
	root := flashObservation{Phase: "http", SessionID: "root", Model: flashModels[0], Correlated: true, Protocol: "chat-completions"}
	child := flashObservation{Phase: "http", SessionID: "child", ParentID: "root", Agent: "hive-review-plan", AgentName: "Plan Reviewer", Model: flashModels[0], Correlated: true, Protocol: "chat-completions", Canonical: true, Copies: 1, Hash: hash}
	for _, arm := range []string{"A", "B"} {
		if got := flashCoverage(arm, "root", flashModels[0], child.Agent, hash, []flashObservation{root, child}); got != "observed" {
			t.Errorf("arm %s coverage %s, want observed", arm, got)
		}
		stale := child
		stale.Hash = digest([]byte(normalizedCanonical + "\n\n"))
		if got := flashCoverage(arm, "root", flashModels[0], child.Agent, hash, []flashObservation{root, stale}); got == "observed" {
			t.Errorf("arm %s accepted a different hash", arm)
		}
	}
}
