package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Read only the launch definition through the caller's typed projection. Never
// return raw parse errors, which can quote values from private configuration.
// When both native files exist, precedence is not established by this tooling.
func readOpenCodeNativeConfig(directory string) ([]byte, error) {
	var selected []byte
	found := false
	for _, name := range []string{"opencode.json", "opencode.jsonc"} {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("real OpenCode configuration could not be read")
		}
		if found {
			return nil, fmt.Errorf("both native OpenCode configuration files exist; precedence is unverified")
		}
		found, selected = true, data
	}
	if !found {
		return nil, fmt.Errorf("real OpenCode configuration is unavailable")
	}
	return normalizeNativeJSONC(selected)
}

// The native format accepts comments and trailing commas. Preserve every byte
// inside JSON strings; normalize syntax only, before encoding/json validation.
func normalizeNativeJSONC(input []byte) ([]byte, error) {
	out := make([]byte, 0, len(input))
	quoted, escaped := false, false
	for i := 0; i < len(input); i++ {
		c := input[i]
		if quoted {
			out = append(out, c)
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				quoted = false
			}
			continue
		}
		if c == '"' {
			quoted = true
		}
		if c == '/' && i+1 < len(input) {
			if input[i+1] == '/' {
				for i+1 < len(input) && input[i+1] != '\n' {
					i++
				}
				out = append(out, ' ')
				continue
			}
			if input[i+1] == '*' {
				i += 2
				for i+1 < len(input) && !(input[i] == '*' && input[i+1] == '/') {
					i++
				}
				if i+1 >= len(input) {
					return nil, fmt.Errorf("real OpenCode configuration has an unterminated comment")
				}
				i++
				out = append(out, ' ')
				continue
			}
		}
		out = append(out, c)
	}
	quoted, escaped = false, false
	for i, c := range out {
		if quoted {
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				quoted = false
			}
			continue
		}
		if c == '"' {
			quoted = true
		}
		if c != ',' {
			continue
		}
		j := i + 1
		for j < len(out) && (out[j] == ' ' || out[j] == '\t' || out[j] == '\n' || out[j] == '\r') {
			j++
		}
		previous := i - 1
		for previous >= 0 && (out[previous] == ' ' || out[previous] == '\t' || out[previous] == '\n' || out[previous] == '\r') {
			previous--
		}
		if previous >= 0 && out[previous] != '{' && out[previous] != '[' && out[previous] != ',' && out[previous] != ':' && j < len(out) && (out[j] == '}' || out[j] == ']') {
			out[i] = ' '
		}
	}
	return out, nil
}
