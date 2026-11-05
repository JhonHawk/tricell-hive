package management

import (
	"fmt"
	"path/filepath"

	"tricell-hive/integrations/grok"
	"tricell-hive/integrations/opencode"
	"tricell-hive/integrations/pi"
	"tricell-hive/integrations/target"
)

// A generated hook is a file Hive owns whole, written from the path of the hive
// binary instead of from a file of content/. All of them are records of kind
// "hook": Grok's two native hook files, OpenCode's event plugin and Pi's event
// extension. The kind names what the file is for (a host calls Hive on events),
// not its format. The shared JSON files of Claude Code, Codex and Cursor are
// not records: they hold other tools' entries too, so host_hooks.go edits them
// entry by entry.
type generatedHook struct {
	// Host is the only host that may consume the file.
	Host string
	// Name is the file name and Dir the name of its folder.
	Name, Dir string
	// Path says where the file lives for a configuration.
	Path func(target.Config) string
}

func generatedHooks() map[string]generatedHook {
	return map[string]generatedHook{
		grok.ReadHookSource:        {Host: "grok", Name: grok.ReadHookName, Dir: "hooks", Path: func(c target.Config) string { return grok.ReadHookPath(c.GrokHome) }},
		grok.SessionEventsSource:   {Host: "grok", Name: grok.SessionEventsName, Dir: "hooks", Path: func(c target.Config) string { return grok.SessionEventsPath(c.GrokHome) }},
		opencode.EventPluginSource: {Host: "opencode", Name: opencode.EventPluginName, Dir: "plugins", Path: func(c target.Config) string { return opencode.EventPluginPath(c.OpenCodeHome) }},
		pi.EventExtensionSource:    {Host: "pi", Name: pi.EventExtensionName, Dir: "extensions", Path: func(c target.Config) string { return pi.EventExtensionPath(c.PiHome) }},
	}
}

// generatedSource reports whether source names a payload the manager generates.
func generatedSource(source string) bool {
	_, ok := generatedHooks()[source]
	return ok
}

// validRecordedHook checks a recorded hook's shape: a known source, its file
// name in its folder.
func validRecordedHook(source, path string) bool {
	g, ok := generatedHooks()[source]
	return ok && filepath.Base(path) == g.Name && filepath.Base(filepath.Dir(path)) == g.Dir
}

// renderGeneratedHook returns the bytes of the hook file for a resolved target.
func renderGeneratedHook(t target.Target, p Plan) ([]byte, error) {
	g, ok := generatedHooks()[t.Source]
	if !ok || p.Config.Scope != "user" || t.Path != g.Path(p.Config) {
		return nil, fmt.Errorf("invalid %s hook resource", t.Host)
	}
	if err := checkHookExecutable(p.Executable); err != nil {
		return nil, err
	}
	switch t.Source {
	case grok.ReadHookSource:
		return grok.RenderReadHook(p.Executable, skillsDirFor("grok", p.Config))
	case grok.SessionEventsSource:
		return grok.RenderSessionEventsHook(p.Executable)
	case opencode.EventPluginSource:
		return opencode.RenderEventPlugin(p.Executable)
	default:
		return pi.RenderEventExtension(p.Executable)
	}
}
