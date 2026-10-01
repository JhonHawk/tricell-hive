package pi

import (
	"encoding/json"
	"fmt"
	"strings"
)

// HiveSubagentsSource is the npm source Hive installs when the user settings
// do not already declare pi-subagents.
const HiveSubagentsSource = "npm:pi-subagents@0.74.0"

const subagentsNPMName = "pi-subagents"

// Status is how the planned Pi settings declare pi-subagents.
type Status int

const (
	// Absent means no packages entry names pi-subagents.
	Absent Status = iota
	// Present means the package is declared and its extensions filter is omitted,
	// so Pi loads the package's default resources.
	Present
	// Conflict means the package is declared but the extensions filter is empty
	// or otherwise does not demonstrably load the extension.
	Conflict
)

// Declaration is the classified pi-subagents entry, if any.
type Declaration struct {
	Status Status
	Source string
}

type packageEntry struct {
	Source     string
	Extensions *[]string
}

// ClassifySubagents reports how settings.json declares pi-subagents.
// data is the file bytes; a missing packages key is Absent, not an error.
func ClassifySubagents(data []byte) (Declaration, error) {
	var root struct {
		Packages json.RawMessage `json:"packages"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		return Declaration{}, fmt.Errorf("pi settings: %w", err)
	}
	if len(root.Packages) == 0 || string(root.Packages) == "null" {
		return Declaration{Status: Absent}, nil
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(root.Packages, &raw); err != nil {
		return Declaration{}, fmt.Errorf("pi settings packages: %w", err)
	}
	for _, item := range raw {
		entry, err := parsePackageEntry(item)
		if err != nil {
			return Declaration{}, err
		}
		if !isSubagentsSource(entry.Source) {
			continue
		}
		d := Declaration{Source: entry.Source, Status: Present}
		if entry.Extensions != nil {
			d.Status = Conflict
		}
		return d, nil
	}
	return Declaration{Status: Absent}, nil
}

func parsePackageEntry(item json.RawMessage) (packageEntry, error) {
	var asString string
	if err := json.Unmarshal(item, &asString); err == nil {
		return packageEntry{Source: asString}, nil
	}
	var obj struct {
		Source     string    `json:"source"`
		Extensions *[]string `json:"extensions"`
	}
	if err := json.Unmarshal(item, &obj); err != nil {
		return packageEntry{}, fmt.Errorf("pi settings package entry: %w", err)
	}
	return packageEntry{Source: obj.Source, Extensions: obj.Extensions}, nil
}

func isSubagentsSource(source string) bool {
	const prefix = "npm:" + subagentsNPMName
	if source == prefix {
		return true
	}
	return strings.HasPrefix(source, prefix+"@")
}
