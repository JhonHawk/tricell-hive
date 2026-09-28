package management

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"tricell-hive/integrations/target"
)

type snapshot struct {
	Exists     bool
	Tree       []treeItem `json:",omitempty"`
	Data       []byte
	Mode       uint32
	Kind       string `json:",omitempty"`
	LinkTarget string `json:",omitempty"`
	FileMode   uint32 `json:",omitempty"`
}

func read(path string) (snapshot, error) {
	if err := target.Safe(path); err != nil {
		return snapshot{}, err
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return snapshot{}, nil
	}
	if err != nil {
		return snapshot{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return snapshot{}, err
	}
	if !info.Mode().IsRegular() {
		return snapshot{}, fmt.Errorf("not a regular file: %s", path)
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat.Nlink > 1 {
		return snapshot{}, fmt.Errorf("hard-link conflict: %s", path)
	}
	b, err := io.ReadAll(f)
	return snapshot{Exists: true, Data: b, Mode: uint32(info.Mode().Perm())}, err
}
func finger(s snapshot) Fingerprint {
	treeHash := ""
	if len(s.Tree) > 0 {
		treeHash = hash(encode(s.Tree))
	}
	return Fingerprint{TreeHash: treeHash, Exists: s.Exists, Hash: hash(s.Data), Mode: s.Mode, Kind: s.Kind, LinkTarget: s.LinkTarget, FileMode: s.FileMode}
}
func same(a, b snapshot) bool { return finger(a) == finger(b) }

// removeOrphanWrites deletes ".hive-write-*" temporaries that a killed process
// left in the state directory or next to a journaled destination. It runs under
// the state lock, so no live writer owns them; failures are ignored because the
// files are inert and recovery must not depend on them.
func removeOrphanWrites(stateDir string, j journal) {
	dirs := map[string]bool{stateDir: true, filepath.Join(stateDir, "transactions"): true, filepath.Join(stateDir, "onboarding"): true}
	for _, en := range j.Entries {
		dirs[filepath.Dir(en.Change.Target.Path)] = true
	}
	for dir := range dirs {
		matches, _ := filepath.Glob(filepath.Join(dir, ".hive-write-*"))
		for _, m := range matches {
			if info, err := os.Lstat(m); err == nil && info.Mode().IsRegular() {
				_ = os.Remove(m)
			}
		}
	}
}

func write(path string, s snapshot) error {
	if err := target.Safe(path); err != nil {
		return err
	}
	if !s.Exists {
		err := os.Remove(path)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".hive-write-*")
	if err != nil {
		return err
	}
	temp := f.Name()
	defer os.Remove(temp)
	if err = f.Chmod(os.FileMode(s.Mode)); err == nil {
		_, err = f.Write(s.Data)
	}
	if err == nil {
		err = f.Sync()
	}
	cerr := f.Close()
	if err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	if err = target.Safe(path); err != nil {
		return err
	}
	if err = os.Rename(temp, path); err != nil {
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
func writeJSON(path string, v any) error {
	return write(path, snapshot{Exists: true, Data: encode(v), Mode: 0600})
}
func decodeFile(path string, v any) error {
	s, err := read(path)
	if err != nil {
		return err
	}
	if !s.Exists {
		return os.ErrNotExist
	}
	dec := json.NewDecoder(bytes.NewReader(s.Data))
	dec.DisallowUnknownFields()
	if err = dec.Decode(v); err != nil {
		return err
	}
	var extra any
	if err = dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("trailing or invalid JSON")
	}
	return nil
}
func readState(dir string) (State, string, error) {
	s, err := read(filepath.Join(dir, "state.json"))
	if err != nil {
		return State{}, "", err
	}
	state := emptyState()
	if s.Exists {
		state = State{}
		if err = json.Unmarshal(s.Data, &state); err != nil {
			return state, "", err
		}
		if (state.Version != 1 && state.Version != 2 && state.Version != 3 && state.Version != 4 && state.Version != 5 && state.Version != stateVersion) || state.Records == nil {
			return state, "", fmt.Errorf("unsupported state")
		}
		if err = validateProductState(state); err != nil {
			return state, "", err
		}
		if err = normalizeState(&state); err != nil {
			return state, "", err
		}
	}
	return state, hash(s.Data), nil
}
func lock(dir string) (func(), error) {
	if err := target.Safe(dir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "lock")
	if err := target.Safe(path); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("another Hive operation holds the lock")
	}
	return func() { syscall.Flock(int(f.Fd()), syscall.LOCK_UN); f.Close() }, nil
}

// blockRange requires complete, unambiguous delimiter lines, including EOF.
func blockRange(b []byte, m markers) (int, int, error) {
	text := string(b)
	bc, ec := strings.Count(text, m.begin), strings.Count(text, m.end)
	if bc == 0 && ec == 0 {
		return -1, -1, nil
	}
	if bc != 1 || ec != 1 {
		return 0, 0, fmt.Errorf("malformed or duplicate %s markers", m.name)
	}
	start, end := strings.Index(text, m.begin), strings.Index(text, m.end)
	if end < start {
		return 0, 0, fmt.Errorf("reversed %s markers", m.name)
	}
	for _, r := range [][2]int{{start, start + len(m.begin)}, {end, end + len(m.end)}} {
		if r[0] > 0 && text[r[0]-1] != '\n' {
			return 0, 0, fmt.Errorf("marker is not a full line")
		}
		tail := text[r[1]:]
		if tail != "" && !strings.HasPrefix(tail, "\n") && !strings.HasPrefix(tail, "\r\n") {
			return 0, 0, fmt.Errorf("marker is not a full line")
		}
	}
	end += len(m.end)
	if strings.HasPrefix(text[end:], "\r\n") {
		end += 2
	} else if strings.HasPrefix(text[end:], "\n") {
		end++
	}
	return start, end, nil
}
func managedBlock(body, current []byte, m markers) []byte {
	nl := "\n"
	if i := bytes.IndexByte(current, '\n'); i > 0 && current[i-1] == '\r' {
		nl = "\r\n"
	}
	s := strings.ReplaceAll(string(body), "\r\n", "\n")
	s = strings.TrimRight(s, "\n")
	return []byte(m.begin + nl + strings.ReplaceAll(s, "\n", nl) + nl + m.end + nl)
}
func owned(s snapshot, r Record, m markers) error {
	if !s.Exists {
		return fmt.Errorf("managed file is missing: %s", r.Target.Path)
	}
	if r.Target.Kind == "symlink" {
		if s.Kind != "symlink" || s.LinkTarget != r.Target.LinkTarget {
			return fmt.Errorf("modified managed symlink: %s", r.Target.Path)
		}
		return nil
	}
	if s.Kind != "" {
		return fmt.Errorf("managed resource type changed: %s", r.Target.Path)
	}
	if r.Target.Kind == "skill" || r.Target.Kind == "agent" {
		if !bytes.Equal(s.Data, r.Managed) || (r.Mode != 0 && s.Mode != r.Mode) {
			return fmt.Errorf("modified managed skill: %s", r.Target.Path)
		}
		return nil
	}
	a, b, err := blockRange(s.Data, m)
	if err != nil {
		return err
	}
	if a < 0 || !bytes.Equal(s.Data[a:b], r.Managed) {
		return fmt.Errorf("modified or missing managed block: %s", r.Target.Path)
	}
	return nil
}
func transform(s snapshot, before, after *Record, m markers) (snapshot, error) {
	if before != nil {
		if err := owned(s, *before, m); err != nil {
			return snapshot{}, err
		}
	}
	if before == nil && after == nil {
		return s, nil
	}
	mode := s.Mode
	if !s.Exists {
		mode = 0600
	}
	if after != nil && after.Mode != 0 {
		mode = after.Mode
	}
	if after != nil && (after.Target.Kind == "skill" || after.Target.Kind == "agent") {
		if before == nil && s.Exists {
			return snapshot{}, fmt.Errorf("unowned skill collision")
		}
		return snapshot{Exists: true, Data: after.Managed, Mode: mode}, nil
	}
	if before != nil && (before.Target.Kind == "skill" || before.Target.Kind == "agent") {
		return snapshot{}, nil
	}
	if after != nil && after.Target.Kind == "symlink" {
		if before == nil && s.Exists {
			return snapshot{}, fmt.Errorf("unowned symlink collision")
		}
		return snapshot{Exists: true, Kind: "symlink", LinkTarget: after.Target.LinkTarget}, nil
	}
	if before != nil && before.Target.Kind == "symlink" {
		return snapshot{}, nil
	}
	a, b, err := blockRange(s.Data, m)
	if err != nil {
		return snapshot{}, err
	}
	if before == nil {
		if a >= 0 {
			return snapshot{}, fmt.Errorf("unowned %s block", m.name)
		}
		out := append(append(append([]byte{}, s.Data...), []byte(after.Leading)...), after.Managed...)
		return snapshot{Exists: true, Data: out, Mode: mode}, nil
	}
	if after != nil {
		return snapshot{Exists: true, Data: bytes.Join([][]byte{s.Data[:a], after.Managed, s.Data[b:]}, nil), Mode: mode}, nil
	}
	if before.Leading != "" && a >= len(before.Leading) && string(s.Data[a-len(before.Leading):a]) == before.Leading {
		a -= len(before.Leading)
	}
	out := append(append([]byte{}, s.Data[:a]...), s.Data[b:]...)
	if before.CreatedFile && len(out) == 0 {
		return snapshot{}, nil
	}
	return snapshot{Exists: true, Data: out, Mode: mode}, nil
}
