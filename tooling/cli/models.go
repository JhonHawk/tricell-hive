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
	"path/filepath"
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
	title := modelsCommandTitle(sub, f, groupRows, next)
	terminal := installTerminal{reader: bufio.NewReader(in), out: out, interactive: interactive}
	return confirmAndApplyPlan(p, planRun{
		summary: func(unchanged bool) {
			if !unchanged {
				fmt.Fprintln(out, title)
			}
			showModelsSummary(out, f.host, p, before, stored[f.host], replaced, f.home, 80, 0, unchanged)
		},
		doneVerb:  "Model settings updated",
		unchanged: "Nothing to change: the model settings already match.",
		stateOnly: !modelsFilesChange(p),
	}, out, interactive, f.dry, f.out, terminal)
}

// modelsCommandTitle is the heading the command prints above its summary.
func modelsCommandTitle(sub string, f modelsFlags, groupRows []management.ModelRow, next map[string]management.ModelOverride) string {
	verb := "Change"
	if sub == "reset" {
		verb = "Reset"
	}
	switch {
	case f.all:
		return "Reset every override on " + f.host
	case f.group != "":
		return modelsTitle(verb, "", f.group, f.host, len(groupRows))
	}
	return modelsTitle(verb, f.role, "", f.host, 0)
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

// loses reports whether replacing the override old by next drops a part of old:
// a part old holds that next lacks or changes. Adding a part loses nothing.
func loses(old, next management.ModelOverride) bool {
	return (old.Model != "" && next.Model != old.Model) || (old.Effort != "" && next.Effort != old.Effort)
}

// setGroup returns the CLI's override set once every role of the group has an
// override with only the given parts, replacing its own, and the roles that lose
// a part of their own override by it.
func setGroup(stored map[string]management.ModelOverride, rows []management.ModelRow, set management.ModelOverride) (map[string]management.ModelOverride, []string) {
	next := map[string]management.ModelOverride{}
	for r, v := range stored {
		next[r] = v
	}
	var replaced []string
	for _, r := range rows {
		v := keepShownEffort(r.Host, r, set)
		if old, ok := stored[r.Role]; ok && loses(old, v) {
			replaced = append(replaced, r.Role)
		}
		if v == (management.ModelOverride{}) {
			delete(next, r.Role) // the view can ask for the release's values everywhere
			continue
		}
		next[r.Role] = v
	}
	sort.Strings(replaced)
	return next, replaced
}

// modelsTitle is the heading of a change's confirmation. A group names itself
// and its size; a role names itself.
func modelsTitle(verb, role, group, host string, roles int) string {
	if group == "" {
		return verb + " " + role + " on " + host
	}
	noun := "roles"
	if roles == 1 {
		noun = "role"
	}
	return fmt.Sprintf("%s the %s group on %s (%d %s)", verb, strings.ToUpper(group[:1])+group[1:], host, roles, noun)
}

// showModelsSummary writes the body of a change's confirmation, after its title:
// a table with one row per role whose effective model or effort changes, where
// a part that changes reads "before → after" and the others read plain; the
// roles that lose a part of their own override; and the agent files written,
// with their common directory shown once. width wraps the prose lines; 0 leaves
// them to the caller. tableWidth fits the table to a screen by cutting long ids
// from the left, so the model name stays visible; 0 keeps the full ids. home is the home in use, to abbreviate paths with "~".
func showModelsSummary(out io.Writer, host string, p management.Plan, before []management.ModelRow, stored map[string]management.ModelOverride, replaced []string, home string, width, tableWidth int, unchanged bool) {
	if unchanged {
		return
	}
	rowOf := map[string]management.ModelRow{}
	for _, r := range before {
		rowOf[r.Role] = r
	}
	var files []string
	fileOf := map[string]bool{}
	for _, ch := range p.Changes {
		if ch.Target.Kind == "agent" && ch.Before != nil && ch.After != nil && string(ch.Before.Managed) != string(ch.After.Managed) {
			fileOf[strings.TrimSuffix(pathBase(ch.Target.Source), ".md")] = true
			files = append(files, ch.Target.Path)
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
	type changeRow struct{ role, mb, ma, eb, ea string }
	var table []changeRow
	var notes []string
	for _, role := range roles {
		old, ok := rowOf[role]
		if !ok {
			notes = append(notes, role+": not applied (role not in the installed release); only the stored setting changes")
			continue
		}
		after, err := resolveAfter(p, host, role, old)
		if err != nil {
			notes = append(notes, role+": "+err.Error())
			continue
		}
		b, a := modelCellsFor(old), modelCellsFor(after)
		table = append(table, changeRow{role, b.model, a.model, b.effort, a.effort})
	}
	fmt.Fprintln(out)
	if len(table) > 0 {
		roleW := utf8.RuneCountInString("Role")
		effW := utf8.RuneCountInString("Effort")
		for _, r := range table {
			roleW = max(roleW, utf8.RuneCountInString(r.role))
			effW = max(effW, utf8.RuneCountInString(changeCell(r.eb, r.ea, 0)))
		}
		modelRoom := 0 // 0: no limit
		if tableWidth > 0 {
			roleW = min(roleW, max(tableWidth/3, 12))
			modelRoom = max(tableWidth-roleW-effW-4, 10)
		}
		cells := make([][3]string, 0, len(table)+1)
		cells = append(cells, [3]string{"Role", "Model", "Effort"})
		for _, r := range table {
			cells = append(cells, [3]string{truncateRunes(r.role, roleW), changeModelCell(r.mb, r.ma, modelRoom), changeCell(r.eb, r.ea, 0)})
		}
		modelW := 0
		for _, c := range cells {
			modelW = max(modelW, utf8.RuneCountInString(c[1]))
		}
		for _, c := range cells {
			fmt.Fprintln(out, strings.TrimRight(padRight(c[0], roleW)+"  "+padRight(c[1], modelW)+"  "+c[2], " "))
		}
		fmt.Fprintln(out)
	}
	for _, n := range notes {
		fmt.Fprintln(out, n)
	}
	if len(replaced) > 0 {
		list := strings.Join(replaced, ", ")
		if n := len(replaced); n > 1 {
			list = strings.Join(replaced[:n-1], ", ") + " and " + replaced[n-1]
		}
		text := "Replaces the own override of " + list + "."
		if width > 0 {
			text = strings.Join(wrapLines(text, width), "\n")
		}
		fmt.Fprintln(out, text)
	}
	if len(files) == 0 {
		fmt.Fprintln(out, "No agent file changes; the setting is saved in Hive's state.")
		return
	}
	noun := "files"
	if len(files) == 1 {
		noun = "file"
	}
	fmt.Fprintf(out, "Writes %d %s in %s\n", len(files), noun, abbreviateHome(commonDir(files), home))
	fmt.Fprintln(out, "Open sessions keep the previous model until they restart.")
}

// changeCell reads "before → after", or the plain value when it does not change.
func changeCell(before, after string, _ int) string {
	if before == after {
		return after
	}
	return before + " → " + after
}

// changeModelCell is changeCell for a model, cut to room columns when room is
// positive: an id too long for its half is cut from the left ("…" and the end
// kept, since the model name is what tells models apart), and the shorter id
// gives its spare columns to the longer one.
func changeModelCell(before, after string, room int) string {
	if before == after || room <= 0 {
		return changeCell(before, after, 0)
	}
	room -= len(" → ")
	nb, na := utf8.RuneCountInString(before), utf8.RuneCountInString(after)
	if nb+na <= room {
		return before + " → " + after
	}
	half := room / 2
	switch {
	case nb <= half:
		return before + " → " + cutLeft(after, room-nb)
	case na <= half:
		return cutLeft(before, room-na) + " → " + after
	}
	return cutLeft(before, half) + " → " + cutLeft(after, room-half)
}

// cutLeft keeps the last width columns of s, with "…" in front when it cut.
func cutLeft(s string, width int) string {
	r := []rune(s)
	if len(r) <= width || width < 2 {
		return s
	}
	return "…" + string(r[len(r)-width+1:])
}

// commonDir is the deepest directory that holds every file.
func commonDir(files []string) string {
	dir := strings.Split(filepath.Dir(files[0]), string(filepath.Separator))
	for _, f := range files[1:] {
		parts := strings.Split(filepath.Dir(f), string(filepath.Separator))
		n := 0
		for n < len(dir) && n < len(parts) && dir[n] == parts[n] {
			n++
		}
		dir = dir[:n]
	}
	return strings.Join(dir, string(filepath.Separator))
}

// abbreviateHome writes path relative to the home in use as "~/…". With no
// explicit home, the user's home is used. A path that is not under it stays whole.
func abbreviateHome(path, home string) string {
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	if home == "" {
		return path
	}
	bases := []string{filepath.Clean(home)}
	if real, err := filepath.EvalSymlinks(home); err == nil {
		bases = append(bases, real)
	}
	for _, b := range bases {
		if path == b {
			return "~"
		}
		if strings.HasPrefix(path, b+string(filepath.Separator)) {
			return "~" + path[len(b):]
		}
	}
	return path
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
