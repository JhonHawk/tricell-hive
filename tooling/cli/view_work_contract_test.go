package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestViewWorkOfflineContract(t *testing.T) {
	var out bytes.Buffer
	if err := runView([]string{"work", "contract"}, &out, realViewDeps()); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"archive-rejected", "native-session-id", "4096", "tasks register", "--outcome completed"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("contract missing %q", want)
		}
	}
}

func TestViewWorkRejectsInvalidSemanticInput(t *testing.T) {
	id := strings.Repeat("a", 64)
	source := []string{"--host", "codex", "--native-session-id", "session-one"}
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"mixed selector", []string{"open", "--work-id", id, "--key", "other", "--change", "."}},
		{"unknown state", []string{"stage", "--work-id", id, "--stage", "build", "--state", "approved", "--reason", "test"}},
		{"incomplete block", []string{"task", "update", "--work-id", id, "--key", "T1", "--stage", "build", "--state", "blocked", "--reason", "test"}},
		{"closure reason", []string{"close", "--work-id", id, "--outcome", "completed"}},
		{"internal version", []string{"resume", "--work-id", id, "--expected-version", "1"}},
		{"missing native source", []string{"resume", "--work-id", id, "--host", "codex"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"work"}, tc.args...)
			if tc.name != "missing native source" {
				args = append(args, source...)
			}
			var out bytes.Buffer
			err := runView(args, &out, realViewDeps())
			if err == nil || !strings.HasPrefix(err.Error(), "invalid_input:") {
				t.Fatalf("expected offline semantic rejection, got %v", err)
			}
			if out.Len() != 0 {
				t.Fatal("invalid input emitted success")
			}
		})
	}
}

func TestViewWorkTaskJSONIsClosedAndAtomic(t *testing.T) {
	id := strings.Repeat("a", 64)
	task := `{"key":"T1","title":"Build export","summary":"Required behavior","order":1,"segments":[{"stage":"build","order":1}]}`
	args := []string{"tasks", "register", "--work-id", id, "--host", "codex", "--native-session-id", "session-one", "--plan-revision", "revision-one", "--json", "-"}
	for _, tc := range []struct {
		name, raw string
		valid     bool
	}{
		{"valid", `[` + task + `]`, true},
		{"duplicate key", `[` + task + `,` + task + `]`, false},
		{"duplicate JSON field", `[{"key":"T2",` + task[1:] + `]`, false},
		{"internal operation", `[` + strings.Replace(task, `"order":1,`, `"operation_id":"injected","order":1,`, 1) + `]`, false},
		{"internal assignee", `[` + strings.Replace(task, `"stage":"build"`, `"assigned_producer_id":"invented","stage":"build"`, 1) + `]`, false},
		{"unknown field", `[` + strings.Replace(task, `"key":"T1"`, `"Key":"T1"`, 1) + `]`, false},
		{"missing required summary", `[` + strings.Replace(task, `"summary":"Required behavior",`, "", 1) + `]`, false},
		{"invalid stage", `[` + strings.Replace(task, `"build"`, `"deploy"`, 1) + `]`, false},
		{"duplicate segment", `[` + strings.Replace(task, `"segments":[{"stage":"build","order":1}]`, `"segments":[{"stage":"build","order":1},{"stage":"build","order":2}]`, 1) + `]`, false},
		{"empty", `[]`, false},
		{"null", `null`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := parseViewWorkCommand(args, strings.NewReader(tc.raw))
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v got %v", tc.valid, err)
			}
			if tc.valid && (len(c.Tasks) != 1 || c.Tasks[0].Key != "T1") {
				t.Fatal("semantic task lost")
			}
		})
	}
}

func TestViewWorkSemanticModes(t *testing.T) {
	id := strings.Repeat("a", 64)
	source := []string{"--host", "codex", "--native-session-id", "session-one"}
	for _, args := range [][]string{
		{"open", "--project", ".", "--change", ".", "--key", "ticket-10", "--title", "Export", "--summary", "Implement export"},
		{"publish", "--work-id", id, "--change", ".", "--stage", "plan"},
		{"publish", "--work-id", id, "--change", ".", "--stage", "research", "--document-key", "research", "--document", "notes.md"},
		{"stage", "--work-id", id, "--stage", "build", "--state", "in_review", "--reason", "Review candidate"},
		{"task", "update", "--work-id", id, "--key", "T1", "--stage", "build", "--state", "blocked", "--reason", "Blocked", "--block-reason", "Missing permission", "--block-owner", "Maintainer", "--block-step", "Grant access", "--resume-when", "Access verified"},
		{"reconcile", "--work-id", id, "--pending", "intent-one", "--action", "archive-rejected", "--reason", "Confirmed rejection"},
		{"close", "--work-id", id, "--outcome", "completed", "--reason", "Integrated and verified"},
	} {
		c, err := parseViewWorkCommand(append(args, source...), strings.NewReader(""))
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if c.Action == "publish" && c.Stage == "plan" && (c.DocumentKey != "plan" || c.Status != "draft") {
			t.Fatal("plan defaults missing")
		}
	}
	for _, raw := range []string{`{"title":"New title"}`, `{"stage":"build","segment_order":2,"plan_revision":"revision-one"}`} {
		args := append([]string{"task", "update", "--work-id", id, "--key", "T1", "--reason", "Update metadata", "--json", "-"}, source...)
		if _, err := parseViewWorkCommand(args, strings.NewReader(raw)); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--state", "completed")
		if _, err := parseViewWorkCommand(args, strings.NewReader(raw)); err == nil {
			t.Fatal("mixed metadata/transition accepted")
		}
	}
}

