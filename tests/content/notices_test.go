package content

import (
	"go/version"
	"os"
	"runtime"
	"sort"
	"strings"
	"testing"
)

const stdlibName = "Go standard library"

// parseGoMod returns the module path to version of every require entry
// (direct and indirect, block or single-line form), the version named by the
// go directive, prefixed with "go", and every line that opens a replace
// directive (single-line or block form).
func parseGoMod(data string) (modules map[string]string, goVersion string, replaces []string) {
	modules = map[string]string{}
	inRequire := false
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(strings.SplitN(line, "//", 2)[0])
		fields := strings.Fields(line)
		switch {
		case len(fields) == 0:
		case inRequire:
			if fields[0] == ")" {
				inRequire = false
			} else if len(fields) >= 2 {
				modules[fields[0]] = fields[1]
			}
		case fields[0] == "replace":
			replaces = append(replaces, line)
		case fields[0] == "go" && len(fields) == 2:
			goVersion = "go" + fields[1]
		case fields[0] == "require" && len(fields) == 2 && fields[1] == "(":
			inRequire = true
		case fields[0] == "require" && len(fields) == 3:
			modules[fields[1]] = fields[2]
		}
	}
	return modules, goVersion, replaces
}

// fenceOf reports whether line is a code fence: up to three spaces of
// indentation, then three or more backticks or tildes. closing is true when
// nothing but whitespace follows the run, the only form that can close a block.
func fenceOf(line string) (char byte, n int, closing bool) {
	rest := strings.TrimLeft(line, " ")
	if len(line)-len(rest) > 3 || rest == "" || (rest[0] != '`' && rest[0] != '~') {
		return 0, 0, false
	}
	char = rest[0]
	for n < len(rest) && rest[n] == char {
		n++
	}
	if n < 3 {
		return 0, 0, false
	}
	return char, n, strings.TrimSpace(rest[n:]) == ""
}

// parseNotices returns the component to version pairs found in the summary
// table and in the "## <component> <version>" section headings outside code
// fences. Each source is returned apart so the test can compare both.
func parseNotices(data string) (table, sections map[string]string) {
	table, sections = map[string]string{}, map[string]string{}
	var fenceChar byte // zero outside a code block
	fenceLen := 0
	for _, line := range strings.Split(data, "\n") {
		if char, n, closing := fenceOf(line); char != 0 {
			switch {
			case fenceChar == 0:
				fenceChar, fenceLen = char, n
			case closing && char == fenceChar && n >= fenceLen:
				fenceChar = 0
			}
			continue
		}
		if fenceChar != 0 {
			continue
		}
		if strings.HasPrefix(line, "|") {
			cells := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
			if len(cells) >= 2 {
				name, version := strings.TrimSpace(cells[0]), strings.TrimSpace(cells[1])
				if name != "Component" && !strings.HasPrefix(name, "---") {
					table[name] = version
				}
			}
		} else if rest, ok := strings.CutPrefix(line, "## "); ok {
			if i := strings.LastIndex(rest, " "); i > 0 {
				sections[strings.TrimSpace(rest[:i])] = strings.TrimSpace(rest[i+1:])
			}
		}
	}
	return table, sections
}

// noticesProblems lists every disagreement between go.mod, the Go toolchain
// that builds the manager and the notices. The packager builds with
// GOTOOLCHAIN=local, so the standard library linked into the binary is the
// local toolchain's, passed here as toolchain (runtime.Version() in the test).
// The go directive only sets a floor: the notices must not name an older one.
func noticesProblems(goMod, notices, toolchain string) []string {
	modules, goVersion, replaces := parseGoMod(goMod)
	want := map[string]string{stdlibName: toolchain}
	for path, v := range modules {
		want[path] = v
	}
	table, sections := parseNotices(notices)
	var problems []string
	for _, line := range replaces {
		problems = append(problems, "go.mod has a replace directive ("+line+"): extend the notices check to cover replacements")
	}
	for source, got := range map[string]map[string]string{"summary table": table, "section headings": sections} {
		for name, v := range want {
			switch have, ok := got[name]; {
			case !ok:
				problems = append(problems, source+": missing "+name+" "+v)
			case name == stdlibName && have != v:
				problems = append(problems, source+": "+name+" is "+have+", the local toolchain is "+v)
			case have != v:
				problems = append(problems, source+": "+name+" is "+have+", go.mod has "+v)
			}
		}
		if have, ok := got[stdlibName]; ok && version.Compare(have, goVersion) < 0 {
			problems = append(problems, source+": "+stdlibName+" "+have+" is older than the go directive "+goVersion)
		}
		for name := range got {
			if _, ok := want[name]; !ok {
				problems = append(problems, source+": lists "+name+", which go.mod does not require")
			}
		}
	}
	sort.Strings(problems)
	return problems
}

