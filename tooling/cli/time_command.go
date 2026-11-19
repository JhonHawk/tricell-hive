package main

// time_command.go is `hive time`: the active time of a session, read from the
// transcript its host keeps. The reading rules per host are in time_read.go; the
// contract is the requirement «hive time reports a session's active time» of
// the workflow-lifecycle spec.

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"tricell-hive/tooling/management"
)

const (
	timeDefaultFactor = 2.0
	timeMinFactor     = 1.0
	timeMaxFactor     = 3.0
)

// exitError makes main exit with a code other than 1.
type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string { return e.msg }

func timeInputError(format string, a ...any) error {
	return &exitError{code: 2, msg: fmt.Sprintf(format, a...)}
}

type timeResult struct {
	Result           string  `json:"result"`
	Host             string  `json:"host"`
	Session          string  `json:"session"`
	Since            string  `json:"since"`
	Until            string  `json:"until"`
	ActiveSeconds    int64   `json:"active_seconds"`
	UserWaitSeconds  int64   `json:"user_wait_seconds"`
	CIWaitSeconds    int64   `json:"ci_wait_seconds"`
	Factor           float64 `json:"factor"`
	ConvertedSeconds int64   `json:"converted_seconds"`
	Note             string  `json:"note,omitempty"`
}

// timeHosts are the hosts whose transcript `hive time` reads.
var timeHosts = map[string]bool{"claude-code": true, "codex": true, "grok": true, "pi": true}

func runTime(args []string, w io.Writer) error {
	fs := flag.NewFlagSet("time", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var opts management.Options
	var sinceRaw, host, file, sessionID string
	factor := timeDefaultFactor
	fs.StringVar(&sinceRaw, "since", "", "start of the window, RFC3339 (default: the session's first event)")
	fs.Float64Var(&factor, "factor", timeDefaultFactor, "multiplier from active time to human time, 1 to 3")
	fs.StringVar(&file, "session-file", "", "read this transcript instead of the current session's")
	fs.StringVar(&sessionID, "session", "", "native session ID of an earlier session, with --host: its transcript is located by that ID")
	fs.StringVar(&host, "host", "", "host of --session-file or --session: claude-code, codex, grok or pi")
	fs.StringVar(&opts.Home, "home", "", "explicit synthetic home; ignores host environment paths")
	fs.StringVar(&opts.StateDir, "state-dir", "", "state directory (default: user Application Support/tricell-hive)")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return timeInputError("time: %v", err)
	}
	if fs.NArg() != 0 {
		return timeInputError("time: unexpected positional arguments")
	}
	if math.IsNaN(factor) || factor < timeMinFactor || factor > timeMaxFactor {
		return timeInputError("time: --factor must be between %g and %g", timeMinFactor, timeMaxFactor)
	}
	var since time.Time
	if sinceRaw != "" {
		var err error
		if since, err = time.Parse(time.RFC3339, sinceRaw); err != nil {
			return timeInputError("time: --since must be RFC3339")
		}
	}
	if file != "" && sessionID != "" {
		return timeInputError("time: --session-file and --session exclude each other")
	}
	if (file != "" || sessionID != "") != (host != "") {
		return timeInputError("time: --host goes with --session-file or --session, and only with them")
	}
	if host != "" && !timeHosts[host] {
		return timeInputError("time: --host must be one of claude-code, codex, grok, pi")
	}
	opts.Scope = "user"
	session := sessionID
	switch {
	case sessionID != "":
		var err error
		if file, err = locateHostSessionFile(opts, host, sessionID); err != nil {
			return timeInputError("time: %v", err)
		}
	case file == "":
		_, state, err := management.NormalizeOptions(opts)
		if err != nil {
			return err
		}
		if state, err = filepath.Abs(state); err != nil {
			return err
		}
		id, err := resolveViewIdentity(viewIdentityDepsFn(), state)
		switch {
		case err == nil:
			host, session = id.Host, id.NativeSessionID
		case errors.Is(err, errViewIdentityAmbiguous):
			return errors.New("time: native identity ambiguous: pass --session-file and --host")
		case errors.Is(err, errViewIdentityMissing):
			return errors.New("time: native identity missing: pass --session-file and --host")
		default:
			return errors.New("time: native identity unresolved: pass --session-file and --host")
		}
		if timeHosts[host] {
			if file, err = locateHostSessionFile(opts, host, session); err != nil {
				return err
			}
		}
	default:
		session = timeSessionName(host, file)
	}
	res := timeResult{Result: "ok", Host: host, Session: session, Factor: factor}
	if file == "" {
		res.Result = "no_session_file"
		res.Note = "this host keeps no transcript Hive reads, or none was found for the session: use the marks"
		return writeTimeResult(w, res)
	}
	facts := readTimeFile(host, file)
	if facts.first.IsZero() {
		facts.markPartial("the session file holds no timed event")
	}
	if facts.partial {
		res.Result = "partial"
		res.Note = facts.note
	}
	fillTimeResult(&res, facts, since)
	return writeTimeResult(w, res)
}

