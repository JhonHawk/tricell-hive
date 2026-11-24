package content

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The swarm mode's exceptions to the ordinary flows live only in its skill and
// its role, with one line of the global guidance naming and linking the mode and
// the execution-record script accepting the role. Any other distributed file
// that names them carries an exception into the ordinary flows.
var swarmMention = regexp.MustCompile(`(?i)\bswarm\b|hive-ticket-owner`)

var swarmHomes = []string{
	"content/skills/swarm/",
	"content/agents/development/hive-ticket-owner.md",
	"content/skills/flow-plan/scripts/tasks.py",
}

const swarmGlobalLine = "content/guidance/global.md"

func swarmScopeFindings(root string) ([]string, error) {
	var findings []string
	globalLines := 0
	err := filepath.WalkDir(filepath.Join(root, "content"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		for _, home := range swarmHomes {
			if strings.HasPrefix(rel, home) {
				return nil
			}
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for n, line := range strings.Split(string(data), "\n") {
			if !swarmMention.MatchString(line) {
				continue
			}
			if rel == swarmGlobalLine {
				globalLines++
				continue
			}
			findings = append(findings, rel+":"+strconv.Itoa(n+1))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if globalLines != 1 {
		findings = append(findings, swarmGlobalLine+": names the swarm mode on "+strconv.Itoa(globalLines)+" lines, want exactly 1")
	}
	return findings, nil
}

func TestSwarmModeStaysInItsHomes(t *testing.T) {
	findings, err := swarmScopeFindings("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range findings {
		t.Errorf("%s: the swarm mode or hive-ticket-owner is named outside the swarm skill, its role, one line of the global guidance and tasks.py; move the exception into content/skills/swarm/", f)
	}
}
