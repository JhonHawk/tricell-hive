package management

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"tricell-hive/integrations/target"
)

func consumer(t target.Target) Consumer      { return Consumer{t.Host, t.Scope, t.Context} }
func physical(t target.Target) target.Target { t.Host = ""; t.Scope = ""; t.Context = ""; return t }
func consumerKey(c Consumer) string          { return c.Host + "\x00" + c.Scope + "\x00" + c.Context }
func sortedConsumers(cs []Consumer) []Consumer {
	set := map[Consumer]bool{}
	for _, c := range cs {
		set[c] = true
	}
	out := make([]Consumer, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return consumerKey(out[i]) < consumerKey(out[j]) })
	return out
}
func hasConsumer(cs []Consumer, c Consumer) bool {
	for _, x := range cs {
		if x == c {
			return true
		}
	}
	return false
}
func subtractConsumers(cs, remove []Consumer) []Consumer {
	var out []Consumer
	for _, c := range cs {
		if !hasConsumer(remove, c) {
			out = append(out, c)
		}
	}
	return sortedConsumers(out)
}
func normalizeState(s *State) error {
	for path, r := range s.Records {
		if path != r.Target.Path || !filepath.IsAbs(path) || filepath.Clean(path) != path {
			return fmt.Errorf("invalid recorded path")
		}
		if r.Target.Kind != "block" && r.Target.Kind != "skill" && r.Target.Kind != "symlink" {
			return fmt.Errorf("invalid recorded kind")
		}
		if s.Version == 1 {
			if r.Target.Host == "" || len(r.Consumers) != 0 || r.Target.Kind == "symlink" {
				return fmt.Errorf("invalid legacy ownership")
			}
			r.Consumers = []Consumer{consumer(r.Target)}
			r.Target = physical(r.Target)
		}
		if r.Target != physical(r.Target) || len(r.Consumers) == 0 || !reflect.DeepEqual(r.Consumers, sortedConsumers(r.Consumers)) {
			return fmt.Errorf("invalid consumer ownership")
		}
		for _, c := range r.Consumers {
			if _, err := validateHosts([]string{c.Host}); err != nil {
				return err
			}
			if (c.Scope != "user" && c.Scope != "project") || !filepath.IsAbs(c.Context) || filepath.Clean(c.Context) != c.Context {
				return fmt.Errorf("invalid consumer context")
			}
		}
		s.Records[path] = r
	}
	return nil
}

type resource struct {
	Target        target.Target
	Consumers     []Consumer
	Replaces      *Record
	LegacyRemoval bool
}

func groups(bindings []target.Target) ([]resource, error) {
	var out []resource
	indices := map[string]int{}
	for _, t := range bindings {
		if i, ok := indices[t.Path]; ok {
			if out[i].Target != physical(t) {
				return nil, fmt.Errorf("incompatible shared destination: %s", t.Path)
			}
			out[i].Consumers = sortedConsumers(append(out[i].Consumers, consumer(t)))
		} else {
			indices[t.Path] = len(out)
			out = append(out, resource{Target: physical(t), Consumers: []Consumer{consumer(t)}})
		}
	}
	return out, nil
}

// Only the old user-scoped Claude skill directory has a migration route.
// No other relocated resource or preexisting directory is adopted.
func desiredResources(c target.Config, hosts []string, state State, action string) ([]resource, error) {
	bindings, err := resolve(c, hosts)
	if err != nil {
		return nil, err
	}
	gs, err := groups(bindings)
	if err != nil {
		return nil, err
	}
	for i := range gs {
		g := &gs[i]
		if g.Target.Kind != "symlink" || c.Scope != "user" {
			continue
		}
		oldPath := filepath.Join(g.Target.Path, "SKILL.md")
		old, ok := state.Records[oldPath]
		if !ok {
			continue
		}
		want := Consumer{"claude", "user", c.Home}
		if old.Target.Kind != "skill" || len(old.Consumers) != 1 || old.Consumers[0] != want {
			return nil, fmt.Errorf("legacy Claude ownership conflict")
		}
		if action == "remove" {
			g.Target = old.Target
			g.LegacyRemoval = true
			continue
		}
		ownedDir := false
		for _, p := range state.CreatedDirs {
			if p == g.Target.Path {
				ownedDir = true
			}
		}
		if !ownedDir {
			return nil, fmt.Errorf("legacy Claude directory is not owned: %s", g.Target.Path)
		}
		g.Replaces = &old
	}
	for _, r := range state.Records {
		for _, oldConsumer := range r.Consumers {
			for _, g := range gs {
				if !hasConsumer(g.Consumers, oldConsumer) || r.Target.Kind != g.Target.Kind || r.Target.Path == g.Target.Path {
					continue
				}
				migrating := false
				for _, alias := range gs {
					if alias.Replaces != nil && alias.Replaces.Target.Path == r.Target.Path {
						migrating = true
					}
				}
				// Removal maps the legacy skill itself; it must not also compare it
				// against the current shared skill destination.
				for _, legacy := range gs {
					if legacy.LegacyRemoval && legacy.Target.Path == r.Target.Path {
						migrating = true
					}
				}
				if !migrating {
					return nil, fmt.Errorf("recorded target is shadowed or relocated: %s", r.Target.Path)
				}
			}
		}
	}
	// Dependencies precede aliases on install, and follow aliases on removal.
	sort.SliceStable(gs, func(i, j int) bool {
		a, b := gs[i].Target.Kind == "symlink", gs[j].Target.Kind == "symlink"
		if a != b {
			if action == "remove" {
				return a
			}
			return !a
		}
		return false
	})
	return gs, nil
}
