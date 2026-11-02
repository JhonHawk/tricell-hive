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
	boundaryScope  = "Stay within the parent’s assigned scope, write ownership, and output destination. "
	boundaryWrites = "Do not commit, push, open or merge pull requests, write to trackers, or launch children of your own. Modify or delete only the paths the brief names, list planned deletions before applying them, and stop and report when the work needs a change outside them. "
	boundaryGit    = "A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, modifies every path it touches: never run one in a working tree that other agents may be writing; compare with the base through `git diff` or `git show <base>:<path>`, or in a separate worktree. `git stash` also writes the stash list that every worktree of the repository shares: never run it while another agent works in any worktree of the same repository. "
	boundaryDeps   = "Unless the brief says your dependency installation is not shared, run no command that installs, prunes, or relinks dependencies, such as `npm ci`, `pnpm install`, or adding a package, and do not add one to a manifest or lockfile yourself: pause only the work that needs the dependency, name it in your report, and continue the rest. "
	boundarySvc    = "Run a test suite that sets up, resets, or binds a local service, such as a test database, a cache, or a port, only on the instance the brief assigns; when it assigns none, do not run the suite and report it as not run. "
	boundaryRun    = "Run every install, build and test command in the foreground with an explicit timeout of at most 10 minutes rather than behind a background monitor, and report one that hits the timeout as not completed. "
	boundaryStop   = "Stop any process you started before returning. "
	boundaryReturn = "Return evidence, limitations, and any decision needed from the parent in one final reply; before it, send the parent at most one message, for a blocker that needs the parent's decision, and return it as the final reply when no independent work remains."
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
