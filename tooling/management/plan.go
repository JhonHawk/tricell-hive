package management

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"tricell-hive/integrations/claude"
	"tricell-hive/integrations/codex"
	"tricell-hive/integrations/grok"
	"tricell-hive/integrations/opencode"
	"tricell-hive/integrations/pi"
	"tricell-hive/integrations/target"
	"unicode/utf8"
)

func resolve(c target.Config, hosts []string) ([]target.Target, error) {
	var out []target.Target
	for _, h := range hosts {
		var ts []target.Target
		var err error
		switch h {
		case "codex":
			ts, err = codex.Resolve(c)
		case "claude":
			ts, err = claude.Resolve(c)
		case "grok":
			ts, err = grok.Resolve(c)
		case "pi":
			ts, err = pi.Resolve(c)
		case "opencode":
			ts, err = opencode.Resolve(c)
		default:
			return nil, fmt.Errorf("unsupported host %q", h)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, ts...)
	}
	seen := map[string]bool{}
	for _, t := range out {
		if t.Kind == "symlink" {
			if t.LinkTarget == "" || filepath.IsAbs(t.LinkTarget) {
				return nil, fmt.Errorf("invalid relative link target")
			}
			if err := target.Safe(filepath.Dir(t.Path)); err != nil {
				return nil, err
			}
			if err := target.Safe(filepath.Join(filepath.Dir(t.Path), t.LinkTarget)); err != nil {
				return nil, err
			}
		} else if err := target.Safe(t.Path); err != nil {
			return nil, err
		}
		seen[t.Path] = true
	}
	if _, err := groups(out); err != nil {
		return nil, err
	}
	// Native inheritance may share a physical block; explicit imports still must
	// not inject a second selected instruction file.
	imports := regexp.MustCompile("@([^\\s`]+)")
	checked := map[string]bool{}
	for _, t := range out {
		if t.Kind != "block" || checked[t.Path] {
			continue
		}
		checked[t.Path] = true
		s, err := read(t.Path)
		if err != nil {
			return nil, err
		}
		for _, m := range imports.FindAllSubmatch(s.Data, -1) {
			path := string(m[1])
			if strings.HasPrefix(path, "~/") {
				path = filepath.Join(c.Home, path[2:])
			}
			if !filepath.IsAbs(path) {
				path = filepath.Join(filepath.Dir(t.Path), path)
			}
			if seen[filepath.Clean(path)] {
				return nil, fmt.Errorf("shared instruction import conflict: %s", t.Path)
			}
		}
	}
	return out, nil
}
func loadRelease(o Options, stateDir string) (Release, error) {
	var r Release
	if o.ReleaseID != "" {
		if len(o.ReleaseID) != 64 || strings.ContainsAny(o.ReleaseID, "/\\") {
			return r, fmt.Errorf("invalid release ID")
		}
		if err := decodeFile(filepath.Join(stateDir, "releases", o.ReleaseID+".json"), &r); err != nil {
			return r, err
		}
	} else {
		for _, p := range []string{GlobalSource, SkillSource} {
			s, err := read(filepath.Join(o.Source, p))
			if err != nil {
				return r, err
			}
			if !s.Exists {
				return r, fmt.Errorf("missing source %s", p)
			}
			r.Files = append(r.Files, Payload{p, s.Data})
		}
		r.ID = releaseID(r)
	}
	return r, validateRelease(r)
}
func validateRelease(r Release) error {
	if r.ID != releaseID(r) || len(r.Files) != 2 {
		return fmt.Errorf("invalid release fingerprint or file list")
	}
	for i, p := range []string{GlobalSource, SkillSource} {
		f := r.Files[i]
		if f.Path != p || !utf8.Valid(f.Data) || len(bytes.TrimSpace(f.Data)) == 0 {
			return fmt.Errorf("invalid source %s", p)
		}
		if bytes.Contains(f.Data, []byte(Begin)) || bytes.Contains(f.Data, []byte(End)) {
			return fmt.Errorf("reserved source delimiter")
		}
	}
	return nil
}
func BuildPlan(action string, o Options) (Plan, error) {
	var p Plan
	if action != "install" && action != "remove" {
		return p, fmt.Errorf("action must be install or remove")
	}
	hosts, err := validateHosts(o.Hosts)
	if err != nil {
		return p, err
	}
	c, dir, err := normalize(o)
	if err != nil {
		return p, err
	}
	if pending, err := read(filepath.Join(dir, "pending.json")); err != nil {
		return p, err
	} else if pending.Exists {
		return p, fmt.Errorf("unfinished operation: recover first")
	}
	state, sh, err := readState(dir)
	if err != nil {
		return p, err
	}
	gs, err := desiredResources(c, hosts, state, action)
	if err != nil {
		return p, err
	}
	p = Plan{Version: 2, Action: action, Config: c, Hosts: hosts, StateDir: dir, StateHash: sh}
	if action == "install" {
		r, err := loadRelease(o, dir)
		if err != nil {
			return p, err
		}
		p.Release = &r
	}
	for _, g := range gs {
		t := g.Target
		s, err := readResource(t, g.Replaces != nil)
		if err != nil {
			return p, err
		}
		ch := Change{Target: t, Expected: finger(s), Replaces: g.Replaces}
		if old, ok := state.Records[t.Path]; ok {
			if old.Target != t {
				return p, fmt.Errorf("resource owned by a different target")
			}
			ch.Before = &old
		}
		if action == "remove" && ch.Before == nil && s.Exists {
			a, _, parseErr := blockRange(s.Data)
			if t.Kind != "block" || parseErr != nil || a >= 0 {
				return p, fmt.Errorf("unowned resource preserved: %s", t.Path)
			}
		}
		if action == "install" && ch.Before == nil && t.Kind == "skill" {
			entries, err := os.ReadDir(filepath.Dir(t.Path))
			if err == nil && len(entries) > 0 {
				return p, fmt.Errorf("unowned skill directory: %s", filepath.Dir(t.Path))
			}
			if err != nil && !os.IsNotExist(err) {
				return p, err
			}
		}
		ch.After, err = nextRecord(p, g, ch.Before, s)
		if err != nil {
			return p, err
		}
		if _, err = transformResource(s, ch); err != nil {
			return p, fmt.Errorf("%s: %w", t.Path, err)
		}
		p.Changes = append(p.Changes, ch)
	}
	p.ID = planID(p)
	return p, nil
}

