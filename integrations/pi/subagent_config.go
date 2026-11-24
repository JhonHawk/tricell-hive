package pi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SubagentConfigRel is pi-subagents' own configuration file under the Pi home
// (pi-subagents 0.83.0 docs/configuration.md; the extension resolves the home
// with getAgentDir(), which honors PI_CODING_AGENT_DIR).
const SubagentConfigRel = "extensions/subagent/config.json"

// ToolActivationKey and EagerActivation are the one setting Hive manages in
// that file: with "eager" the lazy subagents_enable loader is not registered
// and the subagent tool is active from a session's first request.
const (
	ToolActivationKey = "toolActivation"
	EagerActivation   = "eager"
)

// MaxSubagentConfigBytes bounds the file Hive edits, as the Codex wait floor
// bounds config.toml.
const MaxSubagentConfigBytes = 1 << 20

// SubagentConfigStatus is what the file says about toolActivation.
type SubagentConfigStatus int

const (
	// ConfigAbsent means the file does not exist.
	ConfigAbsent SubagentConfigStatus = iota
	// KeyMissing means the file is a JSON object without toolActivation.
	KeyMissing
	// KeyPresent means the object holds toolActivation, with any value.
	KeyPresent
	// ConfigUnsupported means the file is not one JSON object with unique
	// keys, so Hive does not edit it.
	ConfigUnsupported
)

// SubagentConfigRefusal is a fixed reason Hive does not touch the file; it
// never repeats the path.
type SubagentConfigRefusal struct{ Cause string }

func (r *SubagentConfigRefusal) Error() string { return r.Cause }

func refuseConfig(cause string) error { return &SubagentConfigRefusal{Cause: cause} }

const eagerMember = `"` + ToolActivationKey + `": "` + EagerActivation + `"`

var createdConfig = []byte("{\n  " + eagerMember + "\n}\n")

type configMember struct {
	key      string
	keyStart int
	valueEnd int
	raw      json.RawMessage
}

// parseConfigObject reads data as one JSON object with unique keys and returns
// the offset right after its opening brace and each member's byte span.
func parseConfigObject(data []byte) (int, []configMember, bool) {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		return 0, nil, false
	}
	open := int(dec.InputOffset())
	var members []configMember
	seen := map[string]bool{}
	for dec.More() {
		before := int(dec.InputOffset())
		tok, err := dec.Token()
		key, isKey := tok.(string)
		if err != nil || !isKey || seen[key] {
			return 0, nil, false
		}
		seen[key] = true
		rel := bytes.IndexByte(data[before:], '"')
		if rel < 0 {
			return 0, nil, false
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return 0, nil, false
		}
		members = append(members, configMember{key: key, keyStart: before + rel, valueEnd: int(dec.InputOffset()), raw: raw})
	}
	if tok, err := dec.Token(); err != nil || tok != json.Delim('}') {
		return 0, nil, false
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return 0, nil, false
	}
	return open, members, true
}

// ClassifySubagentConfig reports how the file's bytes hold toolActivation and,
// when present, its value as JSON text, such as "\"auto\"" or "null".
func ClassifySubagentConfig(data []byte, exists bool) (SubagentConfigStatus, string) {
	if !exists {
		return ConfigAbsent, ""
	}
	_, members, ok := parseConfigObject(data)
	if !ok {
		return ConfigUnsupported, ""
	}
	for _, m := range members {
		if m.key == ToolActivationKey {
			return KeyPresent, strings.TrimSpace(string(m.raw))
		}
	}
	return KeyMissing, ""
}

// IsEagerValue reports whether a value ClassifySubagentConfig returned is the
// one Hive writes.
func IsEagerValue(value string) bool { return value == `"`+EagerActivation+`"` }

// SubagentConfigEmpty reports whether data is a JSON object with no members.
func SubagentConfigEmpty(data []byte) bool {
	_, members, ok := parseConfigObject(data)
	return ok && len(members) == 0
}

