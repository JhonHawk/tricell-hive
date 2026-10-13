package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Patch only the model scalar in the installed shadow role, without serializing
// its frontmatter or rewriting its instructions and unrelated fields.
func overrideFlashRoleModel(shadowHome, role, model string) (beforeHash, afterHash, originalModel string, err error) {
	if role != "hive-review-plan" && role != "hive-review-harness" {
		return "", "", "", fmt.Errorf("unsupported flash role")
	}
	root, err := os.OpenRoot(shadowHome)
	if err != nil {
		return "", "", "", fmt.Errorf("shadow role directory unavailable")
	}
	defer root.Close()
	name := filepath.Join(".config", "opencode", "agents", role+".md")
	info, err := root.Lstat(name)
	if err != nil || !info.Mode().IsRegular() {
		return "", "", "", fmt.Errorf("shadow role must be a regular file")
	}
	raw, err := root.ReadFile(name)
	if err != nil {
		return "", "", "", fmt.Errorf("shadow role unavailable")
	}
	lines := bytes.SplitAfter(raw, []byte("\n"))
	clean := func(line []byte) string { return strings.TrimSuffix(strings.TrimSuffix(string(line), "\n"), "\r") }
	if len(lines) < 2 || clean(lines[0]) != "---" {
		return "", "", "", fmt.Errorf("shadow role frontmatter required")
	}
	offset := len(lines[0])
	closing := -1
	start, end := -1, -1
	replacement := ""
	newline := "\n"
	if bytes.HasSuffix(lines[0], []byte("\r\n")) {
		newline = "\r\n"
	}
	for _, line := range lines[1:] {
		text := clean(line)
		if text == "---" {
			closing = offset
			break
		}
		if flashRoleModelDeclaration(text) {
			if !strings.HasPrefix(text, "model:") {
				return "", "", "", fmt.Errorf("shadow role model key must use plain model: form")
			}
			if start >= 0 {
				return "", "", "", fmt.Errorf("shadow role model is ambiguous")
			}
			valueStart := len("model:")
			for valueStart < len(text) && (text[valueStart] == ' ' || text[valueStart] == '\t') {
				valueStart++
			}
			valueEnd := len(text)
			value := text[valueStart:]
			if len(value) > 0 && (value[0] == '\'' || value[0] == '"') {
				quote := value[0]
				closed := false
				for i := 1; i < len(value); i++ {
					if quote == '"' && value[i] == '\\' {
						i++
						continue
					}
					if value[i] != quote {
						continue
					}
					if quote == '\'' && i+1 < len(value) && value[i+1] == '\'' {
						i++
						continue
					}
					valueEnd = valueStart + i + 1
					closed = true
					break
				}
				if !closed {
					return "", "", "", fmt.Errorf("shadow role model scalar invalid")
				}
				scalar := text[valueStart:valueEnd]
				if quote == '"' {
					originalModel, err = strconv.Unquote(scalar)
					if err != nil {
						return "", "", "", fmt.Errorf("shadow role model scalar invalid")
					}
					replacement = strconv.Quote(model)
				} else {
					originalModel = strings.ReplaceAll(scalar[1:len(scalar)-1], "''", "'")
					replacement = "'" + strings.ReplaceAll(model, "'", "''") + "'"
				}
			} else {
				for i := 0; i < len(value); i++ {
					if value[i] == '#' && (i == 0 || value[i-1] == ' ' || value[i-1] == '\t') {
						valueEnd = valueStart + i
						break
					}
				}
				for valueEnd > valueStart && (text[valueEnd-1] == ' ' || text[valueEnd-1] == '\t') {
					valueEnd--
				}
				originalModel = text[valueStart:valueEnd]
				replacement = model
			}
			start, end = offset+valueStart, offset+valueEnd
		}
		offset += len(line)
	}
	if closing < 0 {
		return "", "", "", fmt.Errorf("shadow role frontmatter closing delimiter required")
	}
	if start < 0 {
		start, end = closing, closing
		replacement = "model: " + strconv.Quote(model) + newline
	}
	patched := make([]byte, 0, len(raw)+len(replacement))
	patched = append(patched, raw[:start]...)
	patched = append(patched, replacement...)
	patched = append(patched, raw[end:]...)
	if err := root.WriteFile(name, patched, info.Mode().Perm()); err != nil {
		return "", "", "", fmt.Errorf("shadow role model override failed")
	}
	return digest(raw), digest(patched), originalModel, nil
}

// Recognize equivalent YAML scalar keys before deciding whether the supported
// plain spelling is safe to patch. Quoted aliases must never cause insertion.
func flashRoleModelDeclaration(line string) bool {
	key, _, ok := strings.Cut(line, ":")
	if !ok {
		return false
	}
	key = strings.TrimSpace(key)
	if len(key) >= 2 && key[0] == '"' && key[len(key)-1] == '"' {
		decoded, err := strconv.Unquote(key)
		return err == nil && decoded == "model"
	}
	if len(key) >= 2 && key[0] == '\'' && key[len(key)-1] == '\'' {
		key = strings.ReplaceAll(key[1:len(key)-1], "''", "'")
	}
	return key == "model"
}
