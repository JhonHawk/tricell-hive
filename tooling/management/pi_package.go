package management

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"tricell-hive/integrations/pi"
)

// applyPiCmd, when non-nil, replaces the production runner. Tests use it to
// record calls. A nil value selects syntheticApplyPi on a synthetic home so
// CLI tests never invoke the real pi binary.
var applyPiCmd func(piHome, action, source string) error

func resolveApplyPi(synthetic bool) func(string, string, string) error {
	if applyPiCmd != nil {
		return applyPiCmd
	}
	if synthetic {
		return syntheticApplyPi
	}
	return execApplyPi
}

func execApplyPi(piHome, action, source string) error {
	bin, err := exec.LookPath("pi")
	if err != nil {
		return fmt.Errorf("pi executable: %w", err)
	}
	cmd := exec.Command(bin, action, source)
	cmd.Env = append(os.Environ(), "PI_CODING_AGENT_DIR="+piHome)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pi %s: %w", action, err)
	}
	return nil
}

func syntheticApplyPi(home, action, source string) error {
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
	return fmt.Errorf("unsupported pi package action %q", action)
}

func hostsIncludePi(hosts []string) bool {
	for _, h := range hosts {
		if h == "pi" {
			return true
		}
	}
	return false
}

func revalidatePiPackage(p Plan) error {
	if p.PiPackage == nil || p.PiPackage.Action != PackageOmit {
		return nil
	}
	d, err := readSubagents(p.Config.PiHome)
	if err != nil {
		return err
	}
	if d.Status != pi.Present {
		return fmt.Errorf("stale plan: pi-subagents is no longer declared")
	}
	return nil
}

func applyPiPackage(p Plan, next *State, j *journal, jp string, fail func(string) error) error {
	run := resolveApplyPi(p.Config.Synthetic)
	var action, source string
	switch {
	case p.Action == "remove" && hostsIncludePi(p.Hosts) && next.PiSubagentsSource != "":
		d, err := readSubagents(p.Config.PiHome)
		if err != nil {
			return err
		}
		if d.Status == pi.Absent || d.Source != next.PiSubagentsSource {
			next.PiSubagentsSource = ""
			j.AfterState = snapshot{Exists: true, Data: encode(*next), Mode: 0600}
			return saveJournal(jp, *j)
		}
		action, source = "remove", next.PiSubagentsSource
	case p.PiPackage != nil && p.PiPackage.Action == PackageOmit:
		return revalidatePiPackage(p)
	case p.PiPackage != nil && p.PiPackage.Action == PackageInstall:
		if err := revalidateAbsent(p.Config.PiHome); err != nil {
			return err
		}
		action, source = "install", p.PiPackage.Source
	default:
		return nil
	}
	j.PackagePending = true
	if err := saveJournal(jp, *j); err != nil {
		return err
	}
	if fail != nil {
		if err := fail("pi-package-run"); err != nil {
			return err
		}
	}
	if err := run(p.Config.PiHome, action, source); err != nil {
		return err
	}
	j.PackagePending = false
	j.PackageDone = true
	if action == "install" {
		d, err := readSubagents(p.Config.PiHome)
		if err != nil {
			return err
		}
		if d.Status != pi.Present || d.Source != source {
			return fmt.Errorf("pi-subagents install did not declare %s", source)
		}
		next.PiSubagentsSource = source
	} else {
		next.PiSubagentsSource = ""
	}
	j.AfterState = snapshot{Exists: true, Data: encode(*next), Mode: 0600}
	if err := saveJournal(jp, *j); err != nil {
		return err
	}
	if fail != nil {
		return fail("pi-package-after")
	}
	return nil
}

func undoPiPackage(j journal) error {
	p := j.Plan
	run := resolveApplyPi(p.Config.Synthetic)
	d, err := readSubagents(p.Config.PiHome)
	if err != nil {
		return err
	}
	if p.PiPackage != nil && p.PiPackage.Action == PackageInstall {
		if d.Status == pi.Present && d.Source == p.PiPackage.Source {
			return run(p.Config.PiHome, "remove", p.PiPackage.Source)
		}
		return nil
	}
	if p.Action == "remove" && hostsIncludePi(p.Hosts) {
		var before State
		if len(j.BeforeState.Data) == 0 {
			return nil
		}
		if err := json.Unmarshal(j.BeforeState.Data, &before); err != nil {
			return err
		}
		if before.PiSubagentsSource == "" || d.Status != pi.Absent {
			return nil
		}
		return run(p.Config.PiHome, "install", before.PiSubagentsSource)
	}
	return nil
}

func revalidateAbsent(piHome string) error {
	d, err := readSubagents(piHome)
	if err != nil {
		return err
	}
	if d.Status != pi.Absent {
		return fmt.Errorf("stale plan: pi-subagents is already declared")
	}
	return nil
}

func readSubagents(piHome string) (pi.Declaration, error) {
	data, err := os.ReadFile(filepath.Join(piHome, "settings.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return pi.Declaration{Status: pi.Absent}, nil
		}
		return pi.Declaration{}, err
	}
	return pi.ClassifySubagents(data)
}