func TestViewWorkMutationOutputBound(t *testing.T) {
	var out bytes.Buffer
	if err := writeViewWorkResult(&out, viewWorkResult{WorkID: strings.Repeat("a", 64), Result: "confirmed", SessionVersion: 3, RevisionID: "revision-one"}); err != nil {
		t.Fatal(err)
	}
	if out.Len() > 4096 || strings.Contains(out.String(), "document body") {
		t.Fatal("unbounded result")
	}
	out.Reset()
	if err := writeViewWorkResult(&out, viewWorkResult{WorkID: strings.Repeat("a", 64), Result: strings.Repeat("x", 4096), SessionVersion: 3}); err == nil || out.Len() != 0 {
		t.Fatal("oversize success emitted")
	}
}

// The error of a rejected --json names the path and the rule, never a value.
func TestViewWorkJSONErrorsNameThePathAndTheRule(t *testing.T) {
	id := strings.Repeat("a", 64)
	source := []string{"--host", "codex", "--native-session-id", "session-one"}
	task := func(key string, order int, stage string, segOrder int) string {
		return fmt.Sprintf(`{"key":%q,"title":"Build","summary":"Behavior","order":%d,"segments":[{"stage":%q,"order":%d}]}`, key, order, stage, segOrder)
	}
	register := append([]string{"tasks", "register", "--work-id", id, "--plan-revision", "revision-one", "--json", "-"}, source...)
	update := append([]string{"task", "update", "--work-id", id, "--key", "T1", "--reason", "Update", "--json", "-"}, source...)
	for _, tc := range []struct {
		name string
		args []string
		raw  string
		want string
	}{
		{"duplicate key", register, "[" + task("T1", 1, "build", 1) + "," + task("T1", 2, "build", 2) + "]", "tasks[1].key: duplicate task key"},
		{"duplicate plan order", register, "[" + task("T1", 1, "build", 1) + "," + task("T2", 1, "build", 2) + "]", "tasks[1].order: duplicate plan order"},
		{"segment order repeated across tasks", register, "[" + task("T1", 1, "build", 1) + "," + task("T2", 2, "build", 1) + "]", "tasks[1].segments[0].order: duplicate stage order"},
		{"T0 key", register, "[" + task("T0", 1, "build", 1) + "]", "tasks[0].key: invalid key"},
		{"unknown stage", register, "[" + task("T1", 1, "deploy", 1) + "]", "tasks[0].segments[0].stage: invalid stage"},
		{"zero segment order", register, "[" + task("T1", 1, "build", 0) + "]", "tasks[0].segments[0].order: invalid positive integer"},
		{"empty list", register, "[]", "tasks: invalid task count"},
		{"unknown task field", register, `[` + strings.Replace(task("T1", 1, "build", 1), `"order":1,`, `"order":1,"extra":"LEAKED-VALUE",`, 1) + `]`, "tasks[0].extra: unknown field"},
		{"unknown segment field", register, `[` + strings.Replace(task("T1", 1, "build", 1), `"stage":"build"`, `"stage":"build","extra":"LEAKED-VALUE"`, 1) + `]`, "tasks[0].segments[0].extra: unknown field"},
		{"missing summary", register, `[` + strings.Replace(task("T1", 1, "build", 1), `"summary":"Behavior",`, "", 1) + `]`, "tasks[0].summary: required field"},
		{"wrong type", register, `[` + strings.Replace(task("T1", 1, "build", 1), `"order":1`, `"order":"LEAKED-VALUE"`, 1) + `]`, "tasks[0].order: wrong type"},
		{"null field", register, `[` + strings.Replace(task("T1", 1, "build", 1), `"summary":"Behavior"`, `"summary":null`, 1) + `]`, "tasks[0].summary: null not allowed"},
		{"second task wrong type", register, "[" + task("T1", 1, "build", 1) + "," + strings.Replace(task("T2", 2, "build", 2), `"order":2`, `"order":true`, 1) + "]", "tasks[1].order: wrong type"},
		{"duplicate JSON field", register, `[{"key":"T2",` + task("T1", 1, "build", 1)[1:] + `]`, "body: duplicate key"},
		{"metadata unknown field", update, `{"bogus":"LEAKED-VALUE"}`, "metadata.bogus: unknown field"},
		{"metadata wrong type", update, `{"title":5}`, "metadata.title: wrong type"},
		{"metadata invalid stage", update, `{"stage":"deploy","plan_revision":"revision-one"}`, "metadata.stage: invalid stage"},
		{"metadata order without reference", update, `{"order":2}`, "metadata.order: plan order update requires confirmed plan or decision reference"},
		{"metadata empty title", update, `{"title":"  "}`, "metadata.title: invalid text"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseViewWorkCommand(tc.args, strings.NewReader(tc.raw))
			if err == nil || !strings.HasPrefix(err.Error(), "invalid_input: invalid semantic JSON: ") || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %v, want invalid_input naming %q", err, tc.want)
			}
			if strings.Contains(err.Error(), "LEAKED-VALUE") {
				t.Fatalf("the error repeats a received value: %v", err)
			}
			if !strings.Contains(err.Error(), "hive view work contract") {
				t.Fatalf("the error lost the contract suggestion: %v", err)
			}
		})
	}
}

func TestViewWorkContractDocumentsThePathConventionAndTwoRules(t *testing.T) {
	for _, want := range []string{"tasks[<i>].<field>", "metadata.<field>", "T0 is rejected", "unique per stage across all tasks"} {
		if !strings.Contains(viewWorkContract, want) {
			t.Errorf("contract lacks %q", want)
		}
	}
}
