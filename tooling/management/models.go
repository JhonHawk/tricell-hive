package management

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
	"tricell-hive/integrations/agents"
)

// ModelRow is the model and effort one installed agent role gets on one CLI.
// Model is empty when the CLI's own default applies. Effort is empty when the
// CLI has no per-agent level; on OpenCode the level lives in a "#variant"
// suffix of Model instead.
type ModelRow struct {
	Host, Role, Profile, Model, Effort string
	// Group is the folder of the role's source under content/agents/.
	Group string
	// Override is set when a stored override changed Model or Effort.
	Override bool
}

// EffectiveModels lists, read-only, the model and effort of every agent role
// installed for the user scope of o.Home. Each row is resolved from the
// release snapshot of its own record, because a record can stay on an older
// release than the rest of the installation. o.Hosts, when set, limits the
// rows to those CLIs; an empty list means every registered CLI. Rows are
// sorted by CLI, then role. A home with no agents gives an empty list. A role
// the state holds an override for is resolved with it and marked Override.
func EffectiveModels(o Options) ([]ModelRow, error) {
	o.Scope = "user"
	c, dir, err := normalize(o)
	if err != nil {
		return nil, err
	}
	state, _, err := readState(dir)
	if err != nil {
		return nil, err
	}
	wanted := map[string]bool{}
	for _, h := range o.Hosts {
		wanted[h] = true
	}
	// A role can have several records for one CLI, for example a shadowed
	// file left by an older layout beside the current one. Pick one per CLI
	// and role by a fixed rule, not by map order: the record at the CLI's
	// current destination first, then the newest release, then the path.
	current := map[string]map[string]bool{}
	currentPaths := func(host string) map[string]bool {
		paths, ok := current[host]
		if !ok {
			paths = map[string]bool{}
			if ts, err := resolve(c, []string{host}, stateSources(state)...); err == nil {
				for _, t := range ts {
					paths[t.Path] = true
				}
			}
			current[host] = paths
		}
		return paths
	}
	chosen := map[string]modelCandidate{}
	for _, key := range sortedRecordKeys(state.Records) {
		record := state.Records[key]
		if record.Target.Kind != "agent" || !agents.IsSource(record.Target.Source) {
			continue
		}
		var hosts []string
		for _, consumer := range record.Consumers {
			if consumer.Scope == "user" && consumer.Context == c.Home && (len(wanted) == 0 || wanted[consumer.Host]) {
				hosts = append(hosts, consumer.Host)
			}
		}
		if len(hosts) == 0 {
			continue
		}
		if !hexDigest64.MatchString(record.Release) {
			return nil, fmt.Errorf("invalid release ID for %s", record.Target.Path)
		}
		var written time.Time
		if info, err := os.Stat(filepath.Join(dir, "releases", record.Release+".json")); err == nil {
			written = info.ModTime()
		}
		role := strings.TrimSuffix(path.Base(record.Target.Source), ".md")
		for _, host := range hosts {
			cand := modelCandidate{record: record, host: host, role: role, written: written, current: currentPaths(host)[record.Target.Path]}
			if best, ok := chosen[host+"/"+role]; !ok || cand.beats(best) {
				chosen[host+"/"+role] = cand
			}
		}
	}
	keys := make([]string, 0, len(chosen))
	for key := range chosen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	releases := map[string]Release{}
	rows := []ModelRow{}
	for _, key := range keys {
		cand := chosen[key]
		record := cand.record
		release, ok := releases[record.Release]
		if !ok {
			file := filepath.Join(dir, "releases", record.Release+".json")
			if err := decodeFile(file, &release); err != nil {
				return nil, fmt.Errorf("%s: %w", file, err)
			}
			releases[record.Release] = release
		}
		data, found := payloadData(release, record.Target.Source)
		if !found {
			return nil, fmt.Errorf("release %s has no payload for %s", record.Release, record.Target.Source)
		}
		// An override of a role the release no longer has never reaches this
		// point: rows come from installed roles, so it stays in the state unlisted.
		var override *ModelOverride
		if v, ok := state.ModelOverrides[cand.host][cand.role]; ok {
			override = &v
		}
		profile, m, err := agents.Resolve(record.Target.Source, data, release.Profiles, cand.host, override)
		if err != nil {
			return nil, fmt.Errorf("%s on %s: %w", cand.role, cand.host, err)
		}
		rows = append(rows, ModelRow{Host: cand.host, Role: cand.role, Group: path.Base(path.Dir(record.Target.Source)), Profile: profile, Model: m.Model, Effort: m.Effort, Override: override != nil})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Host != rows[j].Host {
			return rows[i].Host < rows[j].Host
		}
		return rows[i].Role < rows[j].Role
	})
	return rows, nil
}

