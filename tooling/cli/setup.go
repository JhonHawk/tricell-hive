package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// Setup recommends optional tools. It never executes their installers, reads
// credentials, or treats a discovered skill file as runtime/auth verification.
func setup(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	home := fs.String("home", "", "synthetic home for local discovery; ignores host path overrides")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("setup accepts no positional arguments")
	}
	synthetic := *home != ""
	if !synthetic {
		var err error
		*home, err = os.UserHomeDir()
		if err != nil {
			return err
		}
	}
	resolved, err := filepath.Abs(*home)
	if err != nil {
		return err
	}
	env := os.Getenv
	if synthetic {
		env = func(string) string { return "" }
	}
	paths := context7Candidates(resolved, env)
	fmt.Fprintln(out, "Context7 — recommended, optional")
	found := 0
	for _, p := range paths {
		info, err := os.Stat(p) // Follow native skill aliases without modifying them.
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			fmt.Fprintf(out, "Not verified: %s (%v)\n", p, err)
			continue
		}
		if !info.Mode().IsRegular() || info.Size() == 0 {
			fmt.Fprintf(out, "Not verified: %s (not a nonempty regular file)\n", p)
			continue
		}
		found++
		fmt.Fprintf(out, "Skill file detected: %s\n", p)
	}
	if found == 0 {
		fmt.Fprintln(out, "No find-docs or context7-mcp skill file detected in the checked locations.")
	}
	fmt.Fprintln(out, "Local discovery does not verify host loading, credentials, service access, or freshness; MCP-only/custom installations may exist elsewhere.")
	fmt.Fprintln(out, "Many Hive development workflows benefit from current library documentation. Context7 is optional; Hive works without a key or this integration.")
	fmt.Fprintln(out, "To install or refresh CLI + Skills, run the official interactive setup when desired:")
	fmt.Fprintln(out, "  npx ctx7@latest setup --cli")
	fmt.Fprintln(out, "This requires Node.js/npm and network access. Choose the intended agents and complete any authentication yourself. Existing MCP users can keep that mode; avoid adding a second integration unintentionally.")
	fmt.Fprintln(out, "@latest selects the current CLI when invoked; setup refreshes the vendor skill and rules. Nothing is updated in the background. Review vendor changes before refreshing.")
	fmt.Fprintln(out, "Documentation queries (ctx7 library/docs) can run without login at lower limits; the hosted setup wizard requires authentication.")
	fmt.Fprintln(out, "Without Context7 access, consult current official documentation directly and state any verification limits.")
	fmt.Fprintln(out, "Continue with hive plan install --hosts <selected-hosts> --scope user --out <plan-file>, then hive apply --plan <plan-file>.")
	return nil
}

func context7Candidates(home string, env func(string) string) []string {
	pick := func(key, fallback string) string {
		if v := env(key); v != "" {
			return v
		}
		return fallback
	}
	roots := []string{
		filepath.Join(home, ".agents"), filepath.Join(home, ".claude"),
		pick("CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude")),
		pick("CODEX_HOME", filepath.Join(home, ".codex")),
		pick("GROK_HOME", filepath.Join(home, ".grok")),
		pick("PI_CODING_AGENT_DIR", filepath.Join(home, ".pi", "agent")),
		filepath.Join(pick("XDG_CONFIG_HOME", filepath.Join(home, ".config")), "opencode"),
	}
	seen := map[string]bool{}
	var paths []string
	for _, root := range roots {
		for _, name := range []string{"find-docs", "context7-mcp"} {
			p := filepath.Join(root, "skills", name, "SKILL.md")
			if !seen[p] {
				seen[p] = true
				paths = append(paths, p)
			}
		}
	}
	sort.Strings(paths)
	return paths
}
