package management

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/pi"
)

func TestPlanPiSubagentsStep(t *testing.T) {
	t.Run("install when absent", func(t *testing.T) {
		o := setup(t)
		o.Hosts = []string{"pi"}
		p := plan(t, "install", o)
		if p.PiPackage == nil || p.PiPackage.Action != PackageInstall || p.PiPackage.Source != pi.HiveSubagentsSource || !p.PiPackage.AddedByHive {
			t.Fatalf("PiPackage = %#v", p.PiPackage)
		}
	})
	t.Run("omit when present", func(t *testing.T) {
		o := setup(t)
		o.Hosts = []string{"pi"}
		put(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"), `{"packages":["npm:pi-subagents@0.67.0"]}`+"\n")
		p := plan(t, "install", o)
		if p.PiPackage == nil || p.PiPackage.Action != PackageOmit || p.PiPackage.Source != "npm:pi-subagents@0.67.0" || p.PiPackage.AddedByHive {
			t.Fatalf("PiPackage = %#v", p.PiPackage)
		}
	})
	t.Run("conflict empty extensions", func(t *testing.T) {
		o := setup(t)
		o.Hosts = []string{"pi"}
		put(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"), `{"packages":[{"source":"npm:pi-subagents@0.74.0","extensions":[]}]}`+"\n")
		_, err := BuildPlan("install", o)
		if err == nil || !strings.Contains(err.Error(), "extensions: []") {
			t.Fatalf("want conflict naming extensions: [], got %v", err)
		}
	})
	t.Run("no step without pi host", func(t *testing.T) {
		o := setup(t)
		p := plan(t, "install", o)
		if p.PiPackage != nil {
			t.Fatalf("unexpected PiPackage %#v", p.PiPackage)
		}
	})
	t.Run("no step on remove", func(t *testing.T) {
		o := setup(t)
		o.Hosts = []string{"pi"}
		apply(t, plan(t, "install", o))
		p := plan(t, "remove", o)
		if p.PiPackage != nil {
			t.Fatalf("remove PiPackage %#v", p.PiPackage)
		}
	})
}

func TestPlanPiSubagentsMissingSettingsFile(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"pi"}
	if err := os.MkdirAll(filepath.Join(o.Home, ".pi", "agent"), 0700); err != nil {
		t.Fatal(err)
	}
	p := plan(t, "install", o)
	if p.PiPackage == nil || p.PiPackage.Action != PackageInstall {
		t.Fatalf("missing file: %#v", p.PiPackage)
	}
}
