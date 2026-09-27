package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- S1: question_after_detail ---

func TestQuestionAfterDetail(t *testing.T) {
	yes := true
	for _, tc := range []struct {
		name   string
		events []traceEvent
		want   string
	}{
		{
			name: "sufficient-spanish-detail-passes",
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Message: "m1", Text: "Reviso el detalle suficiente antes de preguntar, con más de cuarenta caracteres en total aquí."},
				{Kind: "question", Message: "m1"},
			},
			want: "pass",
		},
		{
			name: "below-threshold-fails",
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Message: "m1", Text: strings.Repeat("a", 39)},
				{Kind: "question", Message: "m1"},
			},
			want: "fail",
		},
		{
			name: "exactly-forty-runes-passes",
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Message: "m1", Text: strings.Repeat("á", 40)},
				{Kind: "question", Message: "m1"},
			},
			want: "pass",
		},
		{
			name: "whitespace-does-not-count-toward-threshold",
			// 39 non-whitespace runes padded with spaces: a naive len(string)
			// count (78) would clear 40, but the actual rune count must not.
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Message: "m1", Text: strings.Repeat("a ", 39)},
				{Kind: "question", Message: "m1"},
			},
			want: "fail",
		},
		{
			name: "parallel-sibling-result-same-message-does-not-cut-window",
			events: []traceEvent{
				{Kind: "tool_result", Message: "m1", Success: &yes},
				{Kind: "text", Role: "assistant", Message: "m1", Text: strings.Repeat("b", 40)},
				{Kind: "question", Message: "m1"},
			},
			want: "pass",
		},
		{
			name: "sibling-result-arrives-before-question-still-in-window",
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Message: "m1", Text: strings.Repeat("c", 20)},
				{Kind: "tool_result", Message: "m1", Success: &yes},
				{Kind: "text", Role: "assistant", Message: "m1", Text: strings.Repeat("d", 20)},
				{Kind: "question", Message: "m1"},
			},
			want: "pass",
		},
		{
			name: "different-message-tool-result-cuts-the-window",
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Message: "m0", Text: strings.Repeat("e", 40)},
				{Kind: "tool_result", Message: "m0", Success: &yes},
				{Kind: "question", Message: "m1"},
			},
			want: "fail",
		},
		{
			name: "question-without-any-prior-tool-result",
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Message: "m1", Text: strings.Repeat("f", 40)},
				{Kind: "question", Message: "m1"},
			},
			want: "pass",
		},
		{
			name: "non-assistant-role-text-does-not-count",
			events: []traceEvent{
				{Kind: "text", Role: "user", Message: "m1", Text: strings.Repeat("g", 40)},
				{Kind: "question", Message: "m1"},
			},
			want: "fail",
		},
		{
			name: "pi-toolresult-role-text-does-not-count",
			events: []traceEvent{
				{Kind: "text", Role: "toolResult", Message: "m1", Text: strings.Repeat("h", 40)},
				{Kind: "question", Message: "m1"},
			},
			want: "fail",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Trace: traceReport{Events: tc.events}}
			if got := questionAfterDetail(r).Status; got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

// TestOpenCodeQuestionAfterDetail proves S1 evaluates OpenCode traces
// correctly now that its text and question events carry Role and Message
// (work-close-sequence T2): a question preceded by >=40 non-whitespace
// runes of assistant text, in the same or a different message, passes; one
// without that detail fails.
func TestOpenCodeQuestionAfterDetail(t *testing.T) {
	for _, tc := range []struct {
		name  string
		trace string
		want  string
	}{
		{
			name: "detail-in-same-message-passes",
			trace: `{"type":"text","sessionID":"ses-1","part":{"type":"text","messageID":"msg-1","text":"Reviso el detalle suficiente antes de preguntar, con más de cuarenta caracteres en total aquí."}}
{"type":"tool_use","part":{"type":"tool","tool":"question","callID":"q1","messageID":"msg-1","state":{"status":"completed","input":{},"output":"ok"}}}`,
			want: "pass",
		},
		{
			name: "detail-in-an-earlier-message-passes",
			trace: `{"type":"text","sessionID":"ses-1","part":{"type":"text","messageID":"msg-1","text":"Reviso el detalle suficiente antes de preguntar, con más de cuarenta caracteres en total aquí."}}
{"type":"tool_use","part":{"type":"tool","tool":"question","callID":"q1","messageID":"msg-2","state":{"status":"completed","input":{},"output":"ok"}}}`,
			want: "pass",
		},
		{
			name:  "without-detail-fails",
			trace: `{"type":"tool_use","part":{"type":"tool","tool":"question","callID":"q1","messageID":"msg-1","state":{"status":"completed","input":{},"output":"ok"}}}`,
			want:  "fail",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			trace := parseTrace("opencode", strings.NewReader(tc.trace))
			r := result{Trace: trace}
			if got := questionAfterDetail(r).Status; got != tc.want {
				t.Fatalf("got %s want %s: %+v", got, tc.want, trace.Events)
			}
		})
	}
}

func TestQuestionAfterDetailNotObservedWithoutQuestions(t *testing.T) {
	r := result{Trace: traceReport{Events: []traceEvent{{Kind: "text", Role: "assistant", Text: "hola"}}}}
	if got := questionAfterDetail(r).Status; got != "not_observed" {
		t.Fatalf("got %s", got)
	}
}

func TestQuestionAfterDetailEvidenceNeverIncludesText(t *testing.T) {
	r := result{Trace: traceReport{Events: []traceEvent{
		{Line: 12, Kind: "question", Tool: "AskUserQuestion", Path: "", Message: "m1"},
	}}}
	a := questionAfterDetail(r)
	if a.Status != "fail" {
		t.Fatalf("expected fail, got %+v", a)
	}
	for _, e := range a.Evidence {
		if strings.Contains(e, "AskUserQuestion") == false {
			t.Fatalf("evidence missing tool name: %q", e)
		}
		if !strings.HasPrefix(e, "stdout.jsonl:12 question") {
			t.Fatalf("evidence format changed: %q", e)
		}
	}
}

// --- S3: no_broad_git_add ---

