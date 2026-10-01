// project_write.go edits and writes the `## Hive` section of a repository's
// AGENTS.md (design.md "Escritura de ## Hive"). The edit is a pure function on
// bytes; the write re-checks the file and replaces it atomically. Neither
// touches a file other than AGENTS.md, and neither runs anything but fixed
// `git` argument vectors through newGitRunner.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

const (
	projectFileName  = "AGENTS.md"
	changedSinceText = "AGENTS.md changed since the preview"
	claudeMDText     = "CLAUDE.md does not import @AGENTS.md; Claude Code reads CLAUDE.md instead of AGENTS.md by default and will not read this section."
)

// projectWriteHook, when set by a test, runs after the temporary file is ready
// and before it is published: the window a concurrent writer can use.
var projectWriteHook func()

// hiveKeyRank orders keys as hiveSettingKeys does; unknown keys go last.
func hiveKeyRank(name string) int {
	for i, k := range hiveSettingKeys {
		if k.Name == name {
			return i
		}
	}
	return len(hiveSettingKeys)
}

func isRequiredHiveKey(name string) bool {
	for _, k := range hiveSettingKeys {
		if k.Name == name {
			return k.Required
		}
	}
	return false
}

// editHiveSection returns data with the ## Hive section's items set or unset.
// An existing `- Key:` line is replaced where it stands, a key that is not
// there yet goes after the last item in the order of hiveSettingKeys, and
// everything else, unknown keys and prose included, is left as it was. A
// document without the section gets one at the end. A duplicated section, or a
// repeated key among those edited, is an error because the edit would have no
// single place to land.
func editHiveSection(data []byte, set []hiveItem, unset []string) ([]byte, error) {
	setIdx := map[string]int{}
	for i, it := range set {
		if it.Key == "" || strings.ContainsAny(it.Key, ":\r\n") {
			return nil, fmt.Errorf("invalid key %q", shown(it.Key))
		}
		if strings.ContainsAny(it.Value, "\r\n") {
			return nil, fmt.Errorf("%s: value has a line break", it.Key)
		}
		if _, dup := setIdx[it.Key]; dup {
			return nil, fmt.Errorf("%s is set more than once", it.Key)
		}
		setIdx[it.Key] = i
	}
	unsetKeys := map[string]bool{}
	for _, k := range unset {
		unsetKeys[k] = true
		if _, both := setIdx[k]; both {
			return nil, fmt.Errorf("%s is both set and unset", k)
		}
	}
	content := string(data)
	w := walkHiveSection(content)
	if w.Headings > 1 {
		return nil, fmt.Errorf("%s: section appears %d times; keep one", hiveHeading, w.Headings)
	}
	eol := "\n"
	if first, _, ok := strings.Cut(content, "\n"); ok && strings.HasSuffix(first, "\r") {
		eol = "\r\n"
	}
	ordered := make([]hiveItem, len(set))
	copy(ordered, set)
	sort.SliceStable(ordered, func(i, j int) bool { return hiveKeyRank(ordered[i].Key) < hiveKeyRank(ordered[j].Key) })
	render := func(it hiveItem) string { return "- " + it.Key + ": " + it.Value }

	if w.Headings == 0 {
		if len(set) == 0 {
			return data, nil
		}
		var b bytes.Buffer
		b.Write(data)
		if len(data) > 0 {
			if !bytes.HasSuffix(data, []byte("\n")) {
				b.WriteString(eol)
			}
			b.WriteString(eol)
		}
		b.WriteString(hiveHeading + eol + eol)
		for _, it := range ordered {
			b.WriteString(render(it) + eol)
		}
		return b.Bytes(), nil
	}

	counts := map[string]int{}
	for _, it := range w.Items {
		counts[it.Key]++
	}
	for k, n := range counts {
		if _, edited := setIdx[k]; (edited || unsetKeys[k]) && n > 1 {
			return nil, fmt.Errorf("%s appears more than once in %s", k, hiveHeading)
		}
	}
	// The final newline, and the file's line ending, are kept whatever is added:
	// the lines are edited without the last terminator, which goes back at the end.
	finalNL := strings.HasSuffix(content, "\n")
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	cr := strings.TrimSuffix(eol, "\n") // "\r" for a CRLF file, "" for LF
	itemAt := map[int]hiveItem{}
	for i, line := range w.Lines {
		itemAt[line] = w.Items[i]
	}
	insertAt := w.Heading + 1
	if n := len(w.Lines); n > 0 {
		insertAt = w.Lines[n-1] + 1
	} else if insertAt < len(lines) && strings.TrimSpace(lines[insertAt]) == "" {
		insertAt++
	}
	present := map[string]bool{}
	var added []string
	for _, it := range ordered {
		if _, there := counts[it.Key]; !there {
			added = append(added, render(it)+cr)
		}
		present[it.Key] = true
	}
	if len(w.Lines) == 0 && len(added) > 0 {
		switch {
		case w.Heading+1 >= len(lines):
			// The heading is the last line: a blank line goes between it and the items.
			added = append([]string{cr}, added...)
		case insertAt < len(lines) && strings.TrimSpace(lines[insertAt]) != "":
			// The next heading follows directly: a blank line goes after what is added.
			added = append(added, cr)
		}
	}
	if insertAt >= len(lines) && len(added) > 0 && cr != "" && !strings.HasSuffix(lines[len(lines)-1], "\r") {
		lines[len(lines)-1] += cr // a last line without a terminator gets the file's line ending before more follows
	}
	out := make([]string, 0, len(lines)+len(added))
	for i := 0; i <= len(lines); i++ {
		if i == insertAt {
			out = append(out, added...)
		}
		if i == len(lines) {
			break
		}
		it, isItem := itemAt[i]
		switch {
		case isItem && unsetKeys[it.Key]:
			continue
		case isItem:
			if idx, ok := setIdx[it.Key]; ok && counts[it.Key] == 1 {
				out = append(out, render(set[idx])+trailingCR(lines[i]))
				continue
			} else if ok && counts[it.Key] > 1 {
				return nil, fmt.Errorf("%s appears more than once in %s", it.Key, hiveHeading)
			}
		}
		out = append(out, lines[i])
	}
	result := strings.Join(out, "\n")
	if finalNL {
		result += "\n"
	} else {
		result = strings.TrimSuffix(result, "\r") // no terminator at the end, so no bare CR either
	}
	return []byte(result), nil
}

