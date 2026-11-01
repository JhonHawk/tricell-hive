package main

import "runtime/debug"

// The executable's build metadata identifies the served code, not the current
// checkout or a binary path which may have been replaced since startup.
func viewBuildIdentity() (string, bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown", false
	}
	return viewBuildSettings(info.Settings)
}

func viewBuildSettings(settings []debug.BuildSetting) (string, bool) {
	revision, modified := "unknown", false
	for _, setting := range settings {
		switch setting.Key {
		case "vcs.revision":
			if setting.Value != "" {
				revision = setting.Value
			}
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	return revision, modified
}