// modelCandidate is one record that could supply the row of a CLI and role.
type modelCandidate struct {
	record  Record
	host    string
	role    string
	written time.Time // write time of the record's release snapshot; zero when unreadable
	current bool      // the record's path is where the CLI's layout now installs the role
}

// beats reports whether c should replace other for the same CLI and role.
func (c modelCandidate) beats(other modelCandidate) bool {
	switch {
	case c.current != other.current:
		return c.current
	case !c.written.Equal(other.written):
		return c.written.After(other.written)
	}
	return c.record.Target.Path < other.record.Target.Path
}

func sortedRecordKeys(records map[string]Record) []string {
	keys := make([]string, 0, len(records))
	for key := range records {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func payloadData(r Release, source string) ([]byte, bool) {
	for _, f := range r.Files {
		if f.Path == source {
			return f.Data, true
		}
	}
	return nil, false
}

// BuildModelsPlan returns the plan that makes next the complete set of
// overrides for one CLI's roles. It reinstalls that CLI from the release it
// already has, so only agent files whose rendering changes are written; it never
// restores, migrates or updates anything else. An empty next removes every
// override of the CLI. The plan keeps the CLI's receipt when the reinstalled
// release is the one the receipt names, so the CLI stays verified.
func BuildModelsPlan(o Options, host string, next map[string]ModelOverride) (Plan, error) {
	o.Scope = "user"
	if _, err := validateHosts([]string{host}); err != nil {
		return Plan{}, err
	}
	c, dir, err := normalize(o)
	if err != nil {
		return Plan{}, err
	}
	state, _, err := readState(dir)
	if err != nil {
		return Plan{}, err
	}
	consumer := Consumer{host, "user", c.Home}
	registered, agentRecords := false, 0
	releases := map[string]bool{}
	for _, r := range state.Records {
		if !hasConsumer(r.Consumers, consumer) {
			continue
		}
		registered = true
		if r.Target.Kind == "agent" {
			agentRecords++
			releases[r.Release] = true
		}
	}
	if !registered {
		return Plan{}, fmt.Errorf("%s is not installed; run hive install first", host)
	}
	if agentRecords == 0 {
		return Plan{}, fmt.Errorf("%s has no agent roles installed; run hive install first", host)
	}
	if len(releases) != 1 {
		return Plan{}, fmt.Errorf("Run hive update first; %s has agents from more than one release", host)
	}
	var release string
	for id := range releases {
		release = id
	}
	o.Hosts, o.ReleaseID = []string{host}, release
	p, err := buildPlan("install", o, func(p *Plan, state State) error {
		roles := releaseRoles(p.Release)
		set := map[string]ModelOverride{}
		for role, v := range next {
			if err := agents.ValidateOverride(host, v); err != nil {
				return fmt.Errorf("model override for %s %s: %w", host, role, err)
			}
			// A stored override of a role the release lost is kept as it is.
			if stored, ok := state.ModelOverrides[host][role]; !roles[role] && (!ok || stored != v) {
				return fmt.Errorf("role %q is not in the release installed for %s", role, host)
			}
			set[role] = v
		}
		if len(set) == 0 {
			delete(p.ModelOverrides, host)
		} else {
			if p.ModelOverrides == nil {
				p.ModelOverrides = map[string]map[string]ModelOverride{}
			}
			p.ModelOverrides[host] = set
		}
		p.ModelOverrides = pruneOverrides(p.ModelOverrides)
		// Reinstalling the release the receipt names keeps the receipt, which
		// updateProductState rebuilds from the new hashes. Any other release
		// carries no product identity, as a rollback does.
		if receipt, ok := state.Installations[consumerKey(consumer)]; ok && receipt.Product.ReleaseID == p.Release.ID {
			product := receipt.Product
			if err := validateProduct(&product, state, p.Release.ID); err != nil {
				return err
			}
			p.Product = &product
		}
		return nil
	})
	if err != nil {
		return Plan{}, err
	}
	if err := modelsPlanRefusal(p); err != nil {
		return Plan{}, err
	}
	p.ID = planID(p)
	return p, nil
}

// modelsPlanRefusal rejects a plan that would do anything besides rewrite agent
// files, so that changing a model never restores, migrates or updates a
// resource the user did not ask about.
func modelsPlanRefusal(p Plan) error {
	const review = "; run hive install or hive doctor to review it"
	if len(p.Legacy) > 0 || p.Migration != nil {
		return fmt.Errorf("a legacy migration is pending for %v%s", p.Hosts, review)
	}
	if len(p.Voice) > 0 {
		return fmt.Errorf("the voice block of %v would change%s", p.Hosts, review)
	}
	for _, ch := range p.Changes {
		switch {
		case ch.Gone:
			return fmt.Errorf("%s was deleted by hand%s", ch.Target.Path, review)
		case ch.Target.Kind == "agent" && (ch.Before == nil || ch.After == nil):
			return fmt.Errorf("%s is not installed%s", ch.Target.Path, review)
		case ch.Target.Kind != "agent" && !reflect.DeepEqual(ch.Before, ch.After):
			return fmt.Errorf("%s is not an agent file and would change%s", ch.Target.Path, review)
		}
	}
	return nil
}

// releaseRoles returns the role names of the agents a release carries.
func releaseRoles(r *Release) map[string]bool {
	roles := map[string]bool{}
	for _, f := range r.Files {
		if agents.IsSource(f.Path) {
			roles[strings.TrimSuffix(path.Base(f.Path), ".md")] = true
		}
	}
	return roles
}

// copyOverrides returns a deep copy without empty sets, or nil when none remain.
func copyOverrides(in map[string]map[string]ModelOverride) map[string]map[string]ModelOverride {
	var out map[string]map[string]ModelOverride
	for host, roles := range in {
		if len(roles) == 0 {
			continue
		}
		if out == nil {
			out = map[string]map[string]ModelOverride{}
		}
		out[host] = map[string]ModelOverride{}
		for role, v := range roles {
			out[host][role] = v
		}
	}
	return out
}

// pruneOverrides drops empty sets and returns nil when none remain, so a state
// and a plan without overrides encode as they did before overrides existed.
func pruneOverrides(in map[string]map[string]ModelOverride) map[string]map[string]ModelOverride {
	return copyOverrides(in)
}

func sameOverrides(a, b map[string]map[string]ModelOverride) bool {
	return reflect.DeepEqual(pruneOverrides(a), pruneOverrides(b))
}

// nextModelOverrides returns the overrides the state holds after p. Only a
// user-scope install replaces its CLIs' overrides and only a user-scope remove
// drops them; every other plan leaves them as they are.
func nextModelOverrides(state State, p Plan) map[string]map[string]ModelOverride {
	next := copyOverrides(state.ModelOverrides)
	if p.Config.Scope == "user" && (p.Action == "install" || p.Action == "remove") {
		for _, host := range p.Hosts {
			delete(next, host)
			if p.Action == "install" && len(p.ModelOverrides[host]) > 0 {
				if next == nil {
					next = map[string]map[string]ModelOverride{}
				}
				next[host] = copyOverrides(map[string]map[string]ModelOverride{host: p.ModelOverrides[host]})[host]
			}
		}
	}
	return pruneOverrides(next)
}

// validateModelOverrides checks the overrides a plan carries, for every action.
func validateModelOverrides(p Plan, state State) error {
	if p.Action != "install" || p.Config.Scope != "user" {
		if len(p.ModelOverrides) != 0 {
			return fmt.Errorf("a %s plan at %s scope must not carry model overrides", p.Action, p.Config.Scope)
		}
		return nil
	}
	inPlan := map[string]bool{}
	for _, h := range p.Hosts {
		inPlan[h] = true
	}
	hosts := map[string]bool{}
	for h := range p.ModelOverrides {
		hosts[h] = true
	}
	for h := range state.ModelOverrides {
		hosts[h] = true
	}
	names := make([]string, 0, len(hosts))
	for h := range hosts {
		names = append(names, h)
	}
	sort.Strings(names)
	roles := releaseRoles(p.Release)
	for _, host := range names {
		if _, err := validateHosts([]string{host}); err != nil {
			return fmt.Errorf("invalid model overrides: %w", err)
		}
		set, present := p.ModelOverrides[host]
		if present && len(set) == 0 {
			return fmt.Errorf("empty model override set for %s", host)
		}
		if !inPlan[host] {
			if !reflect.DeepEqual(set, state.ModelOverrides[host]) && !(len(set) == 0 && len(state.ModelOverrides[host]) == 0) {
				return fmt.Errorf("model overrides for %s differ from the state, which this plan does not change", host)
			}
			continue
		}
		for role, v := range set {
			if err := agents.ValidateOverride(host, v); err != nil {
				return fmt.Errorf("model override for %s %s: %w", host, role, err)
			}
			if stored, ok := state.ModelOverrides[host][role]; ok && stored == v {
				continue
			}
			if !roles[role] {
				return fmt.Errorf("role %q is not in the release installed for %s", role, host)
			}
		}
	}
	return nil
}

// StoredModelOverrides returns, read-only, the overrides the state holds for the
// user scope of o.Home, keyed CLI then role. It is empty, never nil, when none
// are stored, and a copy the caller may change.
func StoredModelOverrides(o Options) (map[string]map[string]ModelOverride, error) {
	o.Scope = "user"
	_, dir, err := normalize(o)
	if err != nil {
		return nil, err
	}
	state, _, err := readState(dir)
	if err != nil {
		return nil, err
	}
	out := copyOverrides(state.ModelOverrides)
	if out == nil {
		out = map[string]map[string]ModelOverride{}
	}
	return out, nil
}
