package content

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

const reportEndingsPath = "../../content/skills/hive-core/messages/report-endings.md"

var hostCommand = regexp.MustCompile("(^|[\\s`])(/|\\$|@|/skill:)(flow-|hive-mode)")

// nextStepSyntaxProblems checks that the next-step block's template holds no
// command written in one host's syntax, and that the host syntax line follows
// the template directly: a model that rereads only the template copies its
// literal and misses a syntax line placed sections away.
func nextStepSyntaxProblems(doc string) []string {
	section, _, _ := strings.Cut(doc[strings.Index(doc, "## The next-step block")+1:], "\n## ")
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