// InsertToolActivation returns data with "toolActivation": "eager" appended as
// the object's last member, every other byte kept, and the new member indented
// like the member before it. Empty data, or an empty object, becomes the file
// Hive creates.
func InsertToolActivation(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return append([]byte(nil), createdConfig...), nil
	}
	open, members, ok := parseConfigObject(data)
	if !ok {
		return nil, refuseConfig("the pi-subagents config is not a JSON object with unique keys")
	}
	for _, m := range members {
		if m.key == ToolActivationKey {
			return nil, fmt.Errorf("%s is already set", ToolActivationKey)
		}
	}
	if len(members) == 0 {
		return append([]byte(nil), createdConfig...), nil
	}
	last := members[len(members)-1]
	sep := data[open:last.keyStart]
	if len(members) > 1 {
		region := data[members[len(members)-2].valueEnd:last.keyStart]
		sep = region[bytes.IndexByte(region, ',')+1:]
	}
	var out bytes.Buffer
	out.Write(data[:last.valueEnd])
	out.WriteByte(',')
	out.Write(sep)
	out.WriteString(eagerMember)
	out.Write(data[last.valueEnd:])
	return out.Bytes(), nil
}

// RemoveToolActivation returns data without its toolActivation member, every
// other byte kept, and whether a member was removed.
func RemoveToolActivation(data []byte) ([]byte, bool, error) {
	open, members, ok := parseConfigObject(data)
	if !ok {
		return nil, false, refuseConfig("the pi-subagents config is not a JSON object with unique keys")
	}
	i := -1
	for j, m := range members {
		if m.key == ToolActivationKey {
			i = j
		}
	}
	var out []byte
	switch {
	case i < 0:
		return data, false, nil
	case len(members) == 1:
		out = append(append([]byte(nil), data[:open]...), data[members[0].valueEnd:]...)
	case i == 0:
		out = append(append([]byte(nil), data[:members[0].keyStart]...), data[members[1].keyStart:]...)
	default:
		out = append(append([]byte(nil), data[:members[i-1].valueEnd]...), data[members[i].valueEnd:]...)
	}
	return out, true, nil
}

// CheckSubagentConfigTarget maps the file system facts about the config file
// under the Pi home to refusals, without reading the file, and reports whether
// it exists. It refuses when the file, or a folder Hive would write through, is
// a symbolic link or not of its kind, when the file is larger than 1 MiB or not
// writable by the owner, and when the folder Hive would create the next entry
// in is not writable. Ancestors of the Pi home are not inspected.
func CheckSubagentConfigTarget(piHome string) (bool, error) {
	info, err := os.Lstat(piHome)
	switch {
	case os.IsNotExist(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("inspect Pi home: %w", err)
	case info.Mode()&os.ModeSymlink != 0:
		return false, refuseConfig("the Pi home folder is a symbolic link")
	case !info.IsDir():
		return false, refuseConfig("the Pi home is not a folder")
	}
	parent := piHome
	for _, name := range []string{"extensions", "subagent"} {
		dir := filepath.Join(parent, name)
		info, err := os.Lstat(dir)
		if os.IsNotExist(err) {
			return false, writableFolder(parent)
		}
		if err != nil {
			return false, fmt.Errorf("inspect pi-subagents config folder: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return false, refuseConfig("a pi-subagents config folder is a symbolic link")
		}
		if !info.IsDir() {
			return false, refuseConfig("a pi-subagents config folder is not a folder")
		}
		parent = dir
	}
	info, err = os.Lstat(filepath.Join(parent, "config.json"))
	if os.IsNotExist(err) {
		return false, writableFolder(parent)
	}
	if err != nil {
		return false, fmt.Errorf("inspect pi-subagents config: %w", err)
	}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		return true, refuseConfig("the pi-subagents config is a symbolic link")
	case !info.Mode().IsRegular():
		return true, refuseConfig("the pi-subagents config is not a regular file")
	case info.Size() > MaxSubagentConfigBytes:
		return true, refuseConfig("the pi-subagents config is larger than 1 MiB")
	case info.Mode().Perm()&0200 == 0:
		return true, refuseConfig("the pi-subagents config is not writable")
	}
	return true, nil
}

func writableFolder(dir string) error {
	if info, err := os.Stat(dir); err == nil && info.Mode().Perm()&0200 == 0 {
		return refuseConfig("the pi-subagents config folder is not writable")
	}
	return nil
}
