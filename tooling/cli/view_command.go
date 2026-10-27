package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/version"
	"tricell-hive/tooling/viewer"
)

// The viewer server (`hive view`) is a long-lived local process owned by the
// user: start, stop and status manage it through <state>/viewer/server.lock
// (held by the one running server) and server.json (how to reach it). The
// state itself lives in tooling/viewer; this file is the command line around
// it and injects the readers that need cli helpers.

const (
	viewStartTimeout = 5 * time.Second
	viewHealthLimit  = 2 * time.Second
	viewPollStep     = 50 * time.Millisecond
	viewInfoLimit    = 64 << 10

	// viewTestServeEnv makes the test binary run as the server (see TestMain).
	// Only while it is set does `serve` read viewPollIntervalEnv, so the poll
	// interval is not a user-facing setting.
	viewTestServeEnv       = "HIVE_VIEW_TEST_SERVE"
	viewPollIntervalEnv    = "HIVE_VIEW_POLL_INTERVAL"
	viewShutdownTimeoutEnv = "HIVE_VIEW_SHUTDOWN_TIMEOUT"
)

const (
	viewStopGrace      = 3 * time.Second // stop waits this long beyond the server's shutdown
	viewLockRetryLimit = time.Second     // serve waits this long for a momentary probe of server.lock
)

// viewShutdownTimeout is how long serve lets active requests finish after a
// signal. It is a variable only so tests can shorten it.
var viewShutdownTimeout = 5 * time.Second

// viewStopWait is how long stop waits for the lock after SIGTERM: the whole
// server shutdown plus a margin for the process to exit.
func viewStopWait() time.Duration { return viewShutdownTimeout + viewStopGrace }

// viewLauncher starts the detached server process with the given arguments
// (after the executable) and its output going to logPath, and returns its PID.
type viewLauncher func(args []string, logPath string) (int, error)

// viewOpener opens a URL with the system opener.
type viewOpener func(url string) error

// viewDeps are the effects of `hive view` that tests replace.
type viewDeps struct {
	launch viewLauncher
	open   viewOpener
}

func realViewDeps() viewDeps {
	return viewDeps{launch: defaultViewLauncher, open: defaultViewOpener}
}

// serverInfo is <state>/viewer/server.json (mode 0600).
type serverInfo struct {
	PID      int    `json:"pid"`
	Port     int    `json:"port"`
	Version  string `json:"version"`
	Token    string `json:"token"`
	TTL      string `json:"ttl"`
	StateDir string `json:"state_dir"`
	Started  string `json:"started"`
}

func (i serverInfo) url() string { return fmt.Sprintf("http://127.0.0.1:%d", i.Port) }

// viewPaths are the files of one state directory that the commands share.
type viewPaths struct{ state, dir, lock, info, log string }

func newViewPaths(state string) viewPaths {
	dir := filepath.Join(state, "viewer")
	return viewPaths{
		state: state,
		dir:   dir,
		lock:  filepath.Join(dir, "server.lock"),
		info:  filepath.Join(dir, "server.json"),
		log:   filepath.Join(dir, "server.log"),
	}
}

// viewFlags are the flags beyond --home and --state-dir; each subcommand
// registers the ones it accepts.
type viewFlags struct {
	port    int
	portSet bool // --port was given, so a busy port is an error
	ttl     string
	dir     string
}

