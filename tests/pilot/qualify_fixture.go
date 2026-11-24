package main

// Runner support for --suite qualification (model-qualification-tool T1;
// design.md «Brazos del runner»): task and fixture preparation, the guidance
// hash recorded in run.json, and the separation of the fixture root.
//
// Identifiers the scorer (T2) and the driver (T3) rely on:
//
//   - qualificationSuite, qualifyArmDeployed, qualifyArmNone: suite and arm names.
//   - loadQualifyFixture(source, taskID) resolves a task to the fixture the
//     runner builds; setupQualifyGit prepares its Git state from the case ID.
//   - guidanceHash(host, home) is the single guidance hash for deployed
//     (home == "") and A (home == the isolated home). None records "none".
//   - run.json carries Task, Split, Arm, Effort and GuidanceHash.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

const (
	qualificationSuite = "qualification"
	qualifyArmDeployed = "deployed"
	qualifyArmNone     = "none"
	// qualifyHashNone is the guidance hash of the none arm: no guidance exists.
	qualifyHashNone = "none"
)

// qualifyIsolatedHome reports whether a qualification arm runs in an isolated
// home: A (with guidance) and none (without). deployed uses the real home.
func qualifyIsolatedHome(arm string) bool { return arm == "A" || arm == qualifyArmNone }

// loadQualifyFixture resolves taskID from <source>'s validated suite into the
// fixture the runner builds. A task with `case` keeps that case's ID, so the
// existing detectors and the Git setup keep deriving from it; an inline task
// uses its own ID. It also returns suite.json's bytes, which the runner hashes
// as the fixture corpus.
func loadQualifyFixture(source, taskID string) (fixture, qualifyTask, []byte, error) {
	var none fixture
	suite, err := loadQualificationSuite(source)
	if err != nil {
		return none, qualifyTask{}, nil, err
	}
	cases, err := loadFlowCaseIndex(source)
	if err != nil {
		return none, qualifyTask{}, nil, err
	}
	raw, err := os.ReadFile(filepath.Join(source, "tests/fixtures/qualification/suite.json"))
	if err != nil {
		return none, qualifyTask{}, nil, err
	}
	for _, t := range suite.Tasks {
		if t.ID != taskID {
			continue
		}
		files, cwd, prompt := resolveQualifyTask(t, cases)
		copied := make(map[string]string, len(files))
		for p, c := range files {
			copied[p] = c
		}
		f := fixture{ID: t.ID, Files: copied, Cwd: cwd, Prompt: prompt}
		if t.Case != "" {
			f.ID = t.Case
		}
		return f, t, raw, nil
	}
	return none, qualifyTask{}, nil, fmt.Errorf("unknown qualification task %q", taskID)
}

// setupQualifyGit prepares the Git state a task needs on top of `git init`:
// the delivery fixture for git-delivery and close-sequence, a committed main
// for the question cases and for inline tasks with initial_commit.
func setupQualifyGit(root string, t qualifyTask, f fixture) (*gitDeliveryReport, error) {
	if t.Case != "" {
		switch {
		case f.ID == "git-delivery" || f.ID == "close-sequence":
			return setupGitDelivery(root)
		case questionFlowCase(f.ID):
			return nil, setupInitialCommit(root)
		}
		return nil, nil
	}
	if t.InitialCommit {
		return nil, setupInitialCommit(root)
	}
	return nil, nil
}

// checkQualifyFixtureRoot refuses a fixture root under --source or --out: a
// host would discover the checkout's AGENTS.md or CLAUDE.md upward from it, and
// the evidence directory would contain the fixture.
func checkQualifyFixtureRoot(root, source, output string) error {
	root, err := target.Canonical(root)
	if err != nil {
		return err
	}
	for _, outer := range []string{source, output} {
		if outer == "" {
			continue
		}
		outer, err := target.Canonical(outer)
		if err != nil {
			return err
		}
		if under(root, outer) {
			return fmt.Errorf("the qualification fixture root %s is under %s; set TMPDIR to a directory outside --source and --out", root, outer)
		}
	}
	return nil
}

