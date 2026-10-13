package main

import (
	"regexp"
	"strings"
)

// Protocol evidence is evaluated independently of the strict operation gate.
// A later inspection can decide operations, but cannot repair missing proof.
func flashProtocolTerminal(r result, evidenceHash string) string {
	terminal := terminalState(r.Trace, r.ExitCode, r.Error)
	f := r.FlashClass
	if f == nil || f.PrimaryCompletion == nil || !f.PrimaryCompletion.Completed || f.PrimaryCompletion.SessionID != r.Trace.SessionID || f.PrimaryCompletion.Model != r.ModelRequested || flashCompletion(r.Trace.SessionID, r.ModelRequested, f.Role, f.Children) != nil || f.Completion != "observed" || (f.Memory != "verified_clean" && !(r.MemoryStoreMode == disposableContainer && f.Memory == deferredVolumeDestruction)) || f.MemoryRouting != "observed" || r.ModelObserved != r.ModelRequested || evidenceHash == "absent" || f.EvidenceHash != evidenceHash || f.Error != "" || f.Coverage == "blocked" || !flashPrimaryHTTPVerified(r) || (r.Arm == "C" && f.Coverage != "observed") {
		return "protocol_unverified"
	}
	if r.MemoryStoreMode == disposableContainer && flashCredentialCleanupUnverified(r) {
		return "credential_cleanup_unverified"
	}
	if terminal == "not_verified" {
		return "completed"
	}
	return terminal
}

var flashSimpleWord = regexp.MustCompile(`^[A-Za-z0-9_./-]+$`)

// Persist only a single recognized read-only command with plain arguments.
// This grammar excludes quoting, assignments, expansions, globbing and shell
// composition. Everything outside it is withheld for rejection by inspection.
func flashOperationCommand(command string) string {
	if len(command) > 8192 || strings.ContainsAny(command, ";|&><`$(){}\n\r") {
		return ""
	}
	words := strings.Fields(command)
	if len(words) == 0 {
		return ""
	}
	for _, word := range words {
		if !flashSimpleWord.MatchString(word) {
			return ""
		}
	}
	name, args := words[0], words[1:]
	if name == "git" {
		if len(args) == 0 {
			return ""
		}
		name = "git " + args[0]
		args = args[1:]
	}
	if !flashReadOnlyArguments(name, args) {
		return ""
	}

	return command
}

func flashReadOnlyArguments(name string, args []string) bool {
	options := ""
	switch name {
	case "ls":
		options = "-l -a -la -1"
	case "cat", "pwd":
	case "head", "tail":
		options = "-n"
	case "wc":
		options = "-l -w -c"
	case "rg", "grep":
		options = "-n -i -l -c -F -w"
	case "find":
		options = "-name -iname -type -maxdepth"
	case "git status":
		options = "--short -s"
	case "git log":
		options = "--oneline -n"
	case "git diff", "git show":
		options = "--stat"
	case "git rev-parse":
		return len(args) == 1 && (args[0] == "HEAD" || args[0] == "--show-toplevel")
	default:
		return false
	}
	for i := 0; i < len(args); i++ {
		option := args[i]
		if !strings.HasPrefix(option, "-") {
			continue
		}
		found := false
		for _, allowed := range strings.Fields(options) {
			if option == allowed {
				found = true
				break
			}
		}
		if !found {
			return false
		}
		integerValue := (name == "head" || name == "tail" || name == "git log") && option == "-n"
		if integerValue || name == "find" {
			i++
			if i >= len(args) || strings.HasPrefix(args[i], "-") {
				return false
			}
			if integerValue {
				for _, digit := range args[i] {
					if digit < '0' || digit > '9' {
						return false
					}
				}
			}
		}
	}
	return true
}

func flashPrimaryHTTPVerified(r result) bool {
	if r.FlashClass == nil || r.FlashClass.PrimaryCompletion == nil {
		return false
	}
	primary := r.FlashClass.PrimaryCompletion
	found := false
	for _, o := range r.FlashClass.Observations {
		if o.SessionID != r.Trace.SessionID || o.Phase != "http" {
			continue
		}
		if o.Code != "" || !o.Correlated || o.Protocol == "unsupported" || o.Agent != primary.Agent || o.Model != r.ModelRequested {
			return false
		}
		found = true
	}
	return found
}
