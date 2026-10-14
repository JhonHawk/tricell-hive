package main

import (
	"flag"
	"io"
	"path/filepath"
)

// runGrokHook serves the hidden `hive grok-hook` subcommands that Grok runs as
// PreToolUse hooks. It never fails: Grok treats exit code 2 as deny, and a Go
// panic exits with 2, so every error and panic ends in empty stdout and exit
// 0, which lets the original tool call go ahead.
func runGrokHook(args []string, stdin io.Reader, stdout io.Writer, readWhole func(event []byte, skillsDir string) []byte) {
	defer func() { _ = recover() }()
	if len(args) == 0 || args[0] != "read-whole" {
		return
	}
	fs := flag.NewFlagSet("grok-hook read-whole", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var skills string
	fs.StringVar(&skills, "skills", "", "deployed skills directory")
	if fs.Parse(args[1:]) != nil || fs.NArg() != 0 || !filepath.IsAbs(skills) {
		return
	}
	event, err := io.ReadAll(io.LimitReader(stdin, 1<<20))
	if err != nil {
		return
	}
	// Compute the reply fully before writing so a panic leaves stdout empty.
	if out := readWhole(event, skills); len(out) > 0 {
		_, _ = stdout.Write(out)
	}
}
