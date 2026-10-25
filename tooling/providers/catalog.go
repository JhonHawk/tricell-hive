package providers

// Catalog is static metadata for the onboarding summary. Engram, Context7,
// pi-subagents and codegraph are manual-only: Hive only shows where to look and
// why. agent-browser and codex-question-card have executable recipes (see
// ValidateStep). codex-wait-floor is the one entry where Hive edits a host
// configuration file itself: it writes only the keys it declares, keeps a
// backup and undoes them on removal, through management's plan, not a step.
// New entries go at the end, so existing indexes stay valid.
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
		{
			ID:     AgentBrowser,
			Hosts:  []string{"claude", "codex", "cursor", "grok", "opencode", "pi"},
			Source: "github.com/vercel-labs/agent-browser",
			Reason: "Install it with Homebrew or npm and download a browser with agent-browser install, following the official instructions.",
		},
		{
			ID:     CodexQuestionCard,
			Hosts:  []string{"codex"},
			Source: "github.com/openai/codex",
			Reason: "Run codex features enable default_mode_request_user_input so Codex shows its question card in Default mode.",
		},
		{
			ID:     CodeGraph,
			Hosts:  []string{"claude", "codex", "cursor", "opencode"},
			Source: "github.com/colbymchenry/codegraph",
			Reason: "Install it and run its own codegraph install for your hosts, following the official instructions; Hive does not run it or edit host configuration.",
		},
		{
			ID:     CodexWaitFloor,
			Hosts:  []string{"codex"},
			Source: "github.com/openai/codex",
			Reason: "Hive writes min_wait_timeout_ms and default_wait_timeout_ms under [features.multi_agent_v2] in Codex's config.toml, with a backup, so a parent agent does not wake every 60 seconds; hive plan remove undoes it.",
		},
	}
}