func TestNoBroadGitAdd(t *testing.T) {
	yes := true
	base := func(command string) result {
		return result{Root: "/repo", Cwd: "/repo", Trace: traceReport{Events: []traceEvent{
			{Kind: "shell", Command: command, Success: &yes},
		}}}
	}
	for _, tc := range []struct {
		name string
		r    result
		want string
	}{
		{"dash-C-with-dash-A", base("git -C x add -A"), "fail"},
		{"dash-c-config-then-dot", base("git -c k=v add ."), "fail"},
		{"trailing-slash-directory", base("git add dir/"), "fail"},
		{"directory-without-trailing-slash", func() result {
			r := base("git add somedir")
			r.After = map[string]item{"somedir/file.txt": {}}
			return r
		}(), "fail"},
		{"literal-files-pass", base("git add a.go b.go"), "pass"},
		{"git-mv-is-not-add", base("git mv a.go b.go"), "not_observed"},
		{"chained-with-ampersand", base("cd /repo && git add -A"), "fail"},
		{"piped-command-still-inspected", base("git status | cat && git add -A"), "fail"},
		{"wrapped-in-shell-dash-lc", base(`/bin/zsh -lc "git add -A"`), "fail"},
		{"colon-slash-pathspec", base("git add :/"), "fail"},
		{"dotdot-resolving-to-root", func() result {
			r := base("git add ..")
			r.Cwd = "/repo/sub"
			return r
		}(), "fail"},
		{"absolute-root-path", base("git add /repo"), "fail"},
		{"cd-then-directory-add", func() result {
			r := base("cd frontend && git add src")
			r.After = map[string]item{"frontend/src/app.go": {}}
			return r
		}(), "fail"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := noBroadGitAdd(tc.r).Status; got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func TestNoBroadGitAddEvidenceNeverIncludesCommandText(t *testing.T) {
	yes := true
	r := result{Root: "/repo", Cwd: "/repo", Trace: traceReport{Events: []traceEvent{
		{Line: 3, Kind: "shell", Tool: "shell", Command: "git add -A", Success: &yes},
	}}}
	a := noBroadGitAdd(r)
	if a.Status != "fail" {
		t.Fatalf("expected fail, got %+v", a)
	}
	for _, e := range a.Evidence {
		if strings.Contains(e, "git add") {
			t.Fatalf("evidence leaked command text: %q", e)
		}
	}
}

// --- S5: no_secret_content_read ---

func TestNoSecretContentRead(t *testing.T) {
	yes, no := true, false
	shellEvent := func(command string, success bool) traceEvent {
		s := success
		return traceEvent{Kind: "shell", Command: command, Success: &s}
	}
	for _, tc := range []struct {
		name  string
		event traceEvent
		want  string
	}{
		{"env-example-is-template-not-secret", shellEvent("cat .env.example", true), "pass"},
		{"env-dist-is-template-not-secret", shellEvent("cat .env.dist", true), "pass"},
		{"env-test-is-secret", shellEvent("cat .env.test", true), "fail"},
		{"rg-dash-l-safe", shellEvent("rg -l KEY .env", true), "pass"},
		{"grep-dash-c-safe", shellEvent("grep -c KEY .env", true), "pass"},
		{"rg-dash-L-is-follow-not-safe", shellEvent("rg -L KEY .env", true), "fail"},
		{"grep-dash-o-prints-content", shellEvent("grep -o KEY .env", true), "fail"},
		{"grouped-dash-rl-is-safe", shellEvent("grep -rl KEY .env", true), "pass"},
		{"grouped-dash-nc-is-safe", shellEvent("grep -nc KEY .env", true), "pass"},
		{"cat-piped-to-head-still-fails", shellEvent("cat .env | head", true), "fail"},
		{"rg-glob-env-star-fails", shellEvent(`rg -g '.env*' KEY`, true), "fail"},
		{"failed-read-does-not-count", shellEvent("cat .env", false), "not_observed"},
		{"awk-is-not-judged", shellEvent("awk -F= '{print $2}' .env", true), "not_observed"},
		{"rg-non-secret-target-passes", shellEvent("rg -n foo src/", true), "pass"},
		{"cat-non-secret-passes", shellEvent("cat README.md", true), "pass"},
		{"redirect-into-env-is-a-write", shellEvent("cat .env.example > .env", true), "pass"},
		{"heredoc-into-env-is-a-write", shellEvent("cat > .env <<'EOF'", true), "pass"},
		{"sed-in-place-prints-nothing", shellEvent("sed -i 's/a/b/' .env", true), "pass"},
		{"rg-files-lists-paths-only", shellEvent("rg --files -g '.env*'", true), "pass"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Trace: traceReport{Events: []traceEvent{tc.event}}}
			if got := noSecretContentRead(r).Status; got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
	_ = yes
	_ = no
}

func TestNoSecretContentReadNativeReadAndSearch(t *testing.T) {
	yes := true
	for _, tc := range []struct {
		name  string
		event traceEvent
		want  string
	}{
		{"native-read-of-env", traceEvent{Kind: "read", Path: ".env", Success: &yes}, "fail"},
		{"native-read-of-example", traceEvent{Kind: "read", Path: ".env.example", Success: &yes}, "pass"},
		{"claude-grep-files-with-matches-is-safe", traceEvent{Kind: "search", Tool: "Grep", Path: ".env", Input: []byte(`{"path":".env","output_mode":"files_with_matches"}`), Success: &yes}, "pass"},
		{"claude-grep-content-is-not-safe", traceEvent{Kind: "search", Tool: "Grep", Path: ".env", Input: []byte(`{"path":".env","output_mode":"content"}`), Success: &yes}, "fail"},
		{"pi-grep-has-no-safe-mode", traceEvent{Kind: "search", Tool: "grep", Path: ".env", Input: []byte(`{"path":".env"}`), Success: &yes}, "fail"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Trace: traceReport{Events: []traceEvent{tc.event}}}
			if got := noSecretContentRead(r).Status; got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func TestNoSecretContentReadEvidenceNeverIncludesCommandOrResultText(t *testing.T) {
	yes := true
	r := result{Trace: traceReport{Events: []traceEvent{
		{Line: 7, Kind: "shell", Tool: "shell", Command: "cat .env", Text: "SECRET_KEY=example-not-a-secret", Success: &yes},
	}}}
	a := noSecretContentRead(r)
	if a.Status != "fail" {
		t.Fatalf("expected fail, got %+v", a)
	}
	for _, e := range a.Evidence {
		if strings.Contains(e, "cat .env") || strings.Contains(e, "SECRET_KEY") {
			t.Fatalf("evidence leaked command or result text: %q", e)
		}
	}
}

// --- close_question_after_report ---

func TestCloseQuestionAfterReport(t *testing.T) {
	for _, tc := range []struct {
		name      string
		events    []traceEvent
		completed bool
		want      string
	}{
		{
			name:      "not-completed-is-not-observed-even-without-a-question",
			completed: false,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "Reporte sin pregunta de cierre."},
			},
			want: "not_observed",
		},
		{
			name:      "no-assistant-text-is-not-observed",
			completed: true,
			events:    []traceEvent{{Kind: "tool_result"}},
			want:      "not_observed",
		},
		{
			name:      "question-event-after-report-passes",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "Reporte: entrega completa."},
				{Kind: "question", Tool: "AskUserQuestion"},
			},
			want: "pass",
		},
		{
			name:      "report-ending-in-question-mark-passes-as-text-fallback",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "Reporte: entrega completa.\n¿Seguimos ahora o cerramos la sesión?"},
			},
			want: "pass",
		},
		{
			name:      "trailing-whitespace-after-question-mark-still-passes",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "¿Seguimos?   \n"},
			},
			want: "pass",
		},
		{
			name:      "report-without-question-or-trailing-mark-fails",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "Reporte: entrega completa. Limpieza aplicada."},
			},
			want: "fail",
		},
		{
			name:      "non-question-event-after-report-fails",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "Reporte: entrega completa."},
				{Kind: "tool_result"},
			},
			want: "fail",
		},
		{
			name:      "final-kind-is-excluded-both-as-source-and-as-trailing-evidence",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "Reporte: entrega completa."},
				{Kind: "final", Text: "¿Seguimos ahora?"},
			},
			want: "fail",
		},
		{
			name:      "user-role-text-is-excluded",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "user", Text: "¿Seguimos ahora con lo siguiente?"},
			},
			want: "not_observed",
		},
		{
			name:      "only-the-last-assistant-text-is-considered",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "¿Primera pregunta retórica?"},
				{Kind: "tool_result"},
				{Kind: "text", Role: "assistant", Text: "Reporte final sin pregunta."},
			},
			want: "fail",
		},
		{
			name:      "question-directly-after-last-text-passes-even-with-a-later-final-event",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "Reporte: entrega completa."},
				{Kind: "question", Tool: "AskUserQuestion"},
				{Kind: "final", Text: "Reporte: entrega completa."},
			},
			want: "pass",
		},
		{
			// D17-A false-negative fix: a headless host returns the native
			// question without an answer and the model keeps writing, so
			// the last assistant text trails the question instead of
			// following it directly. As long as nothing but the question's
			// own tool_result and assistant text follow it, this must still
			// pass with "tool" evidence (the question event itself, not the
			// trailing text) — but only once finding 7's detail precondition
			// is also met: the preceding text here is a full report-length
			// paragraph, not a one-line stub, so the question genuinely
			// follows a report rather than preceding one.
			name:      "unanswered-native-question-followed-only-by-its-result-and-more-text-passes",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "Antes de cerrar reviso cada paso pendiente y confirmo que el estado del despliegue quedo correcto."},
				{Kind: "question", Tool: "ask_user_question", ID: "q1", Message: "m1"},
				{Kind: "tool_result", ID: "q1", Message: "m1"},
				{Kind: "text", Role: "assistant", Text: "Como no llego respuesta, dejo el resumen sin pregunta al final."},
			},
			want: "pass",
		},
		{
			// A sibling call issued in the same message as the question (for
			// example a todo-list update alongside ask_user_question)
			// resolves in the same tool_result batch and is not "further
			// work": its result shares the question's Message key even
			// though it carries a different call ID. The report text
			// precedes the todo_write/question pair in the same message, so
			// finding 7's detail precondition is met from that same report.
			name:      "sibling-tool-result-sharing-the-question-message-does-not-count-as-further-work",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Message: "m1", Text: "Antes de cerrar reviso cada paso pendiente y confirmo que el estado del despliegue quedo correcto."},
				{Kind: "tool", Tool: "todo_write", ID: "t1", Message: "m1"},
				{Kind: "question", Tool: "ask_user_question", ID: "q1", Message: "m1"},
				{Kind: "tool_result", ID: "t1", Message: "m1"},
				{Kind: "tool_result", ID: "q1", Message: "m1"},
				{Kind: "text", Role: "assistant", Text: "Resumen final sin pregunta explicita."},
			},
			want: "pass",
		},
		{
			// G3: the question was about scope, early in the run, and real
			// tool calls (work) ran after it before the eventual close
			// report. The unanswered-question shortcut must not fire here,
			// and the ordinary fallback still fails since the final report
			// has no question of its own.
			name:      "native-question-followed-by-a-further-tool-call-then-text-still-fails",
			completed: true,
			events: []traceEvent{
				{Kind: "question", Tool: "ask_user_question", ID: "q1", Message: "m1"},
				{Kind: "tool_result", ID: "q1", Message: "m1"},
				{Kind: "shell", Command: "echo status"},
				{Kind: "text", Role: "assistant", Text: "Reporte: el despliegue quedo completo. Limpieza aplicada."},
			},
			want: "fail",
		},
		{
			// Finding 7: the unanswered-question shortcut must require the
			// report before the question, not merely nothing but text after
			// it. Here the native question fires immediately after a tool
			// call with no preceding report/detail of its own, and the run's
			// real, substantive report is written only afterward, as the
			// trailing text — the inverse of a legitimate close attempt. The
			// shortcut must be denied for insufficient preceding detail, and
			// the ordinary fallback then correctly fails since that trailing
			// report never asks its own close question.
			name:      "mid-flow question then report, no close question",
			completed: true,
			events: []traceEvent{
				{Kind: "shell", Command: "echo status", Message: "m0"},
				{Kind: "tool_result", Message: "m0"},
				{Kind: "question", Tool: "ask_user_question", ID: "q1", Message: "m1"},
				{Kind: "tool_result", ID: "q1", Message: "m1"},
				{Kind: "text", Role: "assistant", Text: "Reporte: la tarea quedo completa tras revisar cada paso pendiente y aplicar los cambios necesarios en el modulo correspondiente."},
			},
			want: "fail",
		},
		{
			// The trailing "final" event that Claude/Grok's own wire format
			// always appends after a completed turn duplicates the last
			// assistant text; it is never itself further work, so it must
			// not block the unanswered-question shortcut. The report
			// precedes the question, meeting finding 7's detail precondition.
			name:      "trailing-final-event-after-question-and-text-still-passes",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Message: "m1", Text: "Antes de cerrar reviso cada paso pendiente y confirmo que el estado del despliegue quedo correcto."},
				{Kind: "question", Tool: "ask_user_question", ID: "q1", Message: "m1"},
				{Kind: "tool_result", ID: "q1", Message: "m1"},
				{Kind: "text", Role: "assistant", Text: "Resumen final sin pregunta explicita."},
				{Kind: "final", Text: "Resumen final sin pregunta explicita."},
			},
			want: "pass",
		},
		{
			// D17-A false-negative fix: the text fallback now looks for "?"
			// anywhere in the report's last paragraph, not only at its very
			// end, so a mid-paragraph Spanish "¿...?" still passes even when
			// more text follows it.
			name:      "question-mark-mid-paragraph-passes-as-text-fallback",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "Es lo que recomiendo, ¿seguimos con el siguiente paso o prefieres revisarlo antes? Quedo atento a cualquiera de las dos rutas."},
			},
			want: "pass",
		},
		{
			// Only the last paragraph (the text after the last blank line)
			// is checked: an earlier rhetorical question in a prior
			// paragraph must not paper over a final paragraph with no
			// question at all.
			name:      "question-mark-in-an-earlier-paragraph-does-not-count",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "¿Repaso el alcance antes de seguir?\n\nReporte final: entrega completa, sin pregunta en este parrafo."},
			},
			want: "fail",
		},
		{
			// An offer to continue with no question mark anywhere is still a
			// fail: the fallback requires an actual "?", not just an
			// invitation to keep going.
			name:      "offer-to-continue-without-a-question-mark-fails",
			completed: true,
			events: []traceEvent{
				{Kind: "text", Role: "assistant", Text: "Reporte: la tarea quedo completa y la limpieza se aplico. Quedo disponible para continuar con lo que sigue cuando lo definas."},
			},
			want: "fail",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Trace: traceReport{Events: tc.events}}
			if got := closeQuestionAfterReport(r, tc.completed).Status; got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func TestCloseQuestionAfterReportEvidenceNeverIncludesText(t *testing.T) {
	r := result{Trace: traceReport{Events: []traceEvent{
		{Line: 9, Kind: "text", Role: "assistant", Text: "Reporte: entrega completa. Limpieza aplicada. Sin pregunta al final."},
	}}}
	a := closeQuestionAfterReport(r, true)
	if a.Status != "fail" {
		t.Fatalf("expected fail, got %+v", a)
	}
	for _, e := range a.Evidence {
		if strings.Contains(e, "Reporte") || strings.Contains(e, "Limpieza") {
			t.Fatalf("evidence leaked report text: %q", e)
		}
		if !strings.HasPrefix(e, "stdout.jsonl:9 text") {
			t.Fatalf("evidence format changed: %q", e)
		}
	}
}

