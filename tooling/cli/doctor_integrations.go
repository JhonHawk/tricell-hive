package main

import "tricell-hive/tooling/management"

// collectIntegrations builds the Integrations section. It is a placeholder
// until task T2 fills it: an empty section has no title and is not rendered.
// deps arrives already adapted to the options (a synthetic home detects and
// executes nothing), and o carries Home, StateDir and Scope.
func collectIntegrations(o management.Options, deps doctorDeps) doctorSection {
	return doctorSection{}
}
