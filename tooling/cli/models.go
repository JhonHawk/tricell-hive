// models.go formats the effective model of every installed agent role. The
// Models view and the `hive models` command both render through it, so they
// show the same words in the same columns.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"tricell-hive/integrations/agents"
	"tricell-hive/tooling/management"
)

// Column widths of the model table. Role is as wide as the longest role and is
// never cut; Profile and Effort are fixed; Model takes the rest of the row.
const (
	modelProfileWidth = 10
	modelEffortWidth  = 8
	modelGap          = 2
	modelMinWidth     = 28 // the least a Model column is given when it can be
	modelFloorWidth   = 12 // the least it is given when a role name leaves no room

	modelHostDefault   = "host default"
	modelInheritLabel  = "inherit (parent session)"
	modelEffortMissing = "-"

	modelMixed          = "mixed"
	modelOverrideMarker = " *"
	modelOverrideNote   = "* set with hive models set"
	modelNotApplied     = "not applied: role not in the installed release"
)

// modelCells are the four texts of one table row.
type modelCells struct{ role, profile, model, effort string }

// modelCellsFor turns a row from management.EffectiveModels into what a
// person reads. An empty model means the CLI picks its own default; the
// literal "inherit" means the agent runs on its parent session's model; and
// OpenCode carries its level as a "#variant" suffix of the model, which is
// shown as the effort instead.
func modelCellsFor(r management.ModelRow) modelCells {
	model, effort := r.Model, r.Effort
	if r.Host == "opencode" {
		if base, variant, found := strings.Cut(model, "#"); found && variant != "" {
			model, effort = base, variant
		}
	}
	switch model {
	case "":
		model = modelHostDefault
	case "inherit":
		model = modelInheritLabel
	}
	if effort == "" {
		effort = modelEffortMissing
	}
	role := r.Role
	if r.Override {
		role += modelOverrideMarker
	}
	return modelCells{role: role, profile: r.Profile, model: model, effort: effort}
}

// modelColumns are the widths of the four columns.
type modelColumns struct{ role, profile, model, effort int }

// layoutModelColumns sizes the columns for rows drawn in width columns. With
// width 0 nothing is cut and every column is as wide as its longest text.
func layoutModelColumns(cells []modelCells, width int) modelColumns {
	c := modelColumns{role: utf8.RuneCountInString("Role"), profile: modelProfileWidth, effort: modelEffortWidth}
	longestModel := utf8.RuneCountInString("Model")
	for _, cell := range cells {
		c.role = max(c.role, utf8.RuneCountInString(cell.role))
		longestModel = max(longestModel, utf8.RuneCountInString(cell.model))
		if width <= 0 {
			c.profile = max(c.profile, utf8.RuneCountInString(cell.profile))
			c.effort = max(c.effort, utf8.RuneCountInString(cell.effort))
		}
	}
	if width <= 0 {
		c.model = longestModel
		return c
	}
	available := width - c.role - c.profile - c.effort - 3*modelGap
	c.model = max(min(max(longestModel, modelMinWidth), available), modelFloorWidth)
	return c
}

// line draws one row. Cells longer than their column end in an ellipsis,
// except Role, which is never cut. Trailing spaces are dropped.
func (c modelColumns) line(cell modelCells) string {
	gap := strings.Repeat(" ", modelGap)
	parts := []string{
		padRight(cell.role, c.role),
		padRight(truncateRunes(cell.profile, c.profile), c.profile),
		padRight(truncateRunes(cell.model, c.model), c.model),
		truncateRunes(cell.effort, c.effort),
	}
	return strings.TrimRight(strings.Join(parts, gap), " ")
}

func (c modelColumns) header() string {
	return c.line(modelCells{role: "Role", profile: "Profile", model: "Model", effort: "Effort"})
}

// modelTable draws the header and one line per row, for rows of one CLI.
func modelTable(rows []management.ModelRow, width int) (header string, lines []string) {
	cells := make([]modelCells, len(rows))
	for i, r := range rows {
		cells[i] = modelCellsFor(r)
	}
	columns := layoutModelColumns(cells, width)
	lines = make([]string, len(cells))
	for i, cell := range cells {
		lines[i] = columns.line(cell)
	}
	return columns.header(), lines
}

// sortedByGroup returns rows ordered by group and then role. The sort is
// stable, so rows already ordered by role keep that order inside a group.
func sortedByGroup(rows []management.ModelRow) []management.ModelRow {
	out := append([]management.ModelRow(nil), rows...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Group != out[j].Group {
			return out[i].Group < out[j].Group
		}
		return out[i].Role < out[j].Role
	})
	return out
}

