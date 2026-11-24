package management

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/pi"
)

func pcOptions(t *testing.T) Options {
	t.Helper()
	o := setup(t)
	o.Hosts = []string{"pi"}
	return o
}

func pcHome(o Options) string { return filepath.Join(o.Home, ".pi", "agent") }
func pcConfig(o Options) string {
	return filepath.Join(pcHome(o), "extensions", "subagent", "config.json")
}

func pcMode(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Mode().Perm()
}

func pcState(t *testing.T, o Options) State {
	t.Helper()
	s, _, err := readState(o.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// AC1: the missing file plans the write; the plan shows path, key and value.
func TestPlanPiConfigMissingFileWrites(t *testing.T) {
	o := pcOptions(t)
	st := plan(t, "install", o).PiSubagentConfig
	if st == nil || st.Status != PiConfigWrite || st.Path != pcConfig(o) || st.Key != "toolActivation" || st.Value != `"eager"` {
		t.Fatalf("step = %#v", st)
	}
	if st.Expected.Exists {
		t.Fatalf("expected fingerprint of an absent file, got %#v", st.Expected)
	}
}

// AC2 (plan side): an object without the key plans a write against its fingerprint.
func TestPlanPiConfigObjectWithoutKeyWrites(t *testing.T) {
	o := pcOptions(t)
	put(t, pcConfig(o), "{\n  \"a\": 1\n}\n")
	st := plan(t, "install", o).PiSubagentConfig
	cur, _ := read(pcConfig(o))
	if st == nil || st.Status != PiConfigWrite || st.Expected != finger(cur) {
		t.Fatalf("step = %#v", st)
	}
}

// AC1: the step also belongs to a package installation and a committed-source
// update; both build their plan through BuildPlan("install"). The plan ID covers it.
func TestPlanPiConfigPlanIDCoversTheStep(t *testing.T) {
	o := pcOptions(t)
	a := plan(t, "install", o)
	b := a
	b.PiSubagentConfig = nil
	if a.ID == planID(b) {
		t.Fatal("plan ID ignores the step")
	}
}

// AC4: any value Hive did not record is the user's: nothing to write or record.
func TestPlanPiConfigUserValues(t *testing.T) {
	for name, body := range map[string]string{
		"auto":             `{"toolActivation": "auto"}`,
		"dynamic":          `{"toolActivation":"dynamic","x":1}`,
		"null":             `{"toolActivation": null}`,
		"eager not ours":   `{"toolActivation": "eager"}`,
		"object":           `{"toolActivation": {"a": 1}}`,
		"long and control": `{"toolActivation": "` + strings.Repeat("x", 500) + `\u0007"}`,
	} {
		t.Run(name, func(t *testing.T) {
			o := pcOptions(t)
			put(t, pcConfig(o), body)
			p := plan(t, "install", o)
			st := p.PiSubagentConfig
			if st == nil || st.Status != PiConfigUserSet || st.Expected.Exists {
				t.Fatalf("step = %#v", st)
			}
			if len(st.Value) > 80 || strings.ContainsAny(st.Value, "\a\n") {
				t.Fatalf("value shown = %q", st.Value)
			}
			apply(t, p)
			if get(t, pcConfig(o)) != body || pcState(t, o).PiSubagentConfig != nil {
				t.Fatal("a user value was written over or recorded")
			}
			absent(t, pcConfig(o)+".hive-backup")
		})
	}
}

// AC4: a file Hive must not edit becomes a manual step; nothing is written.
func TestPlanPiConfigManualSteps(t *testing.T) {
	arrange := map[string]func(t *testing.T, o Options){
		"invalid JSON":   func(t *testing.T, o Options) { put(t, pcConfig(o), `{"a": `) },
		"duplicate keys": func(t *testing.T, o Options) { put(t, pcConfig(o), `{"a": 1, "a": 2}`) },
		"not an object":  func(t *testing.T, o Options) { put(t, pcConfig(o), `[1]`) },
		"symbolic link": func(t *testing.T, o Options) {
			real := filepath.Join(filepath.Dir(o.Home), "real.json")
			put(t, real, "{}")
			if err := os.MkdirAll(filepath.Dir(pcConfig(o)), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(real, pcConfig(o)); err != nil {
				t.Fatal(err)
			}
		},
		"not a regular file": func(t *testing.T, o Options) {
			if err := os.MkdirAll(pcConfig(o), 0700); err != nil {
				t.Fatal(err)
			}
		},
		"larger than 1 MiB": func(t *testing.T, o Options) {
			put(t, pcConfig(o), `{"a": "`+strings.Repeat("x", pi.MaxSubagentConfigBytes)+`"}`)
		},
		"folder not writable": func(t *testing.T, o Options) {
			if err := os.MkdirAll(filepath.Dir(pcConfig(o)), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(filepath.Dir(pcConfig(o)), 0500); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { os.Chmod(filepath.Dir(pcConfig(o)), 0700) })
		},
	}
	// The legacy-installer scan, which runs first, already fails a whole plan on
	// invalid JSON, duplicate keys, a non-object, a symbolic link or a non-regular
	// file, so those reach the step only through planPiSubagentConfig itself.
	reachable := map[string]bool{"larger than 1 MiB": true, "folder not writable": true}
	for name, f := range arrange {
		t.Run(name, func(t *testing.T) {
			o := pcOptions(t)
			var p Plan
			if reachable[name] {
				f(t, o)
				p = plan(t, "install", o)
			} else {
				p = plan(t, "install", o)
				f(t, o)
				p.PiSubagentConfig = nil
				if err := planPiSubagentConfig(&p, State{}); err != nil {
					t.Fatal(err)
				}
			}
			st := p.PiSubagentConfig
			if st == nil || st.Status != PiConfigManual || st.Detail == "" || st.Value != `"eager"` || st.Key != "toolActivation" {
				t.Fatalf("step = %#v", st)
			}
			if strings.Contains(st.Detail, o.Home) {
				t.Fatalf("the cause repeats the path: %q", st.Detail)
			}
			if reachable[name] {
				apply(t, p)
				if pcState(t, o).PiSubagentConfig != nil {
					t.Fatal("a manual step was recorded")
				}
				absent(t, pcConfig(o)+".hive-backup")
			}
		})
	}
}

// AC4: a record for another Pi home is kept and the new home is not written.
func TestPlanPiConfigRecordForAnotherHome(t *testing.T) {
	o := pcOptions(t)
	apply(t, plan(t, "install", o))
	elsewhere := filepath.Join(filepath.Dir(o.Home), "other-pi", "extensions", "subagent", "config.json")
	st := pcState(t, o)
	st.PiSubagentConfig.Path = elsewhere
	if err := writeJSON(filepath.Join(o.StateDir, "state.json"), st); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(pcConfig(o)); err != nil {
		t.Fatal(err)
	}
	p := plan(t, "install", o)
	got := p.PiSubagentConfig
	if got == nil || got.Status != PiConfigOtherHome || got.Detail != elsewhere {
		t.Fatalf("step = %#v", got)
	}
	apply(t, p)
	absent(t, pcConfig(o))
	if rec := pcState(t, o).PiSubagentConfig; rec == nil || rec.Path != elsewhere {
		t.Fatalf("the record was not kept: %#v", rec)
	}
}

func TestPlanPiConfigNeedsPiAndAUserInstall(t *testing.T) {
	o := pcOptions(t)
	o.Hosts = []string{"claude"}
	if st := plan(t, "install", o).PiSubagentConfig; st != nil {
		t.Fatalf("no Pi host, yet %#v", st)
	}
	o = pcOptions(t)
	if st := plan(t, "remove", o).PiSubagentConfig; st != nil {
		t.Fatalf("remove plans a write: %#v", st)
	}
}

// AC9: a release update never plans the step.
func TestPlanPiConfigReleaseUpdateDoesNotPlanTheStep(t *testing.T) {
	o := pcOptions(t)
	p := plan(t, "install", o)
	p.ReleaseUpdate = true
	if err := planPiSubagentConfig(&p, State{}); err != nil || p.PiSubagentConfig != nil {
		t.Fatalf("release update planned %#v (%v)", p.PiSubagentConfig, err)
	}
}

func TestPlanPiConfigHoldsNoFileBytes(t *testing.T) {
	o := pcOptions(t)
	put(t, pcConfig(o), `{"secretToken": "SUPERSECRET"}`)
	if strings.Contains(string(encode(plan(t, "install", o))), "SUPERSECRET") {
		t.Fatal("the plan holds config.json bytes")
	}
}

func TestValidatePlanRejectsAForgedPiConfigStep(t *testing.T) {
	o := pcOptions(t)
	for name, mutate := range map[string]func(p *Plan){
		"unknown status": func(p *Plan) { p.PiSubagentConfig.Status = "overwrite" },
		"other path":     func(p *Plan) { p.PiSubagentConfig.Path = filepath.Join(o.Home, "elsewhere.json") },
		"other key":      func(p *Plan) { p.PiSubagentConfig.Key = "forceTopLevelAsync" },
		"other value":    func(p *Plan) { p.PiSubagentConfig.Value = `"auto"` },
		"remove plan":    func(p *Plan) { p.Action = "remove" },
	} {
		t.Run(name, func(t *testing.T) {
			p := plan(t, "install", o)
			mutate(&p)
			p.ID = planID(p)
			if _, err := (Engine{}).Apply(p); err == nil {
				t.Fatal("forged step accepted")
			}
			absent(t, pcConfig(o))
		})
	}
}