func runView(args []string, w io.Writer, deps viewDeps) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		printViewHelp(w)
		return nil
	}
	if len(args) == 0 {
		return errors.New("view requires start, stop, status, open, register or serve")
	}
	sub, rest := args[0], args[1:]
	fs := flag.NewFlagSet("view "+sub, flag.ContinueOnError)
	var f viewFlags
	switch sub {
	case "start", "serve":
		fs.IntVar(&f.port, "port", viewDefaultPort, "TCP port on 127.0.0.1; 0 picks a free port")
		fs.StringVar(&f.ttl, "ttl", "", "how long an unpinned copy lives, e.g. 72h (default 14 days)")
	case "register":
		fs.StringVar(&f.dir, "dir", "", "directory inside the repository to register (default: the current directory)")
	case "status":
		fs.StringVar(&f.dir, "dir", "", "walkthrough source folder to confirm against the running viewer")
	case "stop", "open":
	default:
		return fmt.Errorf("unknown view subcommand %q", sub)
	}
	o, ok, err := readOnlyOptions(fs, rest)
	if !ok {
		return err
	}
	fs.Visit(func(fl *flag.Flag) {
		if fl.Name == "port" {
			f.portSet = true
		}
	})
	_, state, err := management.NormalizeOptions(o)
	if err != nil {
		return err
	}
	if state, err = filepath.Abs(state); err != nil {
		return err
	}
	p := newViewPaths(state)
	switch sub {
	case "start":
		return viewStart(p, f, w, deps)
	case "serve":
		return viewServe(p, f, w)
	case "stop":
		return viewStop(p, w)
	case "status":
		return viewStatus(p, w, f.dir)
	case "open":
		return viewOpen(p, w, deps)
	default:
		return viewRegister(p, f, w)
	}
}

func parseViewTTL(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("invalid --ttl %q: use a positive duration such as 72h", raw)
	}
	return d, nil
}

// viewDefaultPort spells HIVE on a phone keypad.
const viewDefaultPort = 4483

// listenView binds the port. A port is busy when the bind fails with
// EADDRINUSE or when a connection to it answers: on macOS a 127.0.0.1 bind
// succeeds over another program listening on the wildcard address and would
// shadow it. A busy --port fails; a busy default falls back to a free port.
func listenView(f viewFlags, w io.Writer) (net.Listener, error) {
	if f.port != 0 {
		addr := net.JoinHostPort("127.0.0.1", fmt.Sprint(f.port))
		if conn, err := net.DialTimeout("tcp4", addr, viewHealthLimit); err == nil {
			conn.Close()
		} else if ln, err := viewer.Listen(f.port); !errors.Is(err, syscall.EADDRINUSE) {
			return ln, err
		}
		if f.portSet {
			return nil, fmt.Errorf("port %d is busy", f.port)
		}
		fmt.Fprintf(w, "port %d is busy; listening on a free port\n", f.port)
	}
	return viewer.Listen(0)
}

func checkViewPort(port int) error {
	if port < 0 || port > 65535 {
		return fmt.Errorf("invalid --port %d", port)
	}
	return nil
}

// viewLockHeld reports whether a server owns the state directory. It never
// creates the lock file: no file means no server.
func viewLockHeld(p viewPaths) (bool, error) {
	if _, err := os.Stat(p.lock); errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	lock, err := viewer.AcquireLock(p.lock, false)
	switch {
	case errors.Is(err, viewer.ErrLocked):
		return true, nil
	case err != nil:
		return false, err
	}
	return false, lock.Release()
}

func readServerInfo(path string) (serverInfo, error) {
	var info serverInfo
	st, err := os.Lstat(path)
	if err != nil {
		return info, err
	}
	if !st.Mode().IsRegular() {
		return info, fmt.Errorf("%s is not a regular file", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return info, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, viewInfoLimit))
	if err != nil {
		return info, err
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return info, fmt.Errorf("%s: %w", path, err)
	}
	return info, nil
}

var viewClient = &http.Client{
	Timeout:   viewHealthLimit,
	Transport: &http.Transport{Proxy: nil},
}

