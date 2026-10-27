package viewer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// sessionIdentity matches the frontend's JSON tuple; an absent host is null.
type sessionIdentity struct {
	ProjectKey string  `json:"project_key"`
	Host       *string `json:"host"`
	SessionID  string  `json:"session_id"`
}

type sessionGroup struct {
	Version       int             `json:"version"`
	Identity      sessionIdentity `json:"identity"`
	TitleOverride string          `json:"title_override,omitempty"`
	Archived      bool            `json:"archived"`
	Pinned        bool            `json:"pinned"`
	DeletePending bool            `json:"delete_pending,omitempty"`
}

func identityOf(sess Session) sessionIdentity {
	id := sessionIdentity{ProjectKey: sess.ProjectKey, SessionID: sess.SessionID}
	if sess.Host != "" {
		host := sess.Host
		id.Host = &host
	}
	return id
}

func (id sessionIdentity) key() string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode([]any{id.ProjectKey, id.Host, id.SessionID})
	canonical := strings.TrimSuffix(b.String(), "\n")
	canonical = strings.ReplaceAll(strings.ReplaceAll(canonical, `\u2028`, "\u2028"), `\u2029`, "\u2029")
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

func (s *Store) groupsDir() string { return filepath.Join(s.vdir, "session-groups") }
func (s *Store) groupPath(id sessionIdentity) string {
	return filepath.Join(s.groupsDir(), id.key()+".json")
}

func validateSessionTitle(title string) (string, error) {
	if !utf8.ValidString(title) {
		return "", errors.New("session title must be valid UTF-8")
	}
	for _, r := range title {
		if unicode.IsControl(r) {
			return "", errors.New("session title must not contain control characters")
		}
	}
	title = strings.TrimSpace(title)
	if utf8.RuneCountInString(title) > maxNameRunes {
		return "", errors.New("session title exceeds 200 Unicode code points")
	}
	return title, nil
}

func (s *Store) readGroup(id sessionIdentity) (sessionGroup, error) {
	var g sessionGroup
	data, err := readBounded(s.groupPath(id), maxMetaBytes)
	if err != nil {
		return g, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return g, err
	}
	for _, name := range []string{"version", "identity", "archived", "pinned"} {
		if raw, ok := fields[name]; !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return g, fmt.Errorf("session group missing %s", name)
		}
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&g); err != nil {
		return g, err
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return g, errors.New("session group contains trailing JSON")
	}
	title, err := validateSessionTitle(g.TitleOverride)
	if err != nil || title != g.TitleOverride || g.Version != 1 || g.Identity.key() != id.key() {
		return g, errors.New("invalid session group record")
	}
	return g, nil
}

func (s *Store) writeGroup(g sessionGroup) error {
	if err := os.MkdirAll(s.groupsDir(), 0o700); err != nil {
		return err
	}
	if info, err := os.Lstat(s.groupsDir()); err != nil || !info.IsDir() {
		return errors.New("session-groups must be a real directory")
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return s.writeAtomic(s.groupPath(g.Identity), append(data, '\n'))
}

// writeMeta retires legacy preferences. They are accepted only as import input.
func writeMeta(dir string, m meta) error {
	return writeJSONAtomic(filepath.Join(dir, "meta.json"), struct {
		Source      string    `json:"source"`
		ContentHash string    `json:"content_hash"`
		IngestedAt  time.Time `json:"ingested_at"`
		ProjectKey  string    `json:"project_key,omitempty"`
		ProjectName string    `json:"project_name,omitempty"`
	}{m.Source, m.ContentHash, m.IngestedAt, m.ProjectKey, m.ProjectName})
}

// reconcileGroups imports before retiring legacy fields, and overlays the
// durable authority. The read-only path never writes or repairs a library.
func (s *Store) reconcileGroups(d *indexData, persist bool) error {
	// Re-evaluate preference errors so repaired records clear on the next pass.
	kept := d.Errors[:0:0]
	for _, problem := range d.Errors {
		if problem.Path != s.groupsDir() && !strings.HasPrefix(problem.Path, s.groupsDir()+string(filepath.Separator)) {
			kept = append(kept, problem)
		}
	}
	d.Errors = kept
	groups := map[string][]int{}
	for i, sess := range d.Sessions {
		key := identityOf(sess).key()
		groups[key] = append(groups[key], i)
	}
	var first error
	for _, members := range groups {
		id := identityOf(d.Sessions[members[0]])
		g, err := s.readGroup(id)
		if errors.Is(err, fs.ErrNotExist) {
			g = sessionGroup{Version: 1, Identity: id, Archived: true}
			for _, i := range members {
				g.Archived = g.Archived && d.Sessions[i].Archived
				g.Pinned = g.Pinned || d.Sessions[i].Pinned
			}
			if persist {
				err = s.writeGroup(g)
			} else {
				err = nil
			}
		}
		if err == nil && persist {
			for _, i := range members {
				dir := s.sessionDir(d.Sessions[i].Key)
				m, readErr := readMeta(dir)
				if readErr != nil {
					err = readErr
					break
				}
				// Retire even explicit false legacy fields without rewriting clean metadata.
				data, readErr := readBounded(filepath.Join(dir, "meta.json"), maxMetaBytes)
				if readErr != nil {
					err = readErr
					break
				}
				var fields map[string]json.RawMessage
				_ = json.Unmarshal(data, &fields)
				_, pin := fields["pinned"]
				_, archive := fields["archived"]
				if pin || archive {
					if err = writeMeta(dir, m); err != nil {
						break
					}
				}
			}
		}
		if err != nil {
			problem := fmt.Errorf("session preferences: %w", err)
			d.Errors = append(d.Errors, IngestError{Path: s.groupPath(id), Reason: problem.Error()})
			if first == nil {
				first = problem
			}
			continue
		}
		for _, i := range members {
			d.Sessions[i].Pinned = g.Pinned
			d.Sessions[i].Archived = g.Archived
			d.Sessions[i].SessionTitle = g.TitleOverride
		}
	}
	return first
}

func (s *Store) updateGroup(key string, change func(*sessionGroup)) error {
	if err := validKey(key); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.current().clone()
	if err := s.reconcileGroups(d, true); err != nil {
		return err
	}
	i, ok := d.session(key)
	if !ok {
		return ErrNotFound
	}
	g, err := s.readGroup(identityOf(d.Sessions[i]))
	if err != nil {
		return err
	}
	change(&g)
	if err := s.writeGroup(g); err != nil {
		return err
	}
	return s.commit(d)
}

// SetTitle sets the local name for every member without changing source titles.
func (s *Store) SetTitle(key, title string) error {
	title, err := validateSessionTitle(title)
	if err != nil {
		return err
	}
	return s.updateGroup(key, func(g *sessionGroup) { g.TitleOverride = title })
}
