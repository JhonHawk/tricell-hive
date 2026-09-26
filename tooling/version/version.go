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
	data, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if os.IsNotExist(err) {
		return "dev", nil
	}
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(data))
	if !Valid(value) {
		return "", fmt.Errorf("invalid product version %q", value)
	}
	return value, nil
}