// --- merged_branch_deleted ---

func TestMergedBranchDeleted(t *testing.T) {
	yes := true
	shell := func(command string) traceEvent { return traceEvent{Kind: "shell", Command: command, Success: &yes} }
	failedShell := func(command string) traceEvent {
		no := false
		return traceEvent{Kind: "shell", Command: command, Success: &no}
	}
	for _, tc := range []struct {
		name   string
		events []traceEvent
		base   string
		want   string
	}{
		{"no-merge-is-not-observed", []traceEvent{shell("git status")}, "main", "not_observed"},
		{"gh-pr-merge-without-delete-flag-fails", []traceEvent{shell("gh pr merge 165 --merge")}, "main", "fail"},
		{"gh-pr-merge-delete-branch-false-fails", []traceEvent{shell("gh pr merge 165 --merge --delete-branch=false")}, "main", "fail"},
		{"gh-pr-merge-delete-branch-passes", []traceEvent{shell("gh pr merge 165 --merge --delete-branch")}, "main", "pass"},
		{"gh-pr-merge-short-dash-d-passes", []traceEvent{shell("gh pr merge 165 -d")}, "main", "pass"},
		{"git-merge-then-push-to-base-then-delete-passes", []traceEvent{
			shell("git merge fix/close-sequence"),
			shell("git push origin main"),
			shell("git push origin --delete fix/close-sequence"),
		}, "main", "pass"},
		{"git-merge-then-push-to-base-without-delete-fails", []traceEvent{
			shell("git merge fix/close-sequence"),
			shell("git push origin main"),
		}, "main", "fail"},
		{"git-merge-without-a-push-to-base-is-not-observed", []traceEvent{
			shell("git merge fix/close-sequence"),
		}, "main", "not_observed"},
		{"git-branch-delete-after-gh-merge-passes", []traceEvent{
			shell("gh pr merge 165 --merge"),
			shell("git branch -d fix/close-sequence"),
		}, "main", "pass"},
		{"colon-refspec-delete-passes", []traceEvent{
			shell("gh pr merge 165 --merge"),
			shell("git push origin :fix/close-sequence"),
		}, "main", "pass"},
		{"gh-api-delete-ref-passes", []traceEvent{
			shell("gh pr merge 165 --merge"),
			shell("gh api -X DELETE repos/org/repo/git/refs/heads/fix/close-sequence"),
		}, "main", "pass"},
		{"deletion-before-merge-does-not-count", []traceEvent{
			shell("git branch -d old-branch"),
			shell("gh pr merge 165 --merge"),
		}, "main", "fail"},
		{"failed-shell-event-is-ignored", []traceEvent{
			failedShell("gh pr merge 165 --merge --delete-branch"),
		}, "main", "not_observed"},
		{"wrapped-in-shell-dash-lc-still-detected", []traceEvent{
			shell(`/bin/zsh -lc "gh pr merge 165 --merge --delete-branch"`),
		}, "main", "pass"},
		{"push-to-an-unrelated-branch-does-not-complete-the-merge-until-base", []traceEvent{
			shell("git merge fix/close-sequence"),
			shell("git push origin release"),
			shell("git push origin main"),
		}, "main", "fail"},
		{"deletion-between-local-merge-and-its-confirming-push-still-counts", []traceEvent{
			shell("git merge fix/close-sequence"),
			shell("git branch -d fix/close-sequence"),
			shell("git push origin main"),
		}, "main", "pass"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Root: "/repo", Cwd: "/repo", Trace: traceReport{Events: tc.events}}
			if got := mergedBranchDeleted(r, tc.base).Status; got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func TestMergedBranchDeletedEvidenceNeverIncludesCommandText(t *testing.T) {
	yes := true
	r := result{Root: "/repo", Cwd: "/repo", Trace: traceReport{Events: []traceEvent{
		{Line: 4, Kind: "shell", Tool: "shell", Command: "gh pr merge 165 --merge --delete-branch=false", Success: &yes},
	}}}
	a := mergedBranchDeleted(r, "main")
	if a.Status != "fail" {
		t.Fatalf("expected fail, got %+v", a)
	}
	for _, e := range a.Evidence {
		if strings.Contains(e, "gh pr merge") || strings.Contains(e, "delete-branch") {
			t.Fatalf("evidence leaked command text: %q", e)
		}
	}
}

// --- T3: fixtures with provenance, parsed end-to-end with the real host parser ---

func TestRegressionFixtures(t *testing.T) {
	type tc struct {
		criterion string
		host      string
		file      string
		want      string
		check     func(result) criterionAssessment
	}
	root := "../fixtures/regression"
	for _, c := range []tc{
		{"question_after_detail", "claude", "claude-fail.jsonl", "fail", questionAfterDetail},
		{"question_after_detail", "claude", "claude-pass.jsonl", "pass", questionAfterDetail},
		{"question_after_detail", "claude", "claude-usertext-fail.jsonl", "fail", questionAfterDetail},
		{"question_after_detail", "grok", "grok-fail.jsonl", "fail", questionAfterDetail},
		{"question_after_detail", "grok", "grok-pass.jsonl", "pass", questionAfterDetail},
		{"question_after_detail", "pi", "pi-fail.jsonl", "fail", questionAfterDetail},
		{"question_after_detail", "pi", "pi-pass.jsonl", "pass", questionAfterDetail},
		{"question_after_detail", "pi", "pi-toolresult-fail.jsonl", "fail", questionAfterDetail},
		{"no_broad_git_add", "grok", "grok-fail.jsonl", "fail", noBroadGitAdd},
		{"no_broad_git_add", "grok", "grok-pass.jsonl", "pass", noBroadGitAdd},
		{"no_secret_content_read", "grok", "grok-fail.jsonl", "fail", noSecretContentRead},
		{"no_secret_content_read", "grok", "grok-pass.jsonl", "pass", noSecretContentRead},
		{"close_question_after_report", "codex", "codex-fail.jsonl", "fail", func(r result) criterionAssessment { return closeQuestionAfterReport(r, true) }},
		{"close_question_after_report", "codex", "codex-pass.jsonl", "pass", func(r result) criterionAssessment { return closeQuestionAfterReport(r, true) }},
		{"close_question_after_report", "grok", "grok-fail.jsonl", "fail", func(r result) criterionAssessment { return closeQuestionAfterReport(r, true) }},
		{"close_question_after_report", "grok", "grok-pass.jsonl", "pass", func(r result) criterionAssessment { return closeQuestionAfterReport(r, true) }},
		{"close_question_after_report", "grok", "grok-text-fallback-pass.jsonl", "pass", func(r result) criterionAssessment { return closeQuestionAfterReport(r, true) }},
		{"close_question_after_report", "grok", "grok-tool-then-text-fail.jsonl", "fail", func(r result) criterionAssessment { return closeQuestionAfterReport(r, true) }},
		{"close_question_after_report", "grok", "grok-text-midline-pass.jsonl", "pass", func(r result) criterionAssessment { return closeQuestionAfterReport(r, true) }},
		{"close_question_after_report", "grok", "grok-offer-no-question-fail.jsonl", "fail", func(r result) criterionAssessment { return closeQuestionAfterReport(r, true) }},
		{"merged_branch_deleted", "grok", "grok-fail.jsonl", "fail", func(r result) criterionAssessment { return mergedBranchDeleted(r, "main") }},
		{"merged_branch_deleted", "grok", "grok-pass.jsonl", "pass", func(r result) criterionAssessment { return mergedBranchDeleted(r, "main") }},
		{"merged_branch_deleted", "codex", "codex-fail.jsonl", "fail", func(r result) criterionAssessment { return mergedBranchDeleted(r, "main") }},
		{"merged_branch_deleted", "codex", "codex-pass.jsonl", "pass", func(r result) criterionAssessment { return mergedBranchDeleted(r, "main") }},
		{"cited_id_glossed", "claude", "claude-fail.jsonl", "fail", citedIDGlossed},
		{"cited_id_glossed", "claude", "claude-pass.jsonl", "pass", citedIDGlossed},
		{"cited_id_glossed", "claude", "claude-question-after-text-fail.jsonl", "fail", citedIDGlossed},
		{"cited_id_glossed", "claude", "claude-range-fail.jsonl", "fail", citedIDGlossed},
		{"cited_id_glossed", "claude", "claude-range-pass.jsonl", "pass", citedIDGlossed},
		{"no_bare_url", "claude", "claude-fail.jsonl", "fail", noBareURL},
		{"no_bare_url", "claude", "claude-pass.jsonl", "pass", noBareURL},
		{"ticket_ids_not_packed_in_prose", "grok", "grok-fail.jsonl", "fail", ticketIDsNotPackedInProse},
		{"ticket_ids_not_packed_in_prose", "grok", "grok-pass.jsonl", "pass", ticketIDsNotPackedInProse},
		{"ticket_ids_not_packed_in_prose", "grok", "grok-list-continuation-pass.jsonl", "pass", ticketIDsNotPackedInProse},
		{"ticket_ids_not_packed_in_prose", "grok", "grok-links-dedup-fail.jsonl", "fail", ticketIDsNotPackedInProse},
		{"ticket_ids_not_packed_in_prose", "grok", "grok-heading-label-pass.jsonl", "pass", ticketIDsNotPackedInProse},
		{"ticket_ids_not_packed_in_prose", "grok", "grok-bold-label-pass.jsonl", "pass", ticketIDsNotPackedInProse},
		{"ticket_ids_not_packed_in_prose", "grok", "grok-bold-ids-only-pass.jsonl", "pass", ticketIDsNotPackedInProse},
		{"ticket_ids_not_packed_in_prose", "grok", "grok-bold-label-then-bare-ids-fail.jsonl", "fail", ticketIDsNotPackedInProse},
	} {
		t.Run(c.criterion+"/"+c.file, func(t *testing.T) {
			path := filepath.Join(root, c.criterion, c.file)
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			trace := parseTrace(c.host, f)
			r := result{Root: "/repo", Cwd: "/repo", Trace: trace}
			if got := c.check(r).Status; got != c.want {
				t.Fatalf("%s: got %s want %s", path, got, c.want)
			}
		})
	}
}

// --- regressionCriteria wiring ---

func TestRegressionCriteriaReturnsAllSix(t *testing.T) {
	r := result{Trace: traceReport{}}
	got := regressionCriteria(r)
	if len(got) != 6 {
		t.Fatalf("expected 6 criteria, got %d: %+v", len(got), got)
	}
	names := map[string]bool{}
	for _, c := range got {
		names[c.Criterion] = true
	}
	for _, want := range []string{"question_after_detail", "no_broad_git_add", "no_secret_content_read", "cited_id_glossed", "no_bare_url", "ticket_ids_not_packed_in_prose"} {
		if !names[want] {
			t.Fatalf("missing criterion %s in %+v", want, got)
		}
	}
}

// --- cited_id_glossed (report-readability D1-A) ---

func askQuestion(line int, entries ...askQuestionEntry) traceEvent {
	input, _ := json.Marshal(askQuestionInput{Questions: entries})
	return traceEvent{Line: line, Kind: "question", Tool: "AskUserQuestion", Input: input}
}

func assistantText(line int, message, text string) traceEvent {
	return traceEvent{Line: line, Kind: "text", Role: "assistant", Message: message, Text: text}
}

// defineS1S5 is the pair of table-row definitions ("| S1 | ... |") the unit
// tests below cite later from a different message, matching design.md's own
// worked example.
func defineS1S5(line int, message string) traceEvent {
	return assistantText(line, message, "| ID | Hallazgo |\n|---|---|\n| **S1** | Cierre sin reporte |\n| **S5** | Secreto leído con una herramienta amplia |")
}

func TestCitedIDGlossed(t *testing.T) {
	for _, tc := range []struct {
		name   string
		events []traceEvent
		want   string
	}{
		{
			// design.md's own worked example: "S1–S5 está" (no gloss) fails.
			name: "range-cited-bare-fails",
			events: []traceEvent{
				defineS1S5(1, "m1"),
				assistantText(2, "m2", "El lote S1–S5 está completo y desplegado."),
			},
			want: "fail",
		},
		{
			// "S1–S5 (x)" passes: an immediate "(" glosses the whole range.
			name: "range-cited-glossed-passes",
			events: []traceEvent{
				defineS1S5(1, "m1"),
				assistantText(2, "m2", "El lote S1–S5 (cierre y secretos) está completo."),
			},
			want: "pass",
		},
		{
			// "**D1** (x)" passes: ** is stripped before the ID scan, so the
			// citation reads as "D1 (x)", glossed by the immediate "(".
			name: "bold-marker-stripped-then-glossed-passes",
			events: []traceEvent{
				assistantText(1, "m1", "- **D1**: mover la validación al pipeline de CI."),
				assistantText(2, "m2", "Seguimos con **D1** (x)."),
			},
			want: "pass",
		},
		{
			// "`D1`" is ignored: inline code is masked before the ID scan, so
			// this citation is never found at all — not a fail, not_observed.
			name: "inline-code-citation-ignored",
			events: []traceEvent{
				assistantText(1, "m1", "- **D1**: mover la validación al pipeline de CI."),
				assistantText(2, "m2", "Menciono `D1` sin más contexto."),
			},
			want: "not_observed",
		},
		{
			name: "single-id-defined-as-option-label-cited-bare-in-later-question-fails",
			events: []traceEvent{
				askQuestion(10, askQuestionEntry{Header: "D3 Stack", Question: "D3: ¿cómo seguimos?", Options: []askQuestionOption{
					{Label: "D3-A: piloto en un proyecto (Recomendado)", Description: "El otro proyecto sigue igual."},
					{Label: "D3-B: pausar el otro proyecto", Description: "Pedir que no adopte los cambios nuevos."},
				}}),
				askQuestion(20, askQuestionEntry{Header: "D4", Question: "D4: el piloto terminó. ¿Qué hacemos con D3-A?", Options: []askQuestionOption{
					{Label: "D4-A: cerrar (Recomendado)", Description: "El piloto funcionó."},
				}}),
			},
			want: "fail",
		},
		{
			name: "single-ids-defined-as-option-labels-cited-glossed-in-later-question-passes",
			events: []traceEvent{
				askQuestion(10, askQuestionEntry{Header: "D2 Esfuerzo", Question: "D2: ¿cómo fijamos el esfuerzo?", Options: []askQuestionOption{
					{Label: "D2-A Todos explícitos (Recomendado)", Description: "Ningún rol hereda el esfuerzo."},
				}}, askQuestionEntry{Header: "D3 Portab.", Question: "D3: ¿es portable?", Options: []askQuestionOption{
					{Label: "D3-A effort portable (Recomendado)", Description: "Un campo común."},
				}}),
				askQuestion(20, askQuestionEntry{Header: "Ruta", Question: "¿Cómo seguimos con D2-A (esfuerzo en todos los roles) y D3-A (campo portable)?", Options: []askQuestionOption{
					{Label: "R1 flow-plan (Recomendado)", Description: "Plan versionado."},
				}}),
			},
			want: "pass",
		},
		{
			name: "same-message-does-not-count-as-citation",
			events: []traceEvent{
				assistantText(1, "m1", "- **D5**: mover la validación al pipeline.\nSeguimos con D5 más abajo, en el mismo mensaje."),
			},
			want: "not_observed",
		},
		{
			name: "never-defined-id-is-not-a-citation",
			events: []traceEvent{
				assistantText(1, "m1", "Este mensaje solo menciona D9 una vez."),
			},
			want: "not_observed",
		},
		{
			// The "text right before the question" case: the definition and
			// the bare citation share one real message key ("m1", as they
			// would if a host emitted the text and the AskUserQuestion call
			// as two content blocks of one native turn), which the criterion
			// must still treat as the question's own separate message
			// (design.md's "a question event is always its own message"),
			// not as a same-message non-citation. Reverting that override —
			// using the question's own e.Message instead of a synthetic
			// per-event key — would make this wrongly read "not_observed".
			name: "definition-and-question-sharing-one-real-message-still-cites",
			events: func() []traceEvent {
				text := assistantText(1, "m1", "- **D5**: mover la validación al pipeline de CI antes del despliegue.")
				q := askQuestion(1, askQuestionEntry{Header: "Cierre", Question: "¿Aplico D5 ahora?"})
				q.Message = "m1"
				return []traceEvent{text, q}
			}(),
			want: "fail",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The text-then-question fixture reuses message "m1" for both
			// events on purpose (see its comment); every other case uses a
			// distinct message key per event already.
			r := result{Trace: traceReport{Events: tc.events}}
			if got := citedIDGlossed(r).Status; got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

// TestCitedIDGlossedRangeRequiresMatchingLetters proves the letter-equality
// check findIDOccurrences applies in Go (rangeIDToken cannot express it,
// having no backreference): "D1–S6" does not form one range citation since
// its letters differ. D1 (defined earlier) is read as its own single
// citation, glossed by the immediate en dash that used to look like a range
// separator; S6 was never defined anywhere, so it is not a citation at all.
// Both facts must hold for the result to be "pass" — if the letter check
// were removed, this would combine into a single "D1–S6" range citation
// that is, per design.md, citation-eligible because D1 is a defined
// endpoint, and unglossed (nothing follows "S6"), flipping this to "fail".
func TestCitedIDGlossedRangeRequiresMatchingLetters(t *testing.T) {
	r := result{Trace: traceReport{Events: []traceEvent{
		assistantText(1, "m1", "- **D1**: mover la validación al pipeline de CI."),
		assistantText(2, "m2", "Reviso D1–S6 y sigo con lo demás."),
	}}}
	if got := citedIDGlossed(r).Status; got != "pass" {
		t.Fatalf("got %s want pass", got)
	}
}

func TestCitedIDGlossedEvidenceNeverIncludesFreeText(t *testing.T) {
	r := result{Trace: traceReport{Events: []traceEvent{
		defineS1S5(1, "m1"),
		assistantText(2, "m2", "El lote S1–S5 está completo y desplegado en los seis hosts."),
	}}}
	a := citedIDGlossed(r)
	if a.Status != "fail" {
		t.Fatalf("expected fail, got %+v", a)
	}
	for _, e := range a.Evidence {
		if strings.Contains(e, "completo") || strings.Contains(e, "desplegado") || strings.Contains(e, "hosts") {
			t.Fatalf("evidence leaked free text: %q", e)
		}
		if !strings.HasPrefix(e, "stdout.jsonl:2 text") {
			t.Fatalf("evidence format changed: %q", e)
		}
		if !strings.HasSuffix(e, "S1–S5") {
			t.Fatalf("evidence missing the bare ID token: %q", e)
		}
	}
}

// --- no_bare_url (report-readability D2-A/D3-A) ---

func TestNoBareURL(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
		want string
	}{
		{"plain-bare-url-fails", "Valida en https://example.com/one y confirma.", "fail"},
		{"markdown-link-passes", "Revisa la [documentación](https://example.com/docs) para más detalle.", "pass"},
		{"angle-bracket-autolink-passes", "Abre <https://example.com/ref> directamente.", "pass"},
		{"inline-code-passes", "El comando `curl https://example.com/api` prueba el endpoint.", "pass"},
		{"triple-backtick-fence-passes", "Corre esto:\n```\ncurl https://example.com/raw\n```\n", "pass"},
		{"tilde-fence-passes", "Corre esto:\n~~~\ncurl https://example.com/raw2\n~~~\n", "pass"},
		{"no-url-is-not-observed", "No hay ningún enlace en este mensaje.", "not_observed"},
		{"one-excluded-and-one-bare-still-fails", "Ver [doc](https://example.com/docs) y también https://example.com/bare.", "fail"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Trace: traceReport{Events: []traceEvent{assistantText(1, "m1", tc.text)}}}
			if got := noBareURL(r).Status; got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func TestNoBareURLEvidenceNeverIncludesURL(t *testing.T) {
	r := result{Trace: traceReport{Events: []traceEvent{
		assistantText(7, "m1", "Valida en https://example.com/signed?token=abc123secret y confirma."),
	}}}
	a := noBareURL(r)
	if a.Status != "fail" {
		t.Fatalf("expected fail, got %+v", a)
	}
	for _, e := range a.Evidence {
		if strings.Contains(e, "https://") || strings.Contains(e, "token") {
			t.Fatalf("evidence leaked the URL: %q", e)
		}
		if !strings.HasPrefix(e, "stdout.jsonl:7 text") {
			t.Fatalf("evidence format changed: %q", e)
		}
	}
}

// --- /code-review defects H1-H4 (cited_id_glossed) ---

// TestCitedIDGlossedCodexTextWithNoMessageStillCitesAcrossEvents is H1:
// trace.go's Codex branch (agent_message) sets no Message on its text
// events, so they all carry the same empty key "". Two such events must
// still be treated as separate messages (a per-event fallback key), or a
// citation in the second can never be checked against a definition in the
// first.
func TestCitedIDGlossedCodexTextWithNoMessageStillCitesAcrossEvents(t *testing.T) {
	events := []traceEvent{
		assistantText(1, "", "- **D1**: mover la validación al pipeline de CI antes del despliegue."),
		assistantText(2, "", "Aplico D1."),
	}
	r := result{Trace: traceReport{Events: events}}
	if got := citedIDGlossed(r).Status; got != "fail" {
		t.Fatalf("got %s want fail (two Codex-style text events with no Message must still count as separate messages)", got)
	}
}

// flatAskQuestion builds a "question" event using Pi/Grok's flat
// AskUserQuestion input shape ({"question":…,"options":[…]}, no "questions"
// wrapper array), as seen in question_after_detail/pi-*.jsonl,
// question_after_detail/grok-*.jsonl, and close_question_after_report/
// grok-*.jsonl (after grokToolCall unwraps its use_tool envelope).
func flatAskQuestion(line int, entry askQuestionEntry) traceEvent {
	input, _ := json.Marshal(entry)
	return traceEvent{Line: line, Kind: "question", Tool: "ask_user_question", Input: input}
}

// TestCitedIDGlossedFlatQuestionShapeDefinedViaLabelCitedBareFails is H2's
// "citing bare" case: Pi/Grok's flat shape (no "questions" wrapper) must
// still let an option label define an ID, and a later flat question's own
// text still cites it.
func TestCitedIDGlossedFlatQuestionShapeDefinedViaLabelCitedBareFails(t *testing.T) {
	events := []traceEvent{
		flatAskQuestion(10, askQuestionEntry{Question: "¿Cómo seguimos con el piloto?", Options: []askQuestionOption{
			{Label: "D6-A: piloto en un proyecto (Recomendado)", Description: "El otro proyecto sigue igual."},
			{Label: "D6-B: pausar", Description: "Frenar el otro proyecto."},
		}}),
		flatAskQuestion(20, askQuestionEntry{Question: "¿Qué hacemos con D6-A?"}),
	}
	r := result{Trace: traceReport{Events: events}}
	if got := citedIDGlossed(r).Status; got != "fail" {
		t.Fatalf("got %s want fail (flat-shape option label must still define, flat-shape question text must still cite)", got)
	}
}

// TestCitedIDGlossedFlatQuestionShapeDefinedViaLabelCitedGlossedPasses is
// H2's "defining via label" case checked on its pass path: the same flat
// definition, cited later with an immediate gloss.
func TestCitedIDGlossedFlatQuestionShapeDefinedViaLabelCitedGlossedPasses(t *testing.T) {
	events := []traceEvent{
		flatAskQuestion(10, askQuestionEntry{Question: "¿Cómo seguimos con el piloto?", Options: []askQuestionOption{
			{Label: "D6-A: piloto en un proyecto (Recomendado)", Description: "El otro proyecto sigue igual."},
		}}),
		flatAskQuestion(20, askQuestionEntry{Question: "¿Qué hacemos con D6-A (el piloto en curso)?"}),
	}
	r := result{Trace: traceReport{Events: events}}
	if got := citedIDGlossed(r).Status; got != "pass" {
		t.Fatalf("got %s want pass", got)
	}
}

// TestCitedIDGlossedMentionBeforeItsOwnDefinitionIsNotACitation is H3: a
// token mentioned in an earlier message, then defined only later, must not
// be read as a citation of that later definition — definedness must respect
// trace order.
func TestCitedIDGlossedMentionBeforeItsOwnDefinitionIsNotACitation(t *testing.T) {
	events := []traceEvent{
		assistantText(1, "m1", "Menciono S3 en el texto, sin definirlo todavía."),
		assistantText(2, "m2", "- S3: hallazgo definido recién ahora, más abajo en la sesión."),
	}
	r := result{Trace: traceReport{Events: events}}
	if got := citedIDGlossed(r).Status; got != "not_observed" {
		t.Fatalf("got %s want not_observed (S3's mention precedes its only definition, so it is not yet a citation)", got)
	}
}

// TestCitedIDGlossedRangeDefinitionExpandsAllMembers is H4's range-expansion
// case: a range at a definition position must define every ID it spans, not
// only its two endpoints, so a later bare citation of a middle member (S2,
// neither endpoint of S1–S3) is still recognized as a citation and fails.
func TestCitedIDGlossedRangeDefinitionExpandsAllMembers(t *testing.T) {
	events := []traceEvent{
		assistantText(1, "m1", "- S1–S3: tres hallazgos relacionados del mismo cierre."),
		assistantText(2, "m2", "Reviso S2 antes de seguir."),
	}
	r := result{Trace: traceReport{Events: events}}
	if got := citedIDGlossed(r).Status; got != "fail" {
		t.Fatalf("got %s want fail (S2 is a member of the S1–S3 range definition, not just an endpoint)", got)
	}
}

// TestCitedIDGlossedRangeDefinitionCapsSpanExpansion proves the expansion
// above is capped: a very wide same-letter range (A1–A99, a 99-member span)
// must not expand every member, only register its two endpoints, so a later
// bare mention of a middle member (A50) is still never treated as a
// citation at all.
func TestCitedIDGlossedRangeDefinitionCapsSpanExpansion(t *testing.T) {
	events := []traceEvent{
		assistantText(1, "m1", "- A1–A99: rango deliberadamente amplio para esta prueba."),
		assistantText(2, "m2", "Reviso A50 antes de seguir."),
	}
	r := result{Trace: traceReport{Events: events}}
	if got := citedIDGlossed(r).Status; got != "not_observed" {
		t.Fatalf("got %s want not_observed (A50 must not be registered as defined: the 99-member span exceeds the expansion cap)", got)
	}
}

// TestCitedIDGlossedQuestionTextOpeningWithIDDefinesIt is H4's second case:
// a native question's own question text that opens with an ID (design.md's
// "D4: …" / "D4 — …" convention, observed in real sessions) must define that
// ID, exactly as an assistant-text line or an option label would.
func TestCitedIDGlossedQuestionTextOpeningWithIDDefinesIt(t *testing.T) {
	events := []traceEvent{
		askQuestion(10, askQuestionEntry{Header: "D8", Question: "D8: ¿Aplico el cambio en el pipeline ahora?"}),
		askQuestion(20, askQuestionEntry{Header: "Cierre", Question: "¿Confirmamos D8?"}),
	}
	r := result{Trace: traceReport{Events: events}}
	if got := citedIDGlossed(r).Status; got != "fail" {
		t.Fatalf("got %s want fail (the first question's own text opens with D8, defining it; the second cites it bare)", got)
	}
}
