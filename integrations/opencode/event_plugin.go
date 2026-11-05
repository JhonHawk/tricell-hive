package opencode

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"tricell-hive/integrations/target"
)

// EventPluginName is the plugin file Hive owns in <OpenCodeHome>/plugins, which
// OpenCode loads on its own and reloads when the file changes. The context
// loader under <OpenCodeHome>/hive stays inactive: nothing here writes
// opencode.json.
const EventPluginName = "hive-events.js"

// EventPluginSource names the generated payload of the event plugin. It is not
// a file of content/: the manager renders it from the template below and the
// path of the hive binary.
const EventPluginSource = "integrations/opencode/event-plugin"

//go:embed event-plugin.js
var eventPluginTemplate string

const executablePlaceholder = "__HIVE_EXECUTABLE__"

// EventPluginPath returns where the event plugin lives for an OpenCode home.
func EventPluginPath(openCodeHome string) string {
	return filepath.Join(openCodeHome, "plugins", EventPluginName)
}

// EventPluginTarget is the plugin's destination for c, shaped like the other
// resolved targets; the manager fills in the consumer fields.
func EventPluginTarget(c target.Config, base target.Target) target.Target {
	base.Kind, base.Source, base.Path = "hook", EventPluginSource, EventPluginPath(c.OpenCodeHome)
	return base
}

// RenderEventPlugin returns the plugin with the hive binary written in as a JSON
// string, which is also a valid JavaScript string literal. The path comes from a
// plan that may have been edited by hand, so it is checked here as well.
func RenderEventPlugin(executable string) ([]byte, error) {
	if !filepath.IsAbs(executable) || strings.ContainsRune(executable, 0) {
		return nil, fmt.Errorf("OpenCode event plugin executable must be an absolute path without NUL")
	}
	var quoted bytes.Buffer
	enc := json.NewEncoder(&quoted)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(executable); err != nil {
		return nil, err
	}
	literal := strings.TrimSuffix(quoted.String(), "\n")
	// U+2028 and U+2029 are legal in JSON strings and were not in JavaScript's.
	literal = strings.NewReplacer(" ", ` `, " ", ` `).Replace(literal)
	if strings.Count(eventPluginTemplate, executablePlaceholder) != 1 {
		return nil, fmt.Errorf("OpenCode event plugin template must hold the executable placeholder once")
	}
	return []byte(strings.Replace(eventPluginTemplate, executablePlaceholder, literal, 1)), nil
}
