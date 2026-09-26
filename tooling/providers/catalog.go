package providers

// Catalog is static metadata for the onboarding summary. Every capability
// listed here is manual-only in this build: no host/platform combination has
// passed a native validation gate yet. Hive never installs or configures
// them; it only shows where to look and why.
func Catalog() []Provider {
	return []Provider{
		{
			ID:     Engram,
			Hosts:  []string{"codex", "claude", "opencode"},
			Source: "github.com/Gentleman-Programming/engram",
			Reason: "Native install/host-integration validation is pending; follow the official release and setup instructions.",
		},
		{
			ID:     Context7,
			Hosts:  []string{"codex", "claude", "cursor", "opencode"},
			Source: "context7.com",
			Reason: "Native install/host-integration validation is pending; follow the official CLI + Skills setup instructions.",
		},
		{
			ID:     PiSubagents,
			Hosts:  []string{"pi"},
			Source: "github.com/nicobailon/pi-subagents",
			Reason: "Native install validation is pending; use Pi's own package manager and its official instructions.",
		},
	}
}
