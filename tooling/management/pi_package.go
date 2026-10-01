package management

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"tricell-hive/integrations/pi"
)

// applyPiCmd runs Pi's package command. Tests replace it with a fake.
var applyPiCmd = defaultApplyPi

func defaultApplyPi(piHome, action, source string) error {
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

func hostsIncludePi(hosts []string) bool {
	for _, h := range hosts {
		if h == "pi" {
			return true
		}
	}
	return false
}

func applyPiPackage(p Plan, next *State) error {
	if p.Action == "remove" && hostsIncludePi(p.Hosts) && next.PiSubagentsSource != "" {
		return removeHivePiPackage(p.Config.PiHome, next)
	}
	if p.PiPackage == nil || p.PiPackage.Action != PackageInstall {
		return nil
	}
	if err := revalidateAbsent(p.Config.PiHome); err != nil {
		return err
	}
	if err := applyPiCmd(p.Config.PiHome, "install", p.PiPackage.Source); err != nil {
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

func removeHivePiPackage(piHome string, next *State) error {
	d, err := readSubagents(piHome)
	if err != nil {
		return err
	}
	if d.Status == pi.Absent {
		next.PiSubagentsSource = ""
		return nil
	}
	if d.Source != next.PiSubagentsSource {
		return nil
	}
	if err := applyPiCmd(piHome, "remove", next.PiSubagentsSource); err != nil {
		return err
	}
	next.PiSubagentsSource = ""
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