// groupSummary is what a group header shows: the model and the effort the
// group's roles have in common, or "mixed" for a part on which they differ, and
// whether any of its roles has an override of its own. It reads the same cells
// as the table, so the header and the rows below it never disagree.
func groupSummary(rows []management.ModelRow) (model, effort string, override bool) {
	for i, r := range rows {
		c := modelCellsFor(r)
		override = override || r.Override
		if i == 0 {
			model, effort = c.model, c.effort
			continue
		}
		if c.model != model {
			model = modelMixed
		}
		if c.effort != effort {
			effort = modelMixed
		}
	}
	return model, effort, override
}

// groupHeaderLine is the header line of the group that starts at rows[i], or
// false when the row has no group.
func groupHeaderLine(rows []management.ModelRow, i int) (string, bool) {
	group := rows[i].Group
	if group == "" {
		return "", false
	}
	end := i
	for end < len(rows) && rows[end].Group == group {
		end++
	}
	model, effort, override := groupSummary(rows[i:end])
	name := group
	if override {
		name += modelOverrideMarker
	}
	return name + "  " + model + "  " + effort, true
}

// collectModels reads the registered CLIs and the effective model of every
// installed role, for the Models view and the `hive models` command alike.
func collectModels(o management.Options) (hosts []string, rows []management.ModelRow, err error) {
	o.Hosts = nil // every registered CLI, not the ones a command selected
	if hosts, err = management.RegisteredHosts(o); err != nil {
		return nil, nil, err
	}
	rows, err = management.EffectiveModels(o)
	return hosts, rows, err
}

// renderModelsText writes one table per registered CLI, without cutting any
// text, and the same empty states as the Models view.
func renderModelsText(hosts []string, rows []management.ModelRow, w io.Writer) {
	renderModelsTextWith(hosts, rows, nil, w)
}

// renderModelsTextWith is renderModelsText plus, per CLI, the roles whose stored
// override the installed release cannot apply, and a note when a row is marked.
func renderModelsTextWith(hosts []string, rows []management.ModelRow, unapplied map[string][]string, w io.Writer) {
	if len(hosts) == 0 {
		fmt.Fprintln(w, noHostsText)
		return
	}
	byHost := map[string][]management.ModelRow{}
	for _, r := range rows {
		byHost[r.Host] = append(byHost[r.Host], r)
	}
	hosts = append([]string(nil), hosts...)
	sort.Strings(hosts)
	for i, host := range hosts {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, host)
		if len(byHost[host]) == 0 {
			fmt.Fprintln(w, "  No agents installed for "+host)
		} else {
			hostRows := sortedByGroup(byHost[host])
			header, lines := modelTable(hostRows, 0)
			fmt.Fprintln(w, "  "+header)
			for i, l := range lines {
				if i == 0 || hostRows[i].Group != hostRows[i-1].Group {
					if h, ok := groupHeaderLine(hostRows, i); ok {
						fmt.Fprintln(w, "  "+h)
					}
				}
				fmt.Fprintln(w, "  "+l)
			}
		}
		for _, role := range unapplied[host] {
			fmt.Fprintf(w, "  %s: %s\n", role, modelNotApplied)
		}
	}
	for _, r := range rows {
		if r.Override {
			fmt.Fprintln(w)
			fmt.Fprintln(w, modelOverrideNote)
			break
		}
	}
}

// runModels prints the effective model and effort of every installed role, or
// runs the set and reset subcommands that change them.
func runModels(args []string, w io.Writer) error {
	if len(args) > 0 && (args[0] == "set" || args[0] == "reset") {
		return modelsWrite(args[0], args[1:], os.Stdin, w, terminalInput(os.Stdin))
	}
	o, ok, err := readOnlyOptions(flag.NewFlagSet("models", flag.ContinueOnError), args)
	if !ok {
		return err
	}
	hosts, rows, err := collectModels(o)
	if err != nil {
		return err
	}
	stored, err := management.StoredModelOverrides(o)
	if err != nil {
		return err
	}
	applied := map[string]bool{}
	for _, r := range rows {
		applied[r.Host+"/"+r.Role] = true
	}
	unapplied := map[string][]string{}
	for host, roles := range stored {
		for role := range roles {
			if !applied[host+"/"+role] {
				unapplied[host] = append(unapplied[host], role)
			}
		}
		sort.Strings(unapplied[host])
	}
	renderModelsTextWith(hosts, rows, unapplied, w)
	return nil
}

