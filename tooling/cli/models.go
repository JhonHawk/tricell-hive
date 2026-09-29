// models.go formats the effective model of every installed agent role. The
// Models view and the `hive models` command both render through it, so they
// show the same words in the same columns.
package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

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
	return modelCells{role: r.Role, profile: r.Profile, model: model, effort: effort}
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
			continue
		}
		header, lines := modelTable(byHost[host], 0)
		fmt.Fprintln(w, "  "+header)
		for _, l := range lines {
			fmt.Fprintln(w, "  "+l)
		}
	}
}