func trailingCR(line string) string {
	if strings.HasSuffix(line, "\r") {
		return "\r"
	}
	return ""
}

// readProjectFile reads AGENTS.md at path for editing. A missing file is not
// an error (existed false). A symbolic link, anything that is not a regular
// file, and a file over the 1 MiB limit are refused: the root comes from git's
// physical path and AGENTS.md sits directly in it, so the file itself is the
// only link that could redirect a write.
func readProjectFile(path string) (data []byte, existed bool, mode os.FileMode, err error) {
	info, err := os.Lstat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, false, 0, nil
	case err != nil:
		return nil, false, 0, fmt.Errorf("cannot read %s: %v", projectFileName, err)
	case info.Mode()&os.ModeSymlink != 0:
		return nil, false, 0, fmt.Errorf("%s is a symbolic link; edit the file it points to instead", projectFileName)
	case !info.Mode().IsRegular():
		return nil, false, 0, fmt.Errorf("%s is not a regular file", projectFileName)
	case info.Size() > hiveFileLimit:
		return nil, false, 0, fmt.Errorf("%s is larger than 1 MiB; not editing it", projectFileName)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, false, 0, fmt.Errorf("cannot read %s: %v", projectFileName, err)
	}
	defer f.Close()
	data, err = io.ReadAll(io.LimitReader(f, hiveFileLimit+1))
	if err != nil {
		return nil, false, 0, fmt.Errorf("cannot read %s: %v", projectFileName, err)
	}
	if len(data) > hiveFileLimit {
		return nil, false, 0, fmt.Errorf("%s is larger than 1 MiB; not editing it", projectFileName)
	}
	return data, true, info.Mode(), nil
}

