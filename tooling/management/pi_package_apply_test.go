package management

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/pi"
)

type piCall struct {
	Home, Action, Source string
}

var piCalls []piCall

func stubApplyPi(home, action, source string) error {
	piCalls = append(piCalls, piCall{home, action, source})
	path := filepath.Join(home, "settings.json")
	var root struct {
		Packages []string `json:"packages"`
	}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &root)
	}
	switch action {
	case "install":
		root.Packages = append(root.Packages, source)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}
		data, err := json.Marshal(root)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			return err
		}
		pkg := filepath.Join(home, "npm", "node_modules", "pi-subagents")
		if err := os.MkdirAll(pkg, 0700); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(pkg, "package.json"), []byte(`{"name":"pi-subagents","version":"0.74.0"}`), 0600)
	case "remove":
		var kept []string
		for _, s := range root.Packages {
			if s != source {
				kept = append(kept, s)
			}
		}
		root.Packages = kept
		data, err := json.Marshal(root)
		if err != nil {
			return err
		}
		return os.WriteFile(path, data, 0600)
	}
	return nil
}

func TestApplyInstallsPiSubagentsWhenAbsent(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"pi"}
	put(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"), `{"packages":["npm:other@1.0.0"]}`+"\n")
	apply(t, plan(t, "install", o))
	if len(piCalls) != 1 || piCalls[0].Action != "install" || piCalls[0].Source != pi.HiveSubagentsSource {
		t.Fatalf("piCalls = %#v", piCalls)
	}
	if piCalls[0].Home != filepath.Join(o.Home, ".pi", "agent") {
		t.Fatalf("PI home = %s", piCalls[0].Home)
	}
	d, err := pi.ClassifySubagents([]byte(get(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"))))
	if err != nil || d.Status != pi.Present || d.Source != pi.HiveSubagentsSource {
		t.Fatalf("settings after apply: %+v %v", d, err)
	}
	if !strings.Contains(get(t, filepath.Join(o.Home, ".pi", "agent", "settings.json")), "npm:other@1.0.0") {
		t.Fatal("lost unrelated package")
	}
	if _, err := os.Stat(filepath.Join(o.Home, ".pi", "agent", "npm", "node_modules", "pi-subagents", "package.json")); err != nil {
		t.Fatal(err)
	}
	if stateFor(t, o).PiSubagentsSource != pi.HiveSubagentsSource {
		t.Fatal("state did not record Hive source")
	}
}

func TestApplyOmitsPiInstallWhenPresent(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"pi"}
	put(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"), `{"packages":["npm:pi-subagents@0.67.0"]}`+"\n")
	apply(t, plan(t, "install", o))
	if len(piCalls) != 0 {
		t.Fatalf("omit invoked pi: %#v", piCalls)
	}
}

func TestApplyRejectsStalePiPackagePlan(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"pi"}
	p := plan(t, "install", o)
	put(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"), `{"packages":["npm:pi-subagents@0.67.0"]}`+"\n")
	_, err := (Engine{}).Apply(p)
	if err == nil || !strings.Contains(err.Error(), "stale plan") {
		t.Fatalf("want stale plan, got %v", err)
	}
	if len(piCalls) != 0 {
		t.Fatalf("stale plan invoked pi: %#v", piCalls)
	}
}

func TestRemovePiHostRemovesOnlyHiveAddedPackage(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"pi"}
	apply(t, plan(t, "install", o))
	piCalls = nil
	apply(t, plan(t, "remove", o))
	if len(piCalls) != 1 || piCalls[0].Action != "remove" || piCalls[0].Source != pi.HiveSubagentsSource {
		t.Fatalf("remove calls %#v", piCalls)
	}
	if stateFor(t, o).PiSubagentsSource != "" {
		t.Fatal("state still records Hive source")
	}
}

func TestRemovePiHostLeavesPreexistingPackage(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"pi"}
	put(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"), `{"packages":["npm:pi-subagents@0.67.0"]}`+"\n")
	apply(t, plan(t, "install", o))
	piCalls = nil
	apply(t, plan(t, "remove", o))
	if len(piCalls) != 0 {
		t.Fatalf("remove invoked pi: %#v", piCalls)
	}
	d, err := pi.ClassifySubagents([]byte(get(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"))))
	if err != nil || d.Source != "npm:pi-subagents@0.67.0" {
		t.Fatalf("user package lost: %+v %v", d, err)
	}
}

func TestRemoveSkipsChangedPin(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"pi"}
	apply(t, plan(t, "install", o))
	put(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"), `{"packages":["npm:pi-subagents@0.99.0"]}`+"\n")
	piCalls = nil
	apply(t, plan(t, "remove", o))
	if len(piCalls) != 0 {
		t.Fatalf("changed pin invoked pi: %#v", piCalls)
	}
	if stateFor(t, o).PiSubagentsSource != "" {
		t.Fatal("obsolete Hive receipt was kept")
	}
}

func TestUnchangedOmitRevalidatesDeclaration(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"pi"}
	put(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"), `{"packages":["npm:pi-subagents@0.67.0"]}`+"\n")
	apply(t, plan(t, "install", o))
	p := plan(t, "install", o)
	if p.PiPackage == nil || p.PiPackage.Action != PackageOmit {
		t.Fatalf("want omit, got %#v", p.PiPackage)
	}
	if err := os.Remove(filepath.Join(o.Home, ".pi", "agent", "settings.json")); err != nil {
		t.Fatal(err)
	}
	piCalls = nil
	_, err := (Engine{}).Apply(p)
	if err == nil || !strings.Contains(err.Error(), "no longer declared") {
		t.Fatalf("want stale omit on unchanged, got %v", err)
	}
}

func TestApplyRejectsStaleOmitPlan(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"pi"}
	put(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"), `{"packages":["npm:pi-subagents@0.67.0"]}`+"\n")
	p := plan(t, "install", o)
	if p.PiPackage == nil || p.PiPackage.Action != PackageOmit {
		t.Fatalf("want omit, got %#v", p.PiPackage)
	}
	if err := os.Remove(filepath.Join(o.Home, ".pi", "agent", "settings.json")); err != nil {
		t.Fatal(err)
	}
	_, err := (Engine{}).Apply(p)
	if err == nil || !strings.Contains(err.Error(), "no longer declared") {
		t.Fatalf("want stale omit, got %v", err)
	}
	if len(piCalls) != 0 {
		t.Fatalf("stale omit invoked pi: %#v", piCalls)
	}
}

func TestInterruptedAfterPiInstallRecoverRemovesPackage(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"pi"}
	p := plan(t, "install", o)
	_, err := (Engine{failpoint: func(stage string) error {
		if stage == "pi-package-after" {
			return errors.New("injected")
		}
		return nil
	}}).Apply(p)
	if err == nil {
		t.Fatal("expected failpoint")
	}
	if len(piCalls) != 1 || piCalls[0].Action != "install" {
		t.Fatalf("install calls %#v", piCalls)
	}
	piCalls = nil
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if len(piCalls) != 1 || piCalls[0].Action != "remove" {
		t.Fatalf("recover calls %#v", piCalls)
	}
}

func TestInterruptedPiInstallRecoverDoesNotRemove(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"pi"}
	p := plan(t, "install", o)
	_, err := (Engine{failpoint: func(stage string) error {
		if stage == "pi-package" {
			return errors.New("injected")
		}
		return nil
	}}).Apply(p)
	if err == nil {
		t.Fatal("expected failpoint")
	}
	if len(piCalls) != 0 {
		t.Fatalf("interrupted install invoked pi: %#v", piCalls)
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if len(piCalls) != 0 {
		t.Fatalf("recover invoked pi: %#v", piCalls)
	}
}
