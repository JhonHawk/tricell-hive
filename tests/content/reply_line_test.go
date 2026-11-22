package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// replyFlows end with a row of the report leaves; each names that row in one
// line starting with "Reply:" and adds only what the row does not hold.
var replyFlows = []string{"hive-mode", "flow-plan", "flow-build", "flow-verify", "flow-close", "flow-deploy", "git-workflow"}

// replyLineProblems returns one line per flow whose SKILL.md text does not hold
// exactly one line starting with "Reply:" that links a report leaf.
func replyLineProblems(skills map[string]string) []string {
	var problems []string
	for _, flow := range replyFlows {
		var replies []string
		for _, line := range strings.Split(skills[flow], "\n") {
			if strings.HasPrefix(line, "Reply:") {
				replies = append(replies, line)
			}
		}
		switch {
		case len(replies) != 1:
			problems = append(problems, fmt.Sprintf("%s: want one line starting with Reply:, found %d", flow, len(replies)))
		case !strings.Contains(replies[0], "hive-core/messages/report-endings.md") && !strings.Contains(replies[0], "hive-core/close/close-report.md"):
			problems = append(problems, flow+": its Reply: line links no row of report-endings or close-report")
		}
	}
	return problems
}

func readReplySkills(t *testing.T) map[string]string {
	t.Helper()
	skills := map[string]string{}
	for _, flow := range replyFlows {
		data, err := os.ReadFile(filepath.Join(repoRoot(t), "content", "skills", flow, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		skills[flow] = string(data)
	}
	return skills
}

func TestEachFlowNamesItsReply(t *testing.T) {
	for _, problem := range replyLineProblems(readReplySkills(t)) {
		t.Error(problem)
	}
}

func TestReplyLineCheckNamesTheFlowWithoutOne(t *testing.T) {
	skills := readReplySkills(t)
	var kept []string
	for _, line := range strings.Split(skills["flow-build"], "\n") {
		if !strings.HasPrefix(line, "Reply:") {
			kept = append(kept, line)
		}
	}
	skills["flow-build"] = strings.Join(kept, "\n")
	problems := replyLineProblems(skills)
	if len(problems) != 1 || !strings.HasPrefix(problems[0], "flow-build:") {
		t.Fatalf("removing flow-build's Reply: line gave %v, want one problem naming flow-build", problems)
	}
}