func writeTimeResult(w io.Writer, r timeResult) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(r)
}

// timeSessionName is the native session ID an explicit transcript names.
func timeSessionName(host, path string) string {
	base := strings.TrimSuffix(filepath.Base(path), ".jsonl")
	switch host {
	case "codex":
		if len(base) >= 36 {
			return base[len(base)-36:]
		}
	case "grok":
		return filepath.Base(filepath.Dir(path))
	case "pi":
		if i := strings.LastIndex(base, "_"); i >= 0 {
			return base[i+1:]
		}
	}
	return base
}

// fillTimeResult computes the totals of the window from the facts read.
func fillTimeResult(res *timeResult, f timeFacts, since time.Time) {
	if f.first.IsZero() {
		return
	}
	from, until := f.first, f.last
	if !since.IsZero() && since.After(from) {
		from = since
	}
	if until.Before(from) {
		until = from
	}
	res.Since = from.UTC().Format(time.RFC3339)
	res.Until = until.UTC().Format(time.RFC3339)
	user := mergeSpans(f.waits, from, until)
	both := mergeSpans(append(append([]timeSpan{}, f.waits...), f.ci...), from, until)
	userLen := spansLength(user)
	ciLen := spansLength(both) - userLen
	activeLen := until.Sub(from) - userLen
	res.ActiveSeconds = int64(activeLen.Round(time.Second) / time.Second)
	res.UserWaitSeconds = int64(userLen.Round(time.Second) / time.Second)
	res.CIWaitSeconds = int64(ciLen.Round(time.Second) / time.Second)
	res.ConvertedSeconds = convertTime(res.ActiveSeconds, res.CIWaitSeconds, res.Factor)
}

// convertTime is (active - ci) * factor + ci, rounded up to the minute.
func convertTime(active, ci int64, factor float64) int64 {
	conv := float64(active-ci)*factor + float64(ci)
	minutes := math.Ceil(conv/60 - 1e-9)
	if minutes < 0 {
		minutes = 0
	}
	return int64(minutes) * 60
}

type timeSpan struct{ start, end time.Time }

// mergeSpans clips the spans to [from, until] and merges the overlapping ones.
func mergeSpans(spans []timeSpan, from, until time.Time) []timeSpan {
	var out []timeSpan
	for _, s := range spans {
		if s.start.Before(from) {
			s.start = from
		}
		if s.end.After(until) {
			s.end = until
		}
		if s.end.After(s.start) {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].start.Before(out[j].start) })
	var merged []timeSpan
	for _, s := range out {
		if n := len(merged); n > 0 && !s.start.After(merged[n-1].end) {
			if s.end.After(merged[n-1].end) {
				merged[n-1].end = s.end
			}
			continue
		}
		merged = append(merged, s)
	}
	return merged
}

func spansLength(spans []timeSpan) time.Duration {
	var d time.Duration
	for _, s := range spans {
		d += s.end.Sub(s.start)
	}
	return d
}
