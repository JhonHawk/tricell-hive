package viewer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed static
var staticFiles embed.FS

const (
	contentSecurityPolicy = "default-src 'self'; img-src 'self'; media-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'none'"
	tokenHeader           = "X-Hive-Token"
	minTokenLength        = 16
)

// embeddedAssets are the only static files served under /assets/.
var embeddedAssets = map[string]string{
	"walkthrough.mjs": "text/javascript; charset=utf-8",
	"walkthrough.css": "text/css; charset=utf-8",
	"viewer.mjs":      "text/javascript; charset=utf-8",
	"viewer.css":      "text/css; charset=utf-8",
}

// fileTypes maps the extensions OpenFile can return to a Content-Type. The
// type comes from the extension, never from sniffing the bytes.
var fileTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
	".mp4":  "video/mp4",
}

// ServerOptions configures NewServer.
type ServerOptions struct {
	// Store is the state the server shows and changes. Required.
	Store *Store
	// Token is the secret the CLI and the page send as X-Hive-Token. Required,
	// at least 16 characters.
	Token string
	// Port is the TCP port the listener is bound to, used by the Host and
	// Origin checks. Required (with an ephemeral port, listen first and pass
	// the port that was assigned).
	Port int
	// Version is reported by /api/health.
	Version string
}

// Server is the HTTP face of a Store: the viewer page, its assets and the
// JSON API of design DS4. It holds pending one-use access codes in memory.
// It must be served on a loopback listener (see Listen); the Host check
// protects against DNS rebinding, not against another local user.
type Server struct {
	store       *Store
	token       []byte
	port        string
	version     string
	pid         int
	accessMu    sync.Mutex
	accessCodes map[string]time.Time
	accessNow   func() time.Time
}

// NewServer validates the options and returns the handler.
func NewServer(o ServerOptions) (*Server, error) {
	switch {
	case o.Store == nil:
		return nil, errors.New("viewer: server needs a store")
	case len(o.Token) < minTokenLength:
		return nil, fmt.Errorf("viewer: server token must be at least %d characters", minTokenLength)
	case o.Port < 1 || o.Port > 65535:
		return nil, fmt.Errorf("viewer: invalid server port %d", o.Port)
	}
	return &Server{
		store:       o.Store,
		token:       []byte(o.Token),
		port:        strconv.Itoa(o.Port),
		version:     o.Version,
		pid:         os.Getpid(),
		accessCodes: make(map[string]time.Time),
		accessNow:   time.Now,
	}, nil
}

// Listen opens a TCP listener on 127.0.0.1. Port 0 picks a free port; read it
// back from the listener's address before calling NewServer.
func Listen(port int) (net.Listener, error) {
	return net.Listen("tcp4", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
}

// HTTPServer wraps the handler in an http.Server with timeouts suited to a
// local viewer. WriteTimeout is left unset so a large recording can stream.
func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Handler:           s,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    16 << 10,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("Content-Security-Policy", contentSecurityPolicy)
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("Cross-Origin-Resource-Policy", "same-origin")
	if !s.hostAllowed(r.Host) {
		httpError(w, http.StatusForbidden, "forbidden host")
		return
	}
	p := r.URL.Path
	switch {
	case p == "/":
		s.asset(w, r, "index.html", "text/html; charset=utf-8")
	case strings.HasPrefix(p, "/assets/"):
		name := strings.TrimPrefix(p, "/assets/")
		if ct, ok := embeddedAssets[name]; ok {
			s.asset(w, r, name, ct)
			return
		}
		httpError(w, http.StatusNotFound, "not found")
	case p == "/api/access-codes":
		s.issueAccessCode(w, r)
	case p == "/api/access-codes/exchange":
		s.exchangeAccessCode(w, r)
	case p == "/api/health":
		s.health(w, r)
	case p == "/api/index":
		s.index(w, r)
	case strings.HasPrefix(p, "/api/sessions/"):
		s.session(w, r, strings.TrimPrefix(p, "/api/sessions/"))
	case strings.HasPrefix(p, "/files/"):
		s.file(w, r, strings.TrimPrefix(p, "/files/"))
	default:
		httpError(w, http.StatusNotFound, "not found")
	}
}

func (s *Server) hostAllowed(host string) bool {
	return host == "127.0.0.1:"+s.port || host == "localhost:"+s.port
}

// tokenOK compares in constant time.
func (s *Server) tokenOK(r *http.Request) bool {
	return subtle.ConstantTimeCompare([]byte(r.Header.Get(tokenHeader)), s.token) == 1
}

// originOK: Origin, when sent, must be the server's own origin. A browser
// always sends it on a cross-site or non-GET request; a CLI sends none.
func (s *Server) originOK(r *http.Request, required bool) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return !required
	}
	return origin == "http://"+r.Host
}

// authorize guards a route that needs the token; writes also require Origin.
func (s *Server) authorize(w http.ResponseWriter, r *http.Request, write bool) bool {
	if !s.tokenOK(r) || !s.originOK(r, write) {
		httpError(w, http.StatusForbidden, "forbidden")
		return false
	}
	return true
}

func httpError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	fmt.Fprintln(w, message)
}

func writeJSON(w http.ResponseWriter, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(append(data, '\n'))
}

// storeError maps the Store's sentinel errors to a status without echoing
// internal detail.
func storeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidKey), errors.Is(err, ErrInvalidPath):
		httpError(w, http.StatusBadRequest, "bad request")
	case errors.Is(err, ErrNotFound):
		httpError(w, http.StatusNotFound, "not found")
	default:
		httpError(w, http.StatusInternalServerError, "internal error")
	}
}

