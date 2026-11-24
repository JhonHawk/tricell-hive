package content

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The repository is public. Client names and ticket prefixes cannot be written
// here, so they live in a git-ignored list in the main checkout; a failure
// names the list's line, never the term, because test output gets pasted into
// pull requests and session reports.
const privateListPath = "_support/workspace/local-context/private-identifiers.local"

var (
	localUserPath = regexp.MustCompile(`/Users/([A-Za-z0-9_-]+)(?:[/"'\s]|$)|-Users-([A-Za-z0-9_]+)(?:[-/"'\s]|$)`)
	exampleUsers  = map[string]bool{"alice": true, "someone": true, "test": true, "me": true, "example": true, "user": true, "you": true, "x": true, "u": true, "o": true}
)

type privateTerm struct {
	line int
	re   *regexp.Regexp
}

func parsePrivateTerms(data string) (terms []privateTerm, problems []string) {
	for i, raw := range strings.Split(data, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		term := ""
		if len(fields) == 2 {
			term = strings.TrimSpace(fields[0])
		}
		if term == "" || strings.TrimSpace(fields[len(fields)-1]) == "" || strings.ContainsAny(term, " \t") {
			problems = append(problems, fmt.Sprintf("%s line %d is malformed: want <term><TAB><neutral example>", privateListPath, i+1))
			continue
		}
		pattern := `(?i)(?:^|[^A-Za-z0-9])` + regexp.QuoteMeta(term) + `(?:[^A-Za-z0-9]|$)`
		if prefix, ok := strings.CutPrefix(term, "prefix:"); ok {
			q := regexp.QuoteMeta(prefix)
			pattern = `(?i:(?:^|[^A-Za-z0-9])` + q + `[-_ ]?[0-9])|(?:^|[^A-Za-z0-9])` + q + `(?:-|\\)`
		}
		terms = append(terms, privateTerm{line: i + 1, re: regexp.MustCompile(pattern)})
	}
	return terms, problems
}

// maskedPath hides each private term and local user name in a file path,
// so a report about the file never prints them.
func maskedPath(file string, terms []privateTerm) string {
	masked := file
	for _, m := range localUserPath.FindAllStringSubmatchIndex(file, -1) {
		for g := 2; g < len(m); g += 2 {
			if m[g] >= 0 && !exampleUsers[file[m[g]:m[g+1]]] {
				masked = masked[:m[g]] + strings.Repeat("*", m[g+1]-m[g]) + masked[m[g+1]:]
			}
		}
	}
	for _, term := range terms {
		for _, m := range term.re.FindAllStringIndex(masked, -1) {
			masked = masked[:m[0]] + strings.Repeat("*", m[1]-m[0]) + masked[m[1]:]
		}
	}
	return masked
}

func privateIdentifierProblems(root string, files []string, terms []privateTerm, skipPaths string) []string {
	var problems []string
	check := func(where, text string, paths bool) {
		for _, m := range localUserPath.FindAllStringSubmatch(text, -1) {
			if paths && !exampleUsers[m[1]+m[2]] {
				problems = append(problems, where+": a local user path; write /Users/me/")
				break
			}
		}
		for _, term := range terms {
			if term.re.MatchString(text) {
				problems = append(problems, fmt.Sprintf("%s: matches line %d of %s; write the neutral example that line gives", where, term.line, privateListPath))
			}
		}
	}
	for _, file := range files {
		name := maskedPath(file, terms)
		if name != file {
			problems = append(problems, name+": its path holds a private term or a local user name, shown as *")
		}
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			continue
		}
		if bytes.IndexByte(data[:min(len(data), 8000)], 0) >= 0 {
			continue
		}
		for n, line := range strings.Split(string(data), "\n") {
			check(fmt.Sprintf("%s:%d", name, n+1), line, file != skipPaths)
		}
	}
	return problems
}

func gitOutput(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
	return string(out)
}

func TestNoPrivateIdentifiers(t *testing.T) {
	root := strings.TrimSpace(gitOutput(t, "rev-parse", "--show-toplevel"))
	var files []string
	for _, file := range strings.Split(gitOutput(t, "-C", root, "ls-files", "-z", "--cached", "--others", "--exclude-standard"), "\x00") {
		if file != "" {
			files = append(files, file)
		}
	}
	mainCheckout := filepath.Dir(strings.TrimSpace(gitOutput(t, "-C", root, "rev-parse", "--path-format=absolute", "--git-common-dir")))
	var terms []privateTerm
	t.Run("client names", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(mainCheckout, privateListPath))
		if err != nil {
			t.Skipf("%s not found in the main checkout: client names and ticket prefixes not checked", privateListPath)
		}
		var problems []string
		terms, problems = parsePrivateTerms(string(data))
		for _, problem := range problems {
			t.Error(problem)
		}
	})
	problems := privateIdentifierProblems(root, files, terms, "tests/content/private_identifiers_test.go")
	for i, problem := range problems {
		if i == 50 {
			t.Errorf("%d more", len(problems)-50)
			break
		}
		t.Error(problem)
	}
}

func TestPrivateIdentifierScanHidesTheTerm(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"a.md":                       "see Zorblax here\nlog at /Users/realperson/work\nticket ZQ-12 and ZQ34, zq-5, zq_6 and zq 7\nplaceholder ZQ-XXX and ZQ\\u0007\n$zorblax_log and tmp/-Users-realperson-x\n",
		"b.md":                       "Zorblaxes and /Users/me/ and /Users/x/ are fine; ZQX-1 too, and -Users-me-\n",
		"zorblax-plan.md":            "plain\n",
		"-Users-realperson/notes.md": "cd /Users/realperson\n",
	}
	for name, text := range files {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, name)), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	terms, malformed := parsePrivateTerms("zorblax\tproject-z\nprefix:ZQ\tPRQ\nno separator here\n")
	if len(malformed) != 1 || !strings.Contains(malformed[0], "line 3") || strings.Contains(malformed[0], "separator") {
		t.Fatalf("malformed line report = %q, want line 3 without its content", malformed)
	}
	got := strings.Join(privateIdentifierProblems(root, []string{"a.md", "b.md", "zorblax-plan.md", "-Users-realperson/notes.md"}, terms, ""), "\n")
	for _, want := range []string{"a.md:1: matches line 1", "a.md:2: a local user path", "a.md:3: matches line 2", "a.md:4: matches line 2", "a.md:5: matches line 1", "a.md:5: a local user path", "plan.md: its path holds", "-Users-**********/notes.md: its path holds", "-Users-**********/notes.md:1: a local user path"} {
		if !strings.Contains(got, want) {
			t.Errorf("problems lack %q:\n%s", want, got)
		}
	}
	if strings.Contains(strings.ToLower(got), "zorblax") || strings.Contains(got, "realperson") || strings.Contains(got, "b.md") {
		t.Errorf("problems print a private term or flag b.md:\n%s", got)
	}
}
