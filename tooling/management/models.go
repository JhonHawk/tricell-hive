package management

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
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
}

// EffectiveModels lists, read-only, the model and effort of every agent role
// installed for the user scope of o.Home. Each row is resolved from the
// release snapshot of its own record, because a record can stay on an older
// release than the rest of the installation. o.Hosts, when set, limits the
// rows to those CLIs; an empty list means every registered CLI. Rows are
// sorted by CLI, then role. A home with no agents gives an empty list.
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
		profile, m, err := agents.Resolve(record.Target.Source, data, release.Profiles, cand.host, nil)
		if err != nil {
			return nil, fmt.Errorf("%s on %s: %w", cand.role, cand.host, err)
		}
		rows = append(rows, ModelRow{Host: cand.host, Role: cand.role, Profile: profile, Model: m.Model, Effort: m.Effort})
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
