package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The boundary block is the closing paragraph every role in content/agents
// carries. Its parts are shared, so a role that drops or rewords one loses a
// scope, git, dependency, service or process rule without any other signal.
// The block a role needs depends on what its access profile lets it do.
const (
	boundaryScope             = "Stay within the parent’s assigned scope, write ownership, and output destination. "
	boundaryWrites            = "Do not commit, push, open or merge pull requests, write to trackers or launch children. Modify or delete only the paths the brief names, list planned deletions first, and stop and report when the work needs a change outside them. "
	boundaryOwnerWrites       = "Only under an explicit E2E assignment, create local commits in the assigned repository, worktree, branch and allowed paths through the scoped-commit helper and its secret scan, preserving unrelated staging. Do not publish, push, open or merge pull requests, write to trackers, launch children, create or switch worktrees or branches, integrate other units, or use unassigned resources. Modify or delete only the paths the brief names, list planned deletions first, and stop and report when the work needs a change outside them. "
	boundaryTicketOwnerWrites = "Only in the swarm mode, write the change folder and checkpoint of your assigned tickets in your worktree or in the one change-folder path the brief assigns in another repository's checkout, and create local commits only in your assigned worktree, branch and allowed paths through the scoped-commit helper and its secret scan, preserving unrelated staging; never commit in that other repository. Do not publish, push, open or merge pull requests, write to trackers or persistent memory, launch children, create or switch worktrees or branches, or use unassigned resources. Modify or delete only the paths the brief names, list planned deletions first, and stop and report when the work needs a change outside them. "
	boundaryGit               = "Never run `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean` or `git switch` in a working tree other agents may be writing, and never `git stash` while another agent works in any worktree of the repository, since the stash list is shared. "
	boundaryDeps              = "Unless the brief says your dependency installation is not shared, run no command that installs, prunes or relinks dependencies and add none to a manifest or lockfile; pause only the work that needs one, name it in your report, and continue. "
	boundarySvc               = "Run a test suite that sets up, resets or binds a local service only on the instance the brief assigns; with none assigned, do not run it and report it as not run. "
	boundaryRun               = "Run install, build and test commands in the foreground with a timeout of at most 10 minutes; report a timeout as not completed. "
	boundaryStop              = "Stop any process you started before returning. "
	boundaryReturn            = "Return evidence, limitations and needed decisions in one final reply; before it, send at most one message to the parent, for a blocker that needs its decision, and make that the final reply when no independent work remains."
)

// A role's closing paragraph may start with its own final instruction, so the
// block is checked as the paragraph's ending. The scope sentence uses a
// typographic apostrophe and the return sentence an ASCII one, as every role does.
var boundaryByProfile = map[string]string{
	// implement: may write, so it also carries the write-scope rules.
	"implement": boundaryScope + boundaryWrites + boundaryGit + boundaryDeps + boundarySvc + boundaryRun + boundaryStop + boundaryReturn,
	// observe: does not install, run suites or write.
	"observe": boundaryScope + boundaryGit + boundaryStop + boundaryReturn,
	// verify: runs commands, does not write the work under check.
	"verify": boundaryScope + boundaryGit + boundaryDeps + boundarySvc + boundaryRun + boundaryStop + boundaryReturn,
}

// boundaryOverrides lists the roles whose block is not the one of their
// access profile, each with the profile whose block it carries and why.
var boundaryOverrides = map[string]struct {
	block  string
	reason string
}{
	"hive-build-owner": {
		block:  boundaryScope + boundaryOwnerWrites + boundaryGit + boundaryDeps + boundarySvc + boundaryRun + boundaryStop + boundaryReturn,
		reason: "only this explicitly assigned E2E role may commit locally; ordinary implementers retain the no-commit boundary",
	},
	"hive-ticket-owner": {
		block:  boundaryScope + boundaryTicketOwnerWrites + boundaryGit + boundaryDeps + boundarySvc + boundaryRun + boundaryStop + boundaryReturn,
		reason: "only the swarm mode's ticket owner writes its plan and commits locally in its assigned worktree; ordinary implementers retain the no-commit boundary",
	},
	"hive-review-ux": {
		block:  boundaryByProfile["verify"],
		reason: "observes the source but drives a browser and a dev server, so it runs commands like a verifier",
	},
	"hive-verify-task": {
		// Its step 3 sets foreground runs, so the run sentence is absent.
		block:  boundaryScope + boundaryGit + boundaryDeps + boundarySvc + boundaryStop + boundaryReturn,
		reason: "its step 3 already sets foreground runs, so it carries no long-command sentence",
	},
}

func TestAgentBoundaryBlock(t *testing.T) {
	paths, err := filepath.Glob("../../content/agents/*/*.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no roles found under content/agents")
	}
	seenOverride := map[string]bool{}
	for _, path := range paths {
		rel := filepath.ToSlash(path[len("../../"):])
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		name, ok, err := frontmatterScalar(text, "name")
		if err != nil || !ok {
			t.Errorf("%s: frontmatter has no readable `name`", rel)
			continue
		}
		profile, ok, err := frontmatterScalar(text, "access_profile")
		if err != nil || !ok {
			t.Errorf("%s: frontmatter has no readable `access_profile`", rel)
			continue
		}
		want, known := boundaryByProfile[profile]
		if !known {
			t.Errorf("%s: access profile %q has no expected boundary block; add it to boundaryByProfile", rel, profile)
			continue
		}
		if override, ok := boundaryOverrides[name]; ok {
			if (name == "hive-build-owner" || name == "hive-ticket-owner") && profile != "implement" {
				t.Errorf("%s: local owner must use existing implement access profile, got %q", rel, profile)
			}
			want = override.block
			seenOverride[name] = true
		}
		if !strings.HasSuffix(strings.TrimRight(text, "\n"), want) {
			t.Errorf("%s (access profile %q): the role must end with this boundary block:\n%s", rel, profile, want)
		}
	}
	for name := range boundaryOverrides {
		if !seenOverride[name] {
			t.Errorf("boundaryOverrides names %q, but no role has that name", name)
		}
	}
}
