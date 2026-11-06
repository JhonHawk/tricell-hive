package pi

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"tricell-hive/integrations/target"
)

// EventExtensionName is the extension file Hive owns in <PiHome>/extensions,
// which Pi loads on its own.
const EventExtensionName = "hive-events.ts"

// EventExtensionSource names the generated payload of the event extension. It
// is not a file of content/: the manager renders it from the template below and
// the path of the hive binary.
const EventExtensionSource = "integrations/pi/event-extension"

//go:embed event-extension.ts
var eventExtensionTemplate string

const executablePlaceholder = "__HIVE_EXECUTABLE__"

// EventExtensionPath returns where the event extension lives for a Pi home.
func EventExtensionPath(piHome string) string {
	return filepath.Join(piHome, "extensions", EventExtensionName)
}

// EventExtensionTarget is the extension's destination for c, shaped like the
// other resolved targets; the manager fills in the consumer fields.
func EventExtensionTarget(c target.Config, base target.Target) target.Target {
	base.Kind, base.Source, base.Path = "hook", EventExtensionSource, EventExtensionPath(c.PiHome)
	return base
}

// RenderEventExtension returns the extension with the hive binary written in as
// a JSON string, which is also a valid TypeScript string literal. The path comes
// from a plan that may have been edited by hand, so it is checked here as well.
func RenderEventExtension(executable string) ([]byte, error) {
	if !filepath.IsAbs(executable) || strings.ContainsRune(executable, 0) {
		return nil, fmt.Errorf("Pi event extension executable must be an absolute path without NUL")
	}
	var quoted bytes.Buffer
	enc := json.NewEncoder(&quoted)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(executable); err != nil {
		return nil, err
	}
	literal := strings.TrimSuffix(quoted.String(), "\n")
	literal = strings.NewReplacer(" ", ` `, " ", ` `).Replace(literal)
	if strings.Count(eventExtensionTemplate, executablePlaceholder) != 1 {
		return nil, fmt.Errorf("Pi event extension template must hold the executable placeholder once")
	}
	return []byte(strings.Replace(eventExtensionTemplate, executablePlaceholder, literal, 1)), nil
}