// modelsChange is one role's change: the parts to set and the parts to drop.
// `hive models set` and the Models view both express their request with it, so
// the same request leaves the same override.
type modelsChange struct {
	set                            management.ModelOverride
	dropAll, dropModel, dropEffort bool
}

func (c modelsChange) empty() bool {
	return c.set == (management.ModelOverride{}) && !c.dropAll && !c.dropModel && !c.dropEffort
}

// applyTo returns the CLI's override set after the change to role. Parts the
// change does not name keep their stored value; the release's values are never
// copied in, so the rest keeps following the release.
func (c modelsChange) applyTo(stored map[string]management.ModelOverride, role string) map[string]management.ModelOverride {
	next := map[string]management.ModelOverride{}
	for r, v := range stored {
		next[r] = v
	}
	v := next[role]
	switch {
	case c.dropAll:
		v = management.ModelOverride{}
	default:
		if c.dropModel {
			v.Model = ""
		}
		if c.dropEffort {
			v.Effort = ""
		}
	}
	if c.set.Model != "" {
		v.Model = c.set.Model
	}
	if c.set.Effort != "" {
		v.Effort = c.set.Effort
	}
	if v == (management.ModelOverride{}) {
		delete(next, role)
	} else {
		next[role] = v
	}
	return next
}

// planModelsChange builds the plan that makes next the override set of host,
// with the effective rows as they are before it.
func planModelsChange(o management.Options, host string, next map[string]management.ModelOverride) (management.Plan, []management.ModelRow, error) {
	o.Hosts = []string{host}
	before, err := management.EffectiveModels(o)
	if err != nil {
		return management.Plan{}, nil, err
	}
	p, err := management.BuildModelsPlan(o, host, next)
	return p, before, err
}

// modelsFlags are the flags of `hive models set` and `hive models reset`.
type modelsFlags struct {
	host, role, group, model, effort, only, home, stateDir, out string
	all, dry                                                    bool
	modelGiven, effortGiven                                     bool
}

func parseModelsFlags(sub string, args []string) (f modelsFlags, err error) {
	fs := flag.NewFlagSet("models "+sub, flag.ContinueOnError)
	fs.StringVar(&f.host, "host", "", "CLI whose role is changed")
	fs.StringVar(&f.role, "role", "", "agent role")
	fs.StringVar(&f.group, "group", "", "agent group: every role under content/agents/<group>/")
	if sub == "set" {
		fs.StringVar(&f.model, "model", "", "model for the role")
		fs.StringVar(&f.effort, "effort", "", "effort for the role")
	} else {
		fs.BoolVar(&f.all, "all", false, "remove every override of the CLI")
		fs.StringVar(&f.only, "only", "", "remove only the model or the effort of the role's override")
	}
	fs.StringVar(&f.home, "home", "", "explicit synthetic home; ignores host environment paths")
	fs.StringVar(&f.stateDir, "state-dir", "", "state directory (default: user Application Support/tricell-hive)")
	fs.BoolVar(&f.dry, "dry-run", false, "preview only; do not change anything")
	fs.StringVar(&f.out, "out", "", "save the plan to FILE instead of applying")
	if err = fs.Parse(args); err != nil {
		return f, err
	}
	if fs.NArg() != 0 {
		return f, fmt.Errorf("unexpected positional arguments")
	}
	fs.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "model":
			f.modelGiven = true
		case "effort":
			f.effortGiven = true
		}
	})
	return f, nil
}

