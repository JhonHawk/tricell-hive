package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// --audit-compare: run the auditor of a git ref and the working tree's auditor
// on the same real sessions and print only the criteria whose status changes.
// A line carries the session's file name, the criterion, both statuses and the
// transcript line numbers of the evidence; never any text of the session.

// auditRunner returns the report one auditor gives for a session.
type auditRunner func(path, host string) (auditReport, error)

// auditHostFromPath names the host that wrote a transcript, from where it lives.
func auditHostFromPath(path string) string {
	p := filepath.ToSlash(path)
	switch {
	case strings.Contains(p, "/.claude/projects/"):
		return "claude"
	case strings.Contains(p, "/.codex/sessions/"):
		return "codex"
	case strings.Contains(p, "/.grok/sessions/") && filepath.Base(path) == "updates.jsonl":
		return "grok"
	}
	return ""
}

var compareEvidenceLine = regexp.MustCompile(`^[^:]*:(\d+)\b`)

func evidenceLines(c criterionAssessment) map[int]bool {
	lines := map[int]bool{}
	for _, e := range c.Evidence {
		if m := compareEvidenceLine.FindStringSubmatch(e); m != nil {
			n, _ := strconv.Atoi(m[1])
			lines[n] = true
		}
	}
	return lines
}

// changedLines lists the evidence lines only one side reports; when both
// report the same lines, all of them.
func changedLines(before, after criterionAssessment) string {
	b, a := evidenceLines(before), evidenceLines(after)
	var only, all []int
	for n := range b {
		all = append(all, n)
		if !a[n] {
			only = append(only, n)
		}
	}
	for n := range a {
		if !b[n] {
			all = append(all, n)
			only = append(only, n)
		}
	}
	pick := only
	if len(pick) == 0 {
		pick = all
	}
	sort.Ints(pick)
	parts := make([]string, 0, len(pick))
	seen := -1
	for _, n := range pick {
		if n != seen {
			parts = append(parts, strconv.Itoa(n))
		}
		seen = n
	}
	return strings.Join(parts, ",")
}

func compareAudit(w io.Writer, paths []string, hostOverride string, before, after auditRunner) error {
	for _, path := range paths {
		host := hostOverride
		if host == "" {
			host = auditHostFromPath(path)
		}
		if host == "" {
			return fmt.Errorf("%s: cannot tell the host from the path; pass --host", path)
		}
		old, err := before(path, host)
		if err != nil {
			return fmt.Errorf("%s: auditor-ref: %w", path, err)
		}
		cur, err := after(path, host)
		if err != nil {
			return fmt.Errorf("%s: working tree: %w", path, err)
		}
		index := func(r auditReport) map[string]criterionAssessment {
			m := map[string]criterionAssessment{}
			for _, c := range r.Criteria {
				m[c.Criterion] = c
			}
			return m
		}
		oi, ni := index(old), index(cur)
		seen := map[string]bool{}
		var names []string
		for _, r := range []auditReport{old, cur} {
			for _, c := range r.Criteria {
				if !seen[c.Criterion] {
					seen[c.Criterion] = true
					names = append(names, c.Criterion)
				}
			}
		}
		sort.Strings(names)
		for _, name := range names {
			b, a := "absent", "absent"
			if c, ok := oi[name]; ok {
				b = c.Status
			}
			if c, ok := ni[name]; ok {
				a = c.Status
			}
			if b == a {
				continue
			}
			fmt.Fprintf(w, "%s · %s · %s → %s · %s\n", filepath.Base(path), name, b, a, changedLines(oi[name], ni[name]))
		}
	}
	return nil
}

// buildRefAuditor archives the ref into a temporary directory, builds its
// pilot binary there and returns a runner that calls it. cleanup removes the
// directory.
func buildRefAuditor(source, ref string, opts auditOptions) (run auditRunner, cleanup func(), err error) {
	dir, err := os.MkdirTemp("", "audit-ref-")
	if err != nil {
		return nil, nil, err
	}
	cleanup = func() { os.RemoveAll(dir) }
	fail := func(e error) (auditRunner, func(), error) { cleanup(); return nil, nil, e }
	var stderr bytes.Buffer
	archive := exec.Command("git", "-C", source, "archive", "--format=tar", ref)
	archive.Stderr = &stderr
	tar := exec.Command("tar", "-x", "-C", dir)
	pipe, err := archive.StdoutPipe()
	if err != nil {
		return fail(err)
	}
	tar.Stdin = pipe
	tar.Stderr = &stderr
	if err := tar.Start(); err != nil {
		return fail(err)
	}
	if err := archive.Run(); err != nil {
		tar.Wait()
		return fail(fmt.Errorf("git archive %s: %v: %s", ref, err, strings.TrimSpace(stderr.String())))
	}
	if err := tar.Wait(); err != nil {
		return fail(fmt.Errorf("extracting %s: %v", ref, err))
	}
	bin := filepath.Join(dir, "pilot-ref")
	build := exec.Command("go", "-C", dir, "build", "-o", bin, "./tests/pilot")
	build.Stderr = &stderr
	if out, err := build.Output(); err != nil {
		return fail(fmt.Errorf("building %s: %v: %s%s", ref, err, out, stderr.String()))
	}
	return func(path, host string) (auditReport, error) {
		args := []string{"--audit-session", path, "--host", host}
		if opts.Base != "" {
			args = append(args, "--base", opts.Base)
		}
		if opts.Cwd != "" {
			args = append(args, "--cwd", opts.Cwd)
		}
		var errb bytes.Buffer
		cmd := exec.Command(bin, args...)
		cmd.Stderr = &errb
		out, err := cmd.Output()
		if err != nil {
			return auditReport{}, fmt.Errorf("%v: %s", err, strings.TrimSpace(errb.String()))
		}
		var rep auditReport
		return rep, json.Unmarshal(out, &rep)
	}, cleanup, nil
}

// auditCompareMain implements `--audit-compare --auditor-ref <ref> <sessions…>`.
func auditCompareMain(source, ref, host string, paths []string, opts auditOptions, stdout io.Writer) error {
	if ref == "" || len(paths) == 0 {
		return fmt.Errorf("--audit-compare requires --auditor-ref <ref> and at least one session path")
	}
	before, cleanup, err := buildRefAuditor(source, ref, opts)
	if err != nil {
		return err
	}
	defer cleanup()
	after := func(path, h string) (auditReport, error) { return auditSessionFile(path, h, opts) }
	return compareAudit(stdout, paths, host, before, after)
}
