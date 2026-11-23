package content

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

const reportEndingsPath = "../../content/skills/hive-core/messages/report-endings.md"

var routeFilesWithoutHostCommands = []string{
	"../../content/skills/flow-plan/references/plan-handoff.md",
	"../../content/skills/hive-mode/playbooks/ticket-status.md",
}

var hostCommand = regexp.MustCompile("(^|[\\s`(>])(/skill:|[/$@])(flow-|hive-mode|create-design|create-prd)")

// A model that rereads only the next-step template copies its literal command
// and misses a syntax line placed sections away (Codex sessions, 2026-11-22/23).
func nextStepSyntaxProblems(doc string) []string {
	start := strings.Index(doc, "## The next-step block")
	if start < 0 {
		return []string{"report-endings has no next-step block section"}
	}
	section, _, _ := strings.Cut(doc[start+1:], "\n## ")
	parts := strings.SplitN(section, "```", 3)
	if len(parts) < 3 {
		return []string{"the next-step block has no fenced template"}
	}
	var problems []string
	if m := hostCommand.FindString(parts[1]); m != "" {
		problems = append(problems, "the next-step template writes a command in one host's syntax: "+strings.TrimSpace(m))
	}
	if !strings.HasPrefix(strings.TrimSpace(parts[2]), "Command syntax by host") {
		problems = append(problems, "the host syntax line does not follow the next-step template directly")
	}
	if n := strings.Count(doc, "Command syntax by host"); n != 1 {
		problems = append(problems, "the host syntax line must appear once in report-endings")
	}
	return problems
}

func TestNextStepTemplateKeepsHostSyntaxBesideIt(t *testing.T) {
	b, err := os.ReadFile(reportEndingsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, problem := range nextStepSyntaxProblems(string(b)) {
		t.Error(problem)
	}
}

func TestNextStepRoutesWriteNoHostCommand(t *testing.T) {
	for _, path := range routeFilesWithoutHostCommands {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if m := hostCommand.FindString(string(b)); m != "" {
			t.Errorf("%s writes a command in one host's syntax: %s", path, strings.TrimSpace(m))
		}
	}
}

func TestNextStepSyntaxCatchesAMissingSection(t *testing.T) {
	if len(nextStepSyntaxProblems("## Other\n\n```\n(such as /flow-plan X)\n```\n")) == 0 {
		t.Error("a document without the next-step block section passed")
	}
}