// writeProjectFile replaces path with after. before is what the preview read
// and existed says whether the file was there; the file is read again and must
// still match. An existing file is replaced atomically with its permissions
// kept. A new file is published with a hard link, which fails instead of
// replacing a file someone created in the meantime. The temporary file is
// always removed. A concurrent edit between the second read and the rename
// can still be lost, and a hard link to the old file is left pointing at the
// old content: without a lock shared with editors there is no way around both.
func writeProjectFile(path string, before []byte, existed bool, after []byte) (err error) {
	cur, exists, mode, err := readProjectFile(path)
	if err != nil {
		return err
	}
	if exists != existed || (existed && !bytes.Equal(cur, before)) {
		return errors.New(changedSinceText)
	}
	perm := os.FileMode(0o644)
	if existed {
		perm = mode.Perm()
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+projectFileName+".hive-tmp-*")
	if err != nil {
		return fmt.Errorf("cannot write %s: %v", projectFileName, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return fmt.Errorf("cannot write %s: %v", projectFileName, err)
	}
	if _, err := tmp.Write(after); err != nil {
		tmp.Close()
		return fmt.Errorf("cannot write %s: %v", projectFileName, err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("cannot write %s: %v", projectFileName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cannot write %s: %v", projectFileName, err)
	}
	if projectWriteHook != nil {
		projectWriteHook()
	}
	if existed {
		if err := os.Rename(tmpName, path); err != nil {
			return fmt.Errorf("cannot write %s: %v", projectFileName, err)
		}
	} else if err := os.Link(tmpName, path); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return errors.New(changedSinceText)
		}
		return fmt.Errorf("cannot write %s: %v", projectFileName, err)
	}
	if err := syncDir(dir); err != nil {
		return fmt.Errorf("%s was written but its directory could not be synced: %v", projectFileName, err)
	}
	return nil
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

var githubNamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// scpOriginPattern is the `[user@]github.com:owner/repo` form of a remote.
var scpOriginPattern = regexp.MustCompile(`^(?:[A-Za-z0-9._-]+@)?github\.com:([^/]+)/([^/]+?)/?$`)

// parseGitHubOrigin reads owner and repo from a remote URL. Only the host
// github.com is accepted, exactly; the user and password of an https URL are
// dropped without being read further, so an embedded token never reaches the
// screen or the file. Any other form gives ok false.
func parseGitHubOrigin(raw string) (owner, repo string, ok bool) {
	raw = strings.TrimSpace(raw)
	var path string
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil || (u.Scheme != "https" && u.Scheme != "ssh") || u.Hostname() != "github.com" || u.Port() != "" || u.RawQuery != "" || u.Fragment != "" {
			return "", "", false
		}
		path = strings.Trim(u.Path, "/")
	} else if m := scpOriginPattern.FindStringSubmatch(raw); m != nil {
		path = m[1] + "/" + m[2]
	} else {
		return "", "", false
	}
	path = strings.TrimSuffix(path, ".git")
	owner, repo, found := strings.Cut(path, "/")
	if !found || strings.Contains(repo, "/") || !githubNamePattern.MatchString(owner) || !githubNamePattern.MatchString(repo) {
		return "", "", false
	}
	for _, n := range []string{owner, repo} {
		if n == "." || n == ".." {
			return "", "", false
		}
	}
	return owner, repo, true
}

// suggestHiveValues proposes values for the settings of the repository at
// root. They are read-only queries and only suggestions: the user confirms
// them. Optional settings are never suggested.
func suggestHiveValues(root string, git gitRunner) map[string]string {
	out := map[string]string{}
	if name := filepath.Base(root); name != "" && sanitizeLine(name) == name {
		out["Project"] = name
	}
	if raw, err := git(root, "config", "--get", "remote.origin.url"); err == nil {
		if owner, repo, ok := parseGitHubOrigin(strings.TrimRight(raw, "\r\n")); ok {
			out["Project"] = repo
			out["Tracker"] = "GitHub Issues · " + owner + "/" + repo
		}
	}
	if raw, err := git(root, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		if branch, ok := strings.CutPrefix(strings.TrimRight(raw, "\r\n"), "origin/"); ok && branch != "" && checkBaseBranch(root, branch, git) == "" {
			out["Base branch"] = branch
		}
	}
	if info, err := os.Stat(filepath.Join(root, "_support", "openspec")); err == nil && info.IsDir() {
		out["Specs"] = "_support/openspec"
	}
	return out
}

// claudeMDWarning returns the warning for a CLAUDE.md at root that does not
// import AGENTS.md, or "". A CLAUDE.md that is a link to AGENTS.md is
// AGENTS.md, so it gets none. Source of the text: Claude Code memory
// documentation, «AGENTS.md» (checked 2026-10-01 through Context7).
func claudeMDWarning(root string) string {
	cm := filepath.Join(root, "CLAUDE.md")
	info, err := os.Stat(cm)
	if err != nil || !info.Mode().IsRegular() {
		return ""
	}
	if ag, err := os.Stat(filepath.Join(root, projectFileName)); err == nil && os.SameFile(info, ag) {
		return ""
	}
	f, err := os.Open(cm)
	if err != nil {
		return ""
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, hiveFileLimit))
	if err != nil {
		return ""
	}
	fenced := false
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			fenced = !fenced
			continue
		}
		if !fenced && line == "@AGENTS.md" {
			return ""
		}
	}
	return claudeMDText
}

// projectEdit is a prepared change to AGENTS.md.
type projectEdit struct {
	Root, File    string
	Existed       bool
	Before, After []byte
	Warnings      []string // do not block the write
	Dirty         bool     // AGENTS.md has uncommitted changes
}