// guidanceHash is the one guidance hash for the deployed and A arms. home ""
// reads the real installation (honoring CODEX_HOME and CLAUDE_CONFIG_DIR, as
// management.Status does without an explicit home); a path reads the isolated
// home. It reads the manager's state without writing anything.
//
// The rows are the Status entries of host at user scope that are Hive's own
// installed resources. A symlink contributes its target; a block contributes
// only the text between management.Begin and management.End; any other
// resource contributes its bytes (a directory, its files in order). Each
// content has the absolute home replaced by $HOME before hashing, so two
// installations of the same checkout in different homes hash alike. Voice
// spans, hook entries and update rows are not guidance text and are skipped.
func guidanceHash(host, home string) (string, error) {
	opts := management.Options{Scope: "user", Home: home, Hosts: []string{host}}
	entries, err := management.Status(opts)
	if err != nil {
		return "", err
	}
	realHome := home
	if realHome == "" {
		if realHome, err = os.UserHomeDir(); err != nil {
			return "", err
		}
	}
	canonicalHome, err := target.Canonical(realHome)
	if err != nil {
		return "", err
	}
	homes := []string{canonicalHome}
	if realHome != canonicalHome {
		homes = append(homes, realHome)
	}
	type pair struct{ path, content string }
	var pairs []pair
	for _, en := range entries {
		if en.Host != host || !guidanceHashStatus(en.Status) {
			continue
		}
		switch en.Kind {
		case "voice", "hook_entry", "update":
			continue
		}
		content, err := guidanceResource(en)
		if err != nil {
			return "", err
		}
		rel := en.Path
		if r, err := filepath.Rel(canonicalHome, en.Path); err == nil && under(en.Path, canonicalHome) {
			rel = r
		}
		for _, h := range homes {
			content = strings.ReplaceAll(content, h, "$HOME")
			rel = strings.ReplaceAll(rel, h, "$HOME")
		}
		pairs = append(pairs, pair{rel, content})
	}
	if len(pairs) == 0 {
		return "", fmt.Errorf("no installed Hive guidance found for host %s", host)
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].path != pairs[j].path {
			return pairs[i].path < pairs[j].path
		}
		return pairs[i].content < pairs[j].content
	})
	h := sha256.New()
	for _, p := range pairs {
		fmt.Fprintf(h, "%d:%s%d:%s", len(p.path), p.path, len(p.content), p.content)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// guidanceHashStatus keeps the rows that hold Hive's installed text: not the
// ones the manager reports as absent, foreign, pending migration or stale.
func guidanceHashStatus(status string) bool {
	switch status {
	case "installed", "retained_shared", "drift", "recovery_required":
		return true
	}
	return false
}

// guidanceResource reads one managed resource as text for the hash.
func guidanceResource(en management.StatusEntry) (string, error) {
	info, err := os.Lstat(en.Path)
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("managed guidance file %s is missing: repair the installation (for example with `hive update`) before qualifying", en.Path)
	}
	if err != nil {
		return "", err
	}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		link, err := os.Readlink(en.Path)
		return "symlink:" + link, err
	case info.IsDir():
		var files []string
		err := filepath.WalkDir(en.Path, func(p string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				files = append(files, p)
			}
			return err
		})
		if err != nil {
			return "", err
		}
		sort.Strings(files)
		var b strings.Builder
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				return "", err
			}
			rel, _ := filepath.Rel(en.Path, f)
			fmt.Fprintf(&b, "%d:%s%d:%s", len(rel), rel, len(data), data)
		}
		return b.String(), nil
	}
	data, err := os.ReadFile(en.Path)
	if err != nil {
		return "", err
	}
	if en.Kind != "block" {
		return string(data), nil
	}
	begin := bytes.Index(data, []byte(management.Begin))
	if begin < 0 {
		return "", fmt.Errorf("%s: managed block begin marker missing", en.Path)
	}
	begin += len(management.Begin)
	end := bytes.Index(data[begin:], []byte(management.End))
	if end < 0 {
		return "", fmt.Errorf("%s: managed block end marker missing", en.Path)
	}
	return string(data[begin : begin+end]), nil
}
