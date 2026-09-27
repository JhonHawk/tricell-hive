// Package version defines the public Hive product version.
package version

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Current is replaced by release builds with -ldflags '-X tricell-hive/tooling/version.Current=...'.
var Current = "dev"

var productVersion = regexp.MustCompile(`^(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:-(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*)?$`)

// Valid accepts development builds and product versions with an optional SemVer prerelease.
func Valid(value string) bool {
	return value == "dev" || productVersion.MatchString(value)
}

// ReadSource reads the canonical product version selected for a source tree.
func ReadSource(root string) (string, error) {
	value, present, err := ReadSourceFile(root)
	if err != nil {
		return "", err
	}
	if !present {
		return "dev", nil
	}
	if !Valid(value) {
		return "", fmt.Errorf("invalid product version %q", value)
	}
	return value, nil
}

// ReadSourceFile reads a source tree's VERSION file without validating its
// content, so callers with different rules for an absent or empty value
// (ReadSource, productFromSource) can each apply their own. present is false
// only when the file does not exist; any other read error propagates unwrapped.
func ReadSourceFile(root string) (value string, present bool, err error) {
	data, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return strings.TrimSpace(string(data)), true, nil
}