// viewHealth asks the server described by info for /api/health with its
// token. It succeeds only when the answer carries the same PID and state
// directory as info, so a stale server.json or an unrelated listener on the
// port never counts as the server.
func viewHealth(info serverInfo, p viewPaths) error {
	if info.Port <= 0 || info.PID <= 0 || info.Token == "" {
		return errors.New("server.json is incomplete")
	}
	req, err := http.NewRequest(http.MethodGet, info.url()+"/api/health", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Hive-Token", info.Token)
	resp, err := viewClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health answered %d", resp.StatusCode)
	}
	var h struct {
		PID      int    `json:"pid"`
		StateDir string `json:"state_dir"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, viewInfoLimit)).Decode(&h); err != nil {
		return err
	}
	if h.PID != info.PID || filepath.Clean(h.StateDir) != filepath.Clean(p.state) {
		return errors.New("health does not match server.json")
	}
	return nil
}

// viewRunning returns the server's info when the lock is held and its health
// matches server.json.
func viewRunning(p viewPaths) (serverInfo, bool, error) {
	held, err := viewLockHeld(p)
	if err != nil || !held {
		return serverInfo{}, false, err
	}
	info, err := readServerInfo(p.info)
	if err != nil || viewHealth(info, p) != nil {
		return serverInfo{}, false, nil
	}
	return info, true, nil
}

func viewStart(p viewPaths, f viewFlags, w io.Writer, deps viewDeps) error {
	if _, err := parseViewTTL(f.ttl); err != nil {
		return err
	}
	if err := checkViewPort(f.port); err != nil {
		return err
	}
	held, err := viewLockHeld(p)
	if err != nil {
		return err
	}
	if held {
		if info, err := readServerInfo(p.info); err == nil {
			fmt.Fprintf(w, "The viewer is already running at %s (pid %d).\n", info.url(), info.PID)
		} else {
			fmt.Fprintln(w, "The viewer is already running (server.json is not readable yet).")
		}
		return nil
	}
	// The lock is free, so a server.json left here is an orphan.
	if err := os.Remove(p.info); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(p.dir, 0o700); err != nil {
		return err
	}
	args := []string{"view", "serve", "--state-dir", p.state}
	if f.portSet {
		args = append(args, "--port", fmt.Sprint(f.port))
	}
	if f.ttl != "" {
		args = append(args, "--ttl", f.ttl)
	}
	pid, err := deps.launch(args, p.log)
	if err != nil {
		return fmt.Errorf("cannot launch the viewer: %w", err)
	}
	deadline := time.Now().Add(viewStartTimeout)
	for {
		if info, err := readServerInfo(p.info); err == nil && (pid <= 0 || info.PID == pid) && viewHealth(info, p) == nil {
			if !f.portSet && info.Port != viewDefaultPort {
				fmt.Fprintf(w, "The default port %d is busy, so the viewer took a free port.\n", viewDefaultPort)
			}
			fmt.Fprintf(w, "The viewer is running at %s (pid %d). Run `hive view open` to open it.\n", info.url(), info.PID)
			return nil
		}
		if pid > 0 && syscall.Kill(pid, 0) == syscall.ESRCH {
			return fmt.Errorf("the viewer exited during startup; see %s", p.log)
		}
		if time.Now().After(deadline) {
			if pid > 0 {
				_ = syscall.Kill(pid, syscall.SIGTERM) // our own child, never answered
			}
			return fmt.Errorf("the viewer did not answer within %s; see %s", viewStartTimeout, p.log)
		}
		time.Sleep(viewPollStep)
	}
}

func viewServe(p viewPaths, f viewFlags, w io.Writer) error {
	ttl, err := parseViewTTL(f.ttl)
	if err != nil {
		return err
	}
	if err := checkViewPort(f.port); err != nil {
		return err
	}
	if err := os.MkdirAll(p.dir, 0o700); err != nil {
		return err
	}
	// start, stop and status probe server.lock by taking it for an instant, so
	// a lock that is busy for a moment is retried before it means "owned".
	var lock *viewer.FileLock
	for deadline := time.Now().Add(viewLockRetryLimit); ; time.Sleep(20 * time.Millisecond) {
		lock, err = viewer.AcquireLock(p.lock, false)
		if !errors.Is(err, viewer.ErrLocked) {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("another viewer server already owns %s", p.state)
		}
	}
	if err != nil {
		return err
	}
	defer lock.Release()

	git := newGitRunner()
	store, err := viewer.Open(viewer.Options{
		StateDir:       p.state,
		TTL:            ttl,
		HiveProject:    viewHiveProject,
		GitRemote:      viewGitRemote(git),
		SourceTracking: viewSourceTracking(runViewerTrackingGit),
	})
	if err != nil {
		return err
	}
	ln, err := listenView(f, w)
	if err != nil {
		return err
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	token, err := newViewToken()
	if err != nil {
		return err
	}
	srv, err := viewer.NewServer(viewer.ServerOptions{Store: store, Token: token, Port: port, Version: version.Current})
	if err != nil {
		return err
	}
	info := serverInfo{
		PID: os.Getpid(), Port: port, Version: version.Current, Token: token,
		TTL: store.TTL().String(), StateDir: p.state, Started: time.Now().UTC().Format(time.RFC3339),
	}
	if err := writeServerInfo(p, info); err != nil {
		return err
	}
	defer os.Remove(p.info)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	runOpts := viewer.RunOptions{OnError: func(err error) {
		fmt.Fprintf(os.Stderr, "%s viewer: %v\n", time.Now().UTC().Format(time.RFC3339), err)
	}}
	if os.Getenv(viewTestServeEnv) == "1" {
		if d, err := time.ParseDuration(os.Getenv(viewShutdownTimeoutEnv)); err == nil && d > 0 {
			viewShutdownTimeout = d
		}
		if d, err := time.ParseDuration(os.Getenv(viewPollIntervalEnv)); err == nil && d > 0 {
			runOpts.PollInterval = d
		}
	}
	// Run performs the one Startup pass itself. A failure of that pass ends
	// Run at once and the select below exits serve with its error, removing
	// server.json, so a broken state directory still fails fast.
	runDone := make(chan error, 1)
	go func() { runDone <- store.Run(ctx, runOpts) }()
	httpSrv := srv.HTTPServer()
	serveErr := make(chan error, 1)
	go func() { serveErr <- httpSrv.Serve(ln) }()
	fmt.Fprintf(w, "viewer listening on %s (pid %d)\n", info.url(), info.PID)

	var result error
	select {
	case <-ctx.Done():
	case result = <-serveErr:
	case result = <-runDone:
		if result == nil {
			result = errors.New("the store loop stopped")
		}
	}
	stop()
	shutdown, cancel := context.WithTimeout(context.Background(), viewShutdownTimeout)
	defer cancel()
	_ = httpSrv.Shutdown(shutdown)
	return result
}

func newViewToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// writeServerInfo replaces server.json atomically with mode 0600.
func writeServerInfo(p viewPaths, info serverInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(p.dir, ".server-json-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), p.info)
}

func viewStop(p viewPaths, w io.Writer) error {
	held, err := viewLockHeld(p)
	if err != nil {
		return err
	}
	if !held {
		if err := os.Remove(p.info); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		fmt.Fprintln(w, "The viewer is not running.")
		return nil
	}
	info, err := readServerInfo(p.info)
	if err != nil {
		return fmt.Errorf("the server lock is held but server.json is unreadable (%v); no signal was sent", err)
	}
	if err := viewHealth(info, p); err != nil {
		return fmt.Errorf("the server lock is held but pid %d from server.json did not confirm it is the viewer (%v); no signal was sent and nothing was removed", info.PID, err)
	}
	if err := syscall.Kill(info.PID, syscall.SIGTERM); err != nil {
		return fmt.Errorf("cannot signal pid %d: %w", info.PID, err)
	}
	deadline := time.Now().Add(viewStopWait())
	for {
		held, err := viewLockHeld(p)
		if err != nil {
			return err
		}
		if !held {
			_ = os.Remove(p.info)
			fmt.Fprintln(w, "The viewer stopped.")
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("pid %d did not release the server lock within %s", info.PID, viewStopWait())
		}
		time.Sleep(viewPollStep)
	}
}

func viewStatus(p viewPaths, w io.Writer, dir string) error {
	out := map[string]any{"running": false}
	var source viewer.SourceFingerprint
	var sourceErr error
	if dir != "" {
		out["found"], out["session_status"] = false, "unknown"
		source, sourceErr = viewer.FingerprintSource(dir)
	}
	info, running, err := viewRunning(p)
	if err != nil {
		return err
	}
	switch {
	case running:
		out["running"] = true
		out["url"] = info.url()
		out["pid"] = info.PID
		out["version"] = info.Version
		out["ttl"] = info.TTL
		if idx, err := viewIndex(info); err == nil {
			count := 0
			for _, project := range idx.Projects {
				for _, ticket := range project.Tickets {
					count += len(ticket.Sessions)
				}
			}
			out["repositories"], out["walkthroughs"], out["ingest_errors"] = len(idx.Repositories), count, len(idx.Errors)
			if dir != "" && sourceErr == nil {
				after, err := viewer.FingerprintSource(dir)
				if err == nil && after == source && viewHasSource(idx, source) {
					out["found"], out["session_status"] = true, source.SessionStatus
				}
			}
		}
	default:
		if held, _ := viewLockHeld(p); held {
			out["problem"] = "the server lock is held but the server does not answer as the viewer"
			if stale, err := readServerInfo(p.info); err == nil {
				out["pid"] = stale.PID
			}
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// viewHasSource matches every identity and revision field, including explicit
// status. A legacy index without session_status cannot confirm a revision.
func viewHasSource(v viewer.View, source viewer.SourceFingerprint) bool {
	matches := func(s viewer.Session) bool {
		return s.Source == source.Source && s.SessionID == source.SessionID && s.Key == source.Key &&
			s.ContentHash == source.ContentHash && s.SessionStatus == source.SessionStatus
	}
	for _, s := range v.ExcludedSessions {
		if matches(s) {
			return true
		}
	}
	for _, project := range v.Projects {
		for _, ticket := range project.Tickets {
			for _, s := range ticket.Sessions {
				if matches(s) {
					return true
				}
			}
		}
	}
	return false
}

// viewIndex reads the existing authenticated index without opening a Store.
// Redirects must not forward the token to another listener.
func viewIndex(info serverInfo) (viewer.View, error) {
	var v viewer.View
	req, err := http.NewRequest(http.MethodGet, info.url()+"/api/index", nil)
	if err != nil {
		return v, err
	}
	req.Header.Set("X-Hive-Token", info.Token)
	client := *viewClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return v, errors.New("cannot read viewer index")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return v, fmt.Errorf("index answered %d", resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<20)).Decode(&v); err != nil {
		return v, errors.New("invalid viewer index")
	}
	return v, nil
}

func viewOpen(p viewPaths, w io.Writer, deps viewDeps) error {
	info, running, err := viewRunning(p)
	if err != nil {
		return err
	}
	if !running {
		fmt.Fprintln(w, "The viewer is not running. Start it with `hive view start`.")
		return nil
	}
	code, err := viewAccessCode(info)
	if err != nil {
		return err
	}
	if err := deps.open(info.url() + "/#code=" + code); err != nil {
		return errors.New("cannot open the viewer; run `hive view open` to try again")
	}
	fmt.Fprintf(w, "Opened %s\n", info.url())
	return nil
}

// viewAccessCode never forwards response bodies or transport errors, which may
// contain credentials. Redirects must not forward the token to another server.
func viewAccessCode(info serverInfo) (string, error) {
	req, err := http.NewRequest(http.MethodPost, info.url()+"/api/access-codes", nil)
	if err != nil {
		return "", errors.New("cannot request viewer access")
	}
	req.Header.Set("X-Hive-Token", info.Token)
	req.Header.Set("Origin", info.url())
	client := *viewClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("cannot request viewer access; run `hive view open` to try again")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", errors.New("the running viewer needs an update; run `hive view stop`, `hive view start`, then `hive view open`")
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("the viewer did not grant access; run `hive view open` to try again")
	}
	var result struct {
		Code string `json:"code"`
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1025))
	if err != nil || len(data) > 1024 {
		return "", errors.New("invalid viewer access response")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if dec.Decode(&result) != nil || dec.Decode(new(any)) != io.EOF || len(result.Code) != 64 {
		return "", errors.New("invalid viewer access response")
	}
	decoded, err := hex.DecodeString(result.Code)
	if err != nil || len(decoded) != 32 {
		return "", errors.New("invalid viewer access response")
	}
	return result.Code, nil
}

func viewRegister(p viewPaths, f viewFlags, w io.Writer) error {
	dir := f.dir
	if dir == "" {
		var err error
		if dir, err = os.Getwd(); err != nil {
			return err
		}
	}
	res, err := viewer.RegisterRepository(p.state, dir)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}
	if res.Added {
		fmt.Fprintf(w, "Registered %s\n", res.Root)
	} else {
		fmt.Fprintf(w, "Already registered %s\n", res.Root)
	}
	return nil
}

// viewGitRemote is the viewer's GitRemoteReader. Only a Git toplevel has a
// remote of its own: when `git rev-parse --show-toplevel` in root does not
// give root itself (not a repository, or a root that merely lives inside
// another one, such as a non-Git workspace in a versioned home), there is no
// remote and the error is the reader's "no remote" answer, never an ingest
// error.
func viewGitRemote(git gitRunner) viewer.GitRemoteReader {
	return func(root string) (string, error) {
		top, err := git(root, "rev-parse", "--show-toplevel")
		if err != nil {
			return "", err
		}
		if !sameDir(strings.TrimSpace(top), root) {
			return "", errors.New("not a Git toplevel")
		}
		out, err := git(root, "config", "--get", "remote.origin.url")
		if err != nil {
			return "", err
		}
		if out = strings.TrimSpace(out); out == "" {
			return "", errors.New("no origin remote")
		}
		return out, nil
	}
}

// sameDir reports whether a and b name the same directory after resolving
// links.
func sameDir(a, b string) bool {
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	return errA == nil && errB == nil && ra == rb
}

// viewHiveProject is the viewer's HiveProjectReader: the `Project:` value of
// the `## Hive` section of the repository's AGENTS.md.
func viewHiveProject(root string) (string, bool) {
	// readProjectFile refuses a link, a non-regular file and one over 1 MiB.
	data, existed, _, err := readProjectFile(filepath.Join(root, "AGENTS.md"))
	if err != nil || !existed {
		return "", false
	}
	items, _ := parseHiveSection(string(data))
	for _, it := range items {
		if it.Key == "Project" {
			return it.Value, it.Value != ""
		}
	}
	return "", false
}

// defaultViewLauncher starts this executable detached in its own session so
// the server outlives the terminal; its output goes to logPath.
func defaultViewLauncher(args []string, logPath string) (int, error) {
	exe, err := os.Executable()
	if err != nil {
		return 0, err
	}
	logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return 0, err
	}
	defer logFile.Close()
	cmd := exec.Command(exe, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, logFile, logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	go func() { _ = cmd.Wait() }() // reap it if it exits while this process lives
	return cmd.Process.Pid, nil
}

// defaultViewOpener hands the URL to the system opener.
func defaultViewOpener(url string) error {
	var name string
	switch runtime.GOOS {
	case "darwin":
		name = "open"
	case "linux":
		name = "xdg-open"
	default:
		return fmt.Errorf("no system opener for %s", runtime.GOOS)
	}
	cmd := exec.Command(name, url)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