func nextRecord(p Plan, g resource, old *Record, s snapshot) (*Record, error) {
	if p.Action == "remove" {
		if old == nil {
			return nil, nil
		}
		next := *old
		next.Consumers = subtractConsumers(old.Consumers, g.Consumers)
		if len(next.Consumers) == 0 {
			return nil, nil
		}
		return &next, nil
	}
	r := Record{Target: g.Target, Release: p.Release.ID, CreatedFile: !s.Exists, Consumers: g.Consumers}
	if old != nil {
		r.CreatedFile = old.CreatedFile
		r.Leading = old.Leading
		r.Consumers = sortedConsumers(append(append([]Consumer{}, old.Consumers...), g.Consumers...))
	}
	switch g.Target.Kind {
	case "block":
		r.Managed = managedBlock(p.Release.Files[0].Data, s.Data)
		if old == nil && len(s.Data) > 0 && !bytes.HasSuffix(s.Data, []byte("\n")) {
			r.Leading = "\n"
		}
	case "skill":
		r.Managed = p.Release.Files[1].Data
	case "symlink":
		r.CreatedFile = true
	default:
		return nil, fmt.Errorf("unsupported resource kind")
	}
	if old != nil && len(subtractConsumers(old.Consumers, g.Consumers)) > 0 {
		if !bytes.Equal(old.Managed, r.Managed) || old.Target.LinkTarget != r.Target.LinkTarget {
			return nil, fmt.Errorf("shared resource update requires all consumers in --hosts: %s", g.Target.Path)
		}
		r.Release = old.Release
	}
	return &r, nil
}
func SavePlan(path string, p Plan) error {
	if err := target.Safe(path); err != nil {
		return err
	}
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("plan output already exists")
	} else if !os.IsNotExist(err) {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(encode(p))
	ce := f.Close()
	if err == nil {
		err = ce
	}
	return err
}
func LoadPlan(path string) (Plan, error) {
	var p Plan
	err := decodeFile(path, &p)
	if err != nil {
		return p, err
	}
	if p.Version != 2 || p.ID != planID(p) {
		return p, fmt.Errorf("invalid or legacy plan; regenerate with the current manager")
	}
	return p, nil
}
func validatePlan(p Plan, state State) error {
	if p.Version != 2 || p.ID != planID(p) {
		return fmt.Errorf("invalid or legacy plan; regenerate with the current manager")
	}
	h, err := validateHosts(p.Hosts)
	if err != nil || !reflect.DeepEqual(h, p.Hosts) {
		return fmt.Errorf("invalid plan hosts")
	}
	if p.Config.Scope != "user" && p.Config.Scope != "project" {
		return fmt.Errorf("invalid plan scope")
	}
	for _, path := range []string{p.Config.Home, p.Config.CodexHome, p.Config.ClaudeHome, p.Config.GrokHome, p.Config.PiHome, p.Config.OpenCodeHome, p.StateDir} {
		if path == "" || !filepath.IsAbs(path) || filepath.Clean(path) != path {
			return fmt.Errorf("invalid root")
		}
		if err = target.Safe(path); err != nil {
			return err
		}
	}
	if p.Config.Scope == "project" && (!filepath.IsAbs(p.Config.Root) || filepath.Clean(p.Config.Root) != p.Config.Root) {
		return fmt.Errorf("invalid project root")
	}
	if p.Action == "install" {
		if p.Release == nil {
			return fmt.Errorf("release missing")
		}
		if err = validateRelease(*p.Release); err != nil {
			return err
		}
	} else if p.Action != "remove" || p.Release != nil {
		return fmt.Errorf("invalid plan action")
	}
	gs, err := desiredResources(p.Config, p.Hosts, state, p.Action)
	if err != nil {
		return err
	}
	if len(gs) != len(p.Changes) {
		return fmt.Errorf("changed target resolution")
	}
	for i, g := range gs {
		ch := p.Changes[i]
		if ch.Target != g.Target || !reflect.DeepEqual(ch.Replaces, g.Replaces) {
			return fmt.Errorf("changed target resolution")
		}
		var old *Record
		if r, ok := state.Records[g.Target.Path]; ok {
			old = &r
		}
		if !reflect.DeepEqual(old, ch.Before) {
			return fmt.Errorf("stale ownership")
		}
		// Reconstruct consumer and release decisions independently of the proposed
		// records. Preserve the frozen newline choice and verify it against payload.
		s := snapshot{Exists: ch.Expected.Exists}
		if ch.After != nil && bytes.Contains(ch.After.Managed, []byte("\r\n")) {
			s.Data = []byte("\r\n")
		}
		expected, err := nextRecord(p, g, old, s)
		if err != nil {
			return err
		}
		if expected != nil && ch.After != nil {
			if old == nil {
				expected.Leading = ch.After.Leading
			}
			if expected.Leading != "" && expected.Leading != "\n" {
				return fmt.Errorf("invalid separator")
			}
		}
		if !reflect.DeepEqual(expected, ch.After) {
			return fmt.Errorf("invalid ownership or release payload")
		}
		// Fresh reads are for metadata checking only. During Apply some resources
		// already contain their after-image, so never infer new ownership here.
		current, err := readResource(g.Target, g.Replaces != nil)
		if err != nil {
			return err
		}
		if finger(current) == ch.Expected && ch.After != nil && old == nil && g.Target.Kind == "block" {
			leading := ""
			if len(current.Data) > 0 && !bytes.HasSuffix(current.Data, []byte("\n")) {
				leading = "\n"
			}
			if ch.After.Leading != leading {
				return fmt.Errorf("invalid initial ownership")
			}
		}
	}
	return nil
}
func Status(o Options) ([]StatusEntry, error) {
	h, err := validateHosts(o.Hosts)
	if err != nil {
		return nil, err
	}
	c, dir, err := normalize(o)
	if err != nil {
		return nil, err
	}
	state, _, err := readState(dir)
	if err != nil {
		return nil, err
	}
	ts, err := resolve(c, h)
	if err != nil {
		return nil, err
	}
	pending, err := read(filepath.Join(dir, "pending.json"))
	if err != nil {
		return nil, err
	}
	var out []StatusEntry
	seen := map[string]bool{}
	for _, t := range ts {
		seen[t.Path] = true
		en := StatusEntry{Path: t.Path, Host: t.Host, Kind: t.Kind, Status: "not_installed"}
		migration := false
		if _, ok := state.Records[filepath.Join(t.Path, "SKILL.md")]; t.Kind == "symlink" && ok {
			migration = true
		}
		s, readErr := readResource(t, migration)
		if r, ok := state.Records[t.Path]; ok {
			en.Release = r.Release
			en.Consumers = r.Consumers
			if hasConsumer(r.Consumers, consumer(t)) {
				en.Status = "installed"
			} else {
				en.Status = "retained_shared"
			}
			if readErr != nil || owned(s, r) != nil {
				en.Status = "drift"
			}
		} else if readErr != nil {
			en.Status = "unowned_or_conflicting"
		} else if t.Kind == "symlink" && migration {
			en.Status = "migration_required"
		} else if t.Kind != "block" && s.Exists {
			en.Status = "unowned"
		} else if t.Kind == "block" {
			a, _, e := blockRange(s.Data)
			if e != nil || a >= 0 {
				en.Status = "unowned_or_conflicting"
			}
		}
		if pending.Exists {
			en.Status = "recovery_required"
		}
		out = append(out, en)
	}
	for _, r := range state.Records {
		if seen[r.Target.Path] {
			continue
		}
		for _, binding := range r.Consumers {
			if binding.Scope != c.Scope || (binding.Scope == "user" && binding.Context != c.Home) || (binding.Scope == "project" && binding.Context != c.Root) {
				continue
			}
			selected := false
			for _, host := range h {
				if host == binding.Host {
					selected = true
				}
			}
			if selected {
				out = append(out, StatusEntry{Path: r.Target.Path, Host: binding.Host, Kind: r.Target.Kind, Status: "shadowed", Release: r.Release, Consumers: r.Consumers})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path == out[j].Path {
			return out[i].Host < out[j].Host
		}
		return out[i].Path < out[j].Path
	})
	return out, nil
}