// modelsWrite runs `hive models set` or `hive models reset`: it builds the plan
// of management.BuildModelsPlan from the CLI's stored overrides and the change
// asked for, then previews, confirms and applies it like `hive voice set`.
func modelsWrite(sub string, args []string, in io.Reader, out io.Writer, interactive bool) error {
	f, err := parseModelsFlags(sub, args)
	if err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if f.host == "" {
		return fmt.Errorf("--host is required")
	}
	o := management.Options{Home: f.home, StateDir: f.stateDir}
	stored, err := management.StoredModelOverrides(o)
	if err != nil {
		return err
	}
	if f.group != "" && f.role != "" {
		return fmt.Errorf("--group and --role cannot be combined")
	}
	var next map[string]management.ModelOverride
	var groupRows []management.ModelRow
	var replaced []string
	if sub == "set" {
		if f.role == "" && f.group == "" {
			return fmt.Errorf("--role or --group is required")
		}
		if f.modelGiven && f.model == "" {
			return fmt.Errorf("--model must not be empty; use hive models reset --only model to return to the release's model")
		}
		if f.effortGiven && f.effort == "" {
			return fmt.Errorf("--effort must not be empty; use hive models reset --only effort to return to the release's effort")
		}
		if !f.modelGiven && !f.effortGiven {
			return fmt.Errorf("hive models set needs --model or --effort")
		}
		if f.group != "" {
			if groupRows, err = groupRowsOf(o, f.host, f.group, stored[f.host]); err != nil {
				return err
			}
			next, replaced = setGroup(stored[f.host], groupRows, management.ModelOverride{Model: f.model, Effort: f.effort})
		} else {
			set := management.ModelOverride{Model: f.model, Effort: f.effort}
			if set.Effort == "" {
				// Only the opencode rule can add an effort; it needs the role's row.
				rows, err := hostModelRows(o, f.host)
				if err != nil {
					return err
				}
				for _, r := range rows {
					if r.Role == f.role {
						set = keepShownEffort(f.host, r, set)
					}
				}
			}
			next = modelsChange{set: set}.applyTo(stored[f.host], f.role)
		}
	} else {
		switch {
		case f.all && f.group != "":
			return fmt.Errorf("--all cannot be combined with --group")
		case f.all && (f.role != "" || f.only != ""):
			if f.role != "" {
				return fmt.Errorf("--all cannot be combined with --role")
			}
			return fmt.Errorf("--only needs --role or --group")
		case !f.all && f.role == "" && f.group == "":
			return fmt.Errorf("hive models reset needs --role, --group or --all")
		case f.only != "" && f.only != "model" && f.only != "effort":
			return fmt.Errorf("--only must be model or effort")
		}
		change := modelsChange{dropAll: f.only == "", dropModel: f.only == "model", dropEffort: f.only == "effort"}
		switch {
		case f.all:
			next = map[string]management.ModelOverride{}
		case f.group != "":
			if groupRows, err = groupRowsOf(o, f.host, f.group, stored[f.host]); err != nil {
				return err
			}
			next = stored[f.host]
			for _, r := range groupRows {
				next = change.applyTo(next, r.Role)
			}
		default:
			next = change.applyTo(stored[f.host], f.role)
		}
	}
	p, before, err := planModelsChange(o, f.host, next)
	if err != nil {
		return err
	}
	if sub == "reset" && f.role != "" {
		_, known := stored[f.host][f.role]
		for _, r := range before {
			known = known || r.Role == f.role
		}
		if !known {
			return fmt.Errorf("role %q is unknown for %s", f.role, f.host)
		}
	}
	terminal := installTerminal{reader: bufio.NewReader(in), out: out, interactive: interactive}
	return confirmAndApplyPlan(p, planRun{
		summary: func(unchanged bool) {
			if !unchanged && len(replaced) > 0 {
				fmt.Fprintf(out, "Replaces the own override of: %s\n", strings.Join(replaced, ", "))
			}
			showModelsSummary(out, f.host, p, before, stored[f.host], unchanged)
		},
		doneVerb:  "Model settings updated",
		unchanged: "Nothing to change: the model settings already match.",
		stateOnly: !modelsFilesChange(p),
	}, out, interactive, f.dry, f.out, terminal)
}

// hostModelRows is the effective rows of one CLI as they are before a change.
func hostModelRows(o management.Options, host string) ([]management.ModelRow, error) {
	o.Hosts = []string{host}
	return management.EffectiveModels(o)
}

// groupRowsOf returns the rows of the roles installed for host in group. An
// unknown group is refused, naming the groups the CLI has. A CLI that cannot be
// changed at all (not installed, no agents) reports that instead, as it does for
// --role.
func groupRowsOf(o management.Options, host, group string, stored map[string]management.ModelOverride) ([]management.ModelRow, error) {
	rows, err := hostModelRows(o, host)
	if err != nil {
		return nil, err
	}
	var in []management.ModelRow
	known := map[string]bool{}
	for _, r := range rows {
		known[r.Group] = true
		if r.Group == group {
			in = append(in, r)
		}
	}
	if len(in) > 0 {
		return in, nil
	}
	if _, err := management.BuildModelsPlan(o, host, stored); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(known))
	for g := range known {
		names = append(names, g)
	}
	sort.Strings(names)
	return nil, fmt.Errorf("group %q is unknown for %s; known groups: %s", group, host, strings.Join(names, ", "))
}

