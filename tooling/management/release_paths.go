package management

import (
	"os"
	"path/filepath"
	"tricell-hive/integrations/target"
)

func releaseHome(home string) (string, error) {
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	return target.Canonical(home)
}

// ReleasePackageStore selects the user package store without creating it. An
// explicit synthetic home is isolated from the calling user's XDG settings.
// Canonicalize only the selected root: downstream checks reject linked store
// descendants instead of silently redirecting writes through them.
func ReleasePackageStore(home string) (string, error) {
	explicit := home != ""
	root, err := releaseHome(home)
	if err != nil {
		return "", err
	}
	data := filepath.Join(root, ".local", "share")
	if !explicit && filepath.IsAbs(os.Getenv("XDG_DATA_HOME")) {
		data, err = target.Canonical(os.Getenv("XDG_DATA_HOME"))
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(data, "hive", "packages"), nil
}

// ManagedCommandPath selects the stable command without editing PATH or
// resolving an existing command link into the retained executable.
func ManagedCommandPath(home string) (string, error) {
	root, err := releaseHome(home)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".local", "bin", "hive"), nil
}