// validateHiveRequest checks what the user asked for before any text is
// built: only known keys are set, values are safe to write, and a required
// key is not removed.
func validateHiveRequest(set []hiveItem, unset []string) error {
	var problems []string
	for _, it := range set {
		switch {
		case !knownHiveKey(it.Key):
			problems = append(problems, fmt.Sprintf("%s: unknown key; known keys are %s", shown(it.Key), strings.Join(hiveKeyNames(false), ", ")))
		case sanitizeLine(it.Value) != it.Value:
			problems = append(problems, it.Key+": value has a line break or control character")
		case it.Value == "" && !isRequiredHiveKey(it.Key):
			problems = append(problems, it.Key+": value is empty; use --unset to remove the key")
		case it.Key == "Base branch" && strings.IndexFunc(it.Value, unicode.IsSpace) >= 0:
			// Git never accepts whitespace in a branch name, of any kind (a non-breaking
			// space or an em space included). The doctor check of
			// the same key reads only the first word, so the writer says it for
			// the whole value; the rest is judged by that check, with its runner.
			problems = append(problems, fmt.Sprintf("Base branch: %s is not a valid branch name", shown(it.Value)))
		}
	}
	for _, k := range unset {
		if isRequiredHiveKey(k) {
			problems = append(problems, k+": required key cannot be unset")
		}
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

// prepareProjectEdit resolves the repository that holds project ("" means the
// current directory), reads its AGENTS.md and builds the edited text and the
// findings, without writing. A finding that blocks the write is returned as
// the error; the others come back as Warnings.
func prepareProjectEdit(project string, set []hiveItem, unset []string, git gitRunner) (projectEdit, error) {
	var e projectEdit
	if err := validateHiveRequest(set, unset); err != nil {
		return e, err
	}
	dir := project
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return e, fmt.Errorf("cannot read the current directory: %v", err)
		}
		dir = wd
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return e, fmt.Errorf("cannot resolve %s: %v", shown(dir), err)
	}
	switch info, err := os.Stat(dir); {
	case errors.Is(err, fs.ErrNotExist):
		return e, fmt.Errorf("directory not found: %s", shown(dir))
	case err != nil:
		return e, fmt.Errorf("cannot read %s: %v", shown(dir), err)
	case !info.IsDir():
		return e, fmt.Errorf("not a directory: %s", shown(dir))
	}
	out, err := git(dir, "rev-parse", "--show-toplevel")
	root := strings.TrimRight(out, "\r\n")
	switch {
	case err != nil && !gitAnswered(err):
		return e, fmt.Errorf("cannot run git: %v", err)
	case err != nil || root == "":
		return e, errors.New(outsideGitLine)
	}
	e.Root = root
	e.File = filepath.Join(root, projectFileName)
	if e.Before, e.Existed, _, err = readProjectFile(e.File); err != nil {
		return e, err
	}
	if e.After, err = editHiveSection(e.Before, set, unset); err != nil {
		return e, err
	}
	w := walkHiveSection(string(e.After))
	blocking, warnings := checkHiveText(w.Items, root, git)
	// The preview shows every value sanitized, so a value the file already
	// holds that sanitizing would change must not be written along unseen. The
	// doctor does not do this check: it only reads.
	for _, it := range w.Items {
		if sanitizeLine(it.Key) != it.Key || sanitizeLine(it.Value) != it.Value {
			blocking = append(blocking, truncateRunes(sanitizeLine(it.Key), 60)+": existing value has a control or hidden character; clean it first")
		}
	}
	if len(blocking) > 0 {
		return e, fmt.Errorf("%s not changed: %s", projectFileName, strings.Join(blocking, "; "))
	}
	e.Warnings = warnings
	if w := claudeMDWarning(root); w != "" {
		e.Warnings = append(e.Warnings, w)
	}
	if e.Existed {
		status, err := git(root, "status", "--porcelain", "--", projectFileName)
		e.Dirty = err == nil && strings.TrimSpace(status) != ""
	}
	return e, nil
}

// hiveSectionLines returns the lines of the ## Hive section in data, heading
// first and trailing blank lines dropped, safe to display; or one line saying
// there is none.
func hiveSectionLines(data []byte) []string {
	w := walkHiveSection(string(data))
	if w.Headings == 0 {
		return []string{"(no " + hiveHeading + " section)"}
	}
	lines := strings.Split(string(data), "\n")[w.Heading:w.End]
	for len(lines) > 1 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = shown(strings.TrimRight(l, "\r"))
	}
	return out
}
