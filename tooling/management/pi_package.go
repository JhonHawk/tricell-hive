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

func applyPiPackage(p Plan, next *State) error {
	run := resolveApplyPi(p.Config.Synthetic)
	if p.Action == "remove" && hostsIncludePi(p.Hosts) && next.PiSubagentsSource != "" {
		return removeHivePiPackage(p.Config.PiHome, next, run)
	}
	if p.PiPackage == nil {
		return nil
	}
	if p.PiPackage.Action == PackageOmit {
		d, err := readSubagents(p.Config.PiHome)
		if err != nil {
			return err
		}
		if d.Status != pi.Present {
			return fmt.Errorf("stale plan: pi-subagents is no longer declared")
		}
		return nil
	}
	if p.PiPackage.Action != PackageInstall {
		return nil
	}
	if err := revalidateAbsent(p.Config.PiHome); err != nil {
		return err
	}
	if err := run(p.Config.PiHome, "install", p.PiPackage.Source); err != nil {
		return err
	}
	d, err := readSubagents(p.Config.PiHome)
	if err != nil {
		return err
	}
	if d.Status != pi.Present || d.Source != p.PiPackage.Source {
		return fmt.Errorf("pi-subagents install did not declare %s", p.PiPackage.Source)
	}
	next.PiSubagentsSource = p.PiPackage.Source
	return nil
}

func removeHivePiPackage(piHome string, next *State, run func(string, string, string) error) error {
	d, err := readSubagents(piHome)
	if err != nil {
		return err
	}
	if d.Status == pi.Absent {
		next.PiSubagentsSource = ""
		return nil
	}
	if d.Source != next.PiSubagentsSource {
		next.PiSubagentsSource = ""
		return nil
	}
	if err := run(piHome, "remove", next.PiSubagentsSource); err != nil {
		return err
	}
	next.PiSubagentsSource = ""
	return nil
}

func undoPiPackage(j journal) error {
	p := j.Plan
	run := resolveApplyPi(p.Config.Synthetic)
	if p.PiPackage != nil && p.PiPackage.Action == PackageInstall {
		return run(p.Config.PiHome, "remove", p.PiPackage.Source)
	}
	if p.Action == "remove" && hostsIncludePi(p.Hosts) {
		var before State
		if len(j.BeforeState.Data) == 0 {
			return nil
		}
		if err := json.Unmarshal(j.BeforeState.Data, &before); err != nil {
			return err
		}
		if before.PiSubagentsSource == "" {
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