func allow(w http.ResponseWriter, r *http.Request, methods ...string) bool {
	for _, m := range methods {
		if r.Method == m {
			return true
		}
	}
	w.Header().Set("Allow", strings.Join(methods, ", "))
	httpError(w, http.StatusMethodNotAllowed, "method not allowed")
	return false
}

func (s *Server) asset(w http.ResponseWriter, r *http.Request, name, contentType string) {
	if !allow(w, r, http.MethodGet, http.MethodHead) {
		return
	}
	data, err := fs.ReadFile(staticFiles, path.Join("static", name))
	if err != nil {
		httpError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if !allow(w, r, http.MethodGet) || !s.authorize(w, r, false) {
		return
	}
	writeJSON(w, struct {
		PID      int    `json:"pid"`
		StateDir string `json:"state_dir"`
		Version  string `json:"version"`
		TTL      string `json:"ttl"`
	}{s.pid, filepath.Dir(s.store.Dir()), s.version, s.store.TTL().String()})
}

// indexSession adds the base of the session's file URLs to a Session. It is
// what the authenticated page mounts the walkthrough with.
type indexSession struct {
	Session
	FileBase string `json:"fileBase"`
}

type indexTicket struct {
	Name     string         `json:"ticket"`
	Sessions []indexSession `json:"sessions"`
}

type indexProject struct {
	Key     string        `json:"key"`
	Name    string        `json:"name"`
	Absent  bool          `json:"absent"`
	Tickets []indexTicket `json:"tickets"`
}

type indexView struct {
	TTLSeconds   int64              `json:"ttl_seconds"`
	Projects     []indexProject     `json:"projects"`
	Repositories []RepositoryStatus `json:"repositories"`
	Errors       []IngestError      `json:"errors"`
	Tombstones   []Tombstone        `json:"tombstones"`
}

// fileCap is the capability for the files of one session: the hex
// HMAC-SHA256 of the key under the server token. An <img> or <video> cannot
// send X-Hive-Token, so /files/<key>/<cap>/<path> carries this instead; it
// only reaches the page through the authenticated index.
func (s *Server) fileCap(key string) string {
	mac := hmac.New(sha256.New, s.token)
	mac.Write([]byte(key))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	if !allow(w, r, http.MethodGet) || !s.authorize(w, r, false) {
		return
	}
	v := s.store.View()
	out := indexView{
		TTLSeconds: v.TTLSeconds, Repositories: v.Repositories, Errors: v.Errors, Tombstones: v.Tombstones,
		Projects: make([]indexProject, 0, len(v.Projects)),
	}
	for _, p := range v.Projects {
		ip := indexProject{Key: p.Key, Name: p.Name, Absent: p.Absent, Tickets: make([]indexTicket, 0, len(p.Tickets))}
		for _, t := range p.Tickets {
			it := indexTicket{Name: t.Name, Sessions: make([]indexSession, 0, len(t.Sessions))}
			for _, sess := range t.Sessions {
				it.Sessions = append(it.Sessions, indexSession{Session: sess, FileBase: "/files/" + sess.Key + "/" + s.fileCap(sess.Key) + "/"})
			}
			ip.Tickets = append(ip.Tickets, it)
		}
		out.Projects = append(out.Projects, ip)
	}
	writeJSON(w, out)
}

// session serves /api/sessions/<key>[/pin].
func (s *Server) session(w http.ResponseWriter, r *http.Request, rest string) {
	key, sub, _ := strings.Cut(rest, "/")
	switch sub {
	case "":
		switch r.Method {
		case http.MethodGet:
			if !s.authorize(w, r, false) {
				return
			}
			data, err := s.store.Manifest(key)
			if err != nil {
				storeError(w, err)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			w.Write(data)
		case http.MethodDelete:
			if s.authorize(w, r, true) {
				s.done(w, s.store.Delete(key))
			}
		default:
			allow(w, r, http.MethodGet, http.MethodDelete)
		}
	case "pin":
		if !allow(w, r, http.MethodPost, http.MethodDelete) || !s.authorize(w, r, true) {
			return
		}
		s.done(w, s.store.Pin(key, r.Method == http.MethodPost))
	default:
		httpError(w, http.StatusNotFound, "not found")
	}
}

func (s *Server) done(w http.ResponseWriter, err error) {
	if err != nil {
		storeError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}

// file serves /files/<key>/<cap>/<path> from a stored copy. The Store validates the
// key and the path (AC2 rules, no links); the Content-Type is fixed by the
// extension.
func (s *Server) file(w http.ResponseWriter, r *http.Request, rest string) {
	if !allow(w, r, http.MethodGet, http.MethodHead) {
		return
	}
	key, tail, _ := strings.Cut(rest, "/")
	capability, rel, found := strings.Cut(tail, "/")
	if !found || subtle.ConstantTimeCompare([]byte(capability), []byte(s.fileCap(key))) != 1 {
		httpError(w, http.StatusForbidden, "forbidden")
		return
	}
	f, err := s.store.OpenFile(key, rel)
	if err != nil {
		storeError(w, err)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		httpError(w, http.StatusInternalServerError, "internal error")
		return
	}
	ct, ok := fileTypes[extension(rel)]
	if !ok {
		httpError(w, http.StatusBadRequest, "bad request")
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, "", info.ModTime(), f)
}
