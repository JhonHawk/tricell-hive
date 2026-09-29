package management

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
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
	releases := map[string]Release{}
	seen := map[string]bool{}
	rows := []ModelRow{}
	for _, record := range state.Records {
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
		release, ok := releases[record.Release]
		if !ok {
			if !hexDigest64.MatchString(record.Release) {
				return nil, fmt.Errorf("invalid release ID for %s", record.Target.Path)
			}
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
		role := strings.TrimSuffix(path.Base(record.Target.Source), ".md")
		for _, host := range hosts {
			if seen[host+"/"+role] {
				continue
			}
			seen[host+"/"+role] = true
			profile, m, err := agents.Resolve(record.Target.Source, data, release.Profiles, host)
			if err != nil {
				return nil, fmt.Errorf("%s on %s: %w", role, host, err)
			}
			rows = append(rows, ModelRow{Host: host, Role: role, Profile: profile, Model: m.Model, Effort: m.Effort})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Host != rows[j].Host {
			return rows[i].Host < rows[j].Host
		}
		return rows[i].Role < rows[j].Role
	})
	return rows, nil
}

func payloadData(r Release, source string) ([]byte, bool) {
	for _, f := range r.Files {
		if f.Path == source {
			return f.Data, true
		}
	}
	return nil, false
}
