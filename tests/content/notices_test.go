package content

import (
	"os"
	"sort"
	"strings"
	"testing"
)

const stdlibName = "Go standard library"

// parseGoMod returns the module path to version of every require entry
// (direct and indirect, block or single-line form) and the version named by
// the go directive, prefixed with "go".
func parseGoMod(data string) (modules map[string]string, goVersion string) {
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
		case fields[0] == "go" && len(fields) == 2:
			goVersion = "go" + fields[1]
		case fields[0] == "require" && len(fields) == 2 && fields[1] == "(":
			inRequire = true
		case fields[0] == "require" && len(fields) == 3:
			modules[fields[1]] = fields[2]
		}
	}
	return modules, goVersion
}

// parseNotices returns the component to version pairs found in the summary
// table and in the "## <component> <version>" section headings outside code
// fences. Each source is returned apart so the test can compare both.
func parseNotices(data string) (table, sections map[string]string) {
	table, sections = map[string]string{}, map[string]string{}
	fenced := false
	for _, line := range strings.Split(data, "\n") {
		if strings.HasPrefix(line, "```") {
			fenced = !fenced
			continue
		}
		if fenced {
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

// noticesProblems lists every disagreement between go.mod and the notices.
func noticesProblems(goMod, notices string) []string {
	modules, goVersion := parseGoMod(goMod)
	want := map[string]string{stdlibName: goVersion}
	for path, version := range modules {
		want[path] = version
	}
	table, sections := parseNotices(notices)
	var problems []string
	for source, got := range map[string]map[string]string{"summary table": table, "section headings": sections} {
		for name, version := range want {
			switch have, ok := got[name]; {
			case !ok:
				problems = append(problems, source+": missing "+name+" "+version)
			case have != version:
				problems = append(problems, source+": "+name+" is "+have+", go.mod has "+version)
			}
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
	modules, goVersion := parseGoMod(string(goMod))
	if len(modules) == 0 || goVersion == "" {
		t.Fatalf("parsed %d modules and go version %q from go.mod; the parser no longer understands it", len(modules), goVersion)
	}
	for _, p := range noticesProblems(string(goMod), string(notices)) {
		t.Error(p)
	}
}

func TestNoticesProblemsDetectsDrift(t *testing.T) {
	const goMod = "module m\n\ngo 1.27.0\n\nrequire (\n\texample.com/a v1.0.0\n)\n\nrequire example.com/b v2.0.0 // indirect\n"
	const notices = "| Component | Version | License | Relationship |\n| --- | --- | --- | --- |\n" +
		"| Go standard library | go1.27.0 | BSD-3-Clause | linked |\n| example.com/a | v1.0.0 | MIT | direct |\n| example.com/b | v2.0.0 | MIT | indirect |\n\n" +
		"## Go standard library go1.27.0\n\n````text\n## not a heading v9\n````\n\n## example.com/a v1.0.0\n\n## example.com/b v2.0.0\n"
	if got := noticesProblems(goMod, notices); len(got) != 0 {
		t.Fatalf("consistent notices reported problems: %v", got)
	}
	cases := map[string]string{
		"missing module":       strings.ReplaceAll(strings.Replace(notices, "| example.com/b | v2.0.0 | MIT | indirect |\n", "", 1), "## example.com/b v2.0.0\n", ""),
		"version mismatch":     strings.ReplaceAll(notices, "example.com/a v1.0.0", "example.com/a v1.0.1"),
		"extra module":         notices + "\n## example.com/c v3.0.0\n",
		"stdlib version":       strings.ReplaceAll(notices, "go1.27.0", "go1.26.0"),
		"table only missing":   strings.Replace(notices, "| example.com/a | v1.0.0 | MIT | direct |\n", "", 1),
		"heading only missing": strings.Replace(notices, "## example.com/a v1.0.0\n", "", 1),
	}
	for name, bad := range cases {
		if len(noticesProblems(goMod, bad)) == 0 {
			t.Errorf("%s: drift not detected", name)
		}
	}
}