func TestThirdPartyNoticesMatchGoMod(t *testing.T) {
	goMod, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	notices, err := os.ReadFile("../../THIRD_PARTY_NOTICES.md")
	if err != nil {
		t.Fatal(err)
	}
	modules, goVersion, _ := parseGoMod(string(goMod))
	if len(modules) == 0 || goVersion == "" {
		t.Fatalf("parsed %d modules and go version %q from go.mod; the parser no longer understands it", len(modules), goVersion)
	}
	toolchain := runtime.Version()
	if !version.IsValid(toolchain) {
		t.Fatalf("toolchain version %q is not a release version; run the notices check with a release Go toolchain", toolchain)
	}
	for _, p := range noticesProblems(string(goMod), string(notices), toolchain) {
		t.Error(p)
	}
}

func TestNoticesProblemsDetectsDrift(t *testing.T) {
	const goMod = "module m\n\ngo 1.27.0\n\nrequire (\n\texample.com/a v1.0.0\n)\n\nrequire example.com/b v2.0.0 // indirect\n"
	const notices = "| Component | Version | License | Relationship |\n| --- | --- | --- | --- |\n" +
		"| Go standard library | go1.27.0 | BSD-3-Clause | linked |\n| example.com/a | v1.0.0 | MIT | direct |\n| example.com/b | v2.0.0 | MIT | indirect |\n\n" +
		"## Go standard library go1.27.0\n\n````text\n```\n## not a heading v9\n```\n````\n\n~~~\n## also not a heading v8\n~~~\n\n## example.com/a v1.0.0\n\n## example.com/b v2.0.0\n"
	const toolchain = "go1.27.0"
	if got := noticesProblems(goMod, notices, toolchain); len(got) != 0 {
		t.Fatalf("consistent notices reported problems: %v", got)
	}
	cases := map[string]string{
		"missing module":       strings.ReplaceAll(strings.Replace(notices, "| example.com/b | v2.0.0 | MIT | indirect |\n", "", 1), "## example.com/b v2.0.0\n", ""),
		"version mismatch":     strings.ReplaceAll(notices, "example.com/a v1.0.0", "example.com/a v1.0.1"),
		"extra module":         notices + "\n## example.com/c v3.0.0\n",
		"stdlib version":       strings.ReplaceAll(notices, "go1.27.0", "go1.26.0"),
		"stdlib older than go": strings.ReplaceAll(notices, "go1.27.0", "go1.26.9"),
		"table only missing":   strings.Replace(notices, "| example.com/a | v1.0.0 | MIT | direct |\n", "", 1),
		"heading only missing": strings.Replace(notices, "## example.com/a v1.0.0\n", "", 1),
	}
	for name, bad := range cases {
		if len(noticesProblems(goMod, bad, toolchain)) == 0 {
			t.Errorf("%s: drift not detected", name)
		}
	}
	// The notices follow the local toolchain, not the go directive.
	if got := noticesProblems(goMod, notices, "go1.27.2"); len(got) == 0 {
		t.Error("toolchain newer than the notices: drift not detected")
	}
	patched := strings.ReplaceAll(notices, "go1.27.0", "go1.27.2")
	if got := noticesProblems(goMod, patched, "go1.27.2"); len(got) != 0 {
		t.Errorf("notices matching a patched toolchain reported problems: %v", got)
	}
	for name, mod := range map[string]string{
		"single-line replace": goMod + "replace example.com/a => ../a\n",
		"block replace":       goMod + "replace (\n\texample.com/a => ../a\n)\n",
	} {
		got := noticesProblems(mod, notices, toolchain)
		if len(got) == 0 || !strings.Contains(got[0], "replace directive") {
			t.Errorf("%s: not reported: %v", name, got)
		}
	}
}