// keepShownEffort applies the OpenCode rule: there the effort travels inside
// the model id as "#variant", so changing the model without choosing an effort
// would drop the variant the role shows today. The shown effort is written into
// the override, the one case where a release value is copied into it.
func keepShownEffort(host string, shown management.ModelRow, set management.ModelOverride) management.ModelOverride {
	if host != "opencode" || set.Model == "" || set.Effort != "" {
		return set
	}
	if e := modelCellsFor(shown).effort; e != modelEffortMissing {
		set.Effort = e
	}
	return set
}

// setGroup returns the CLI's override set once every role of the group has an
// override with only the given parts, replacing its own, and the roles whose own
// override that replaces with something different.
func setGroup(stored map[string]management.ModelOverride, rows []management.ModelRow, set management.ModelOverride) (map[string]management.ModelOverride, []string) {
	next := map[string]management.ModelOverride{}
	for r, v := range stored {
		next[r] = v
	}
	var replaced []string
	for _, r := range rows {
		v := keepShownEffort(r.Host, r, set)
		if old, ok := stored[r.Role]; ok && old != v {
			replaced = append(replaced, r.Role)
		}
		next[r.Role] = v
	}
	sort.Strings(replaced)
	return next, replaced
}

// showModelsSummary lists, per role whose override changes, the effective model
// and effort before and after, and the agent file that is rewritten.
func showModelsSummary(out io.Writer, host string, p management.Plan, before []management.ModelRow, stored map[string]management.ModelOverride, unchanged bool) {
	if unchanged {
		return
	}
	rowOf := map[string]management.ModelRow{}
	for _, r := range before {
		rowOf[r.Role] = r
	}
	fileOf := map[string]string{}
	for _, ch := range p.Changes {
		if ch.Target.Kind == "agent" && ch.Before != nil && ch.After != nil && string(ch.Before.Managed) != string(ch.After.Managed) {
			fileOf[strings.TrimSuffix(pathBase(ch.Target.Source), ".md")] = ch.Target.Path
		}
	}
	roleSet := map[string]bool{}
	for role, v := range p.ModelOverrides[host] {
		if old, ok := stored[role]; !ok || old != v {
			roleSet[role] = true
		}
	}
	for role, v := range stored {
		if new, ok := p.ModelOverrides[host][role]; !ok || new != v {
			roleSet[role] = true
		}
	}
	roles := make([]string, 0, len(roleSet))
	for role := range roleSet {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	for _, role := range roles {
		old, ok := rowOf[role]
		if !ok {
			fmt.Fprintf(out, "%s %s: not applied (role not in the installed release); only the stored setting changes\n", host, role)
			continue
		}
		after, err := resolveAfter(p, host, role, old)
		if err != nil {
			fmt.Fprintf(out, "%s %s: %v\n", host, role, err)
			continue
		}
		b, a := modelCellsFor(old), modelCellsFor(after)
		fmt.Fprintf(out, "%s %s: model %s → %s, effort %s → %s\n", host, role, b.model, a.model, b.effort, a.effort)
		if file, ok := fileOf[role]; ok {
			fmt.Fprintf(out, "  File: %s\n", file)
		} else {
			fmt.Fprintln(out, "  No file changes; the setting is saved in Hive's state.")
		}
	}
	fmt.Fprintln(out, "Open sessions keep the previous model until they restart.")
}

// resolveAfter is the model row a role gets once p is applied.
func resolveAfter(p management.Plan, host, role string, old management.ModelRow) (management.ModelRow, error) {
	for _, f := range p.Release.Files {
		if !agents.IsSource(f.Path) || strings.TrimSuffix(pathBase(f.Path), ".md") != role {
			continue
		}
		var override *management.ModelOverride
		if v, ok := p.ModelOverrides[host][role]; ok {
			override = &v
		}
		profile, m, err := agents.Resolve(f.Path, f.Data, p.Release.Profiles, host, override)
		if err != nil {
			return old, err
		}
		return management.ModelRow{Host: host, Role: role, Group: old.Group, Profile: profile, Model: m.Model, Effort: m.Effort, Override: override != nil}, nil
	}
	return old, fmt.Errorf("role is not in the release")
}

func pathBase(p string) string { return p[strings.LastIndex(p, "/")+1:] }

// modelsFilesChange reports whether p rewrites an agent file.
func modelsFilesChange(p management.Plan) bool {
	for _, ch := range p.Changes {
		if ch.Target.Kind == "agent" && ch.Before != nil && ch.After != nil && string(ch.Before.Managed) != string(ch.After.Managed) {
			return true
		}
	}
	return false
}
