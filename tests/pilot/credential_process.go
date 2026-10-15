package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// The supervisor stays alive after the native command finishes. Its unreaped
// PID reserves the owned process-group ID until the entire group is stopped;
// no termination is attempted against a stale, reusable parent PID.
type credentialProcess struct {
	cmd                                          *exec.Cmd
	status, statusWriter, controlReader, control *os.File
	mu                                           sync.Mutex
	finished                                     bool
}

func superviseCredentialProcess(ctx context.Context, native *exec.Cmd) (*credentialProcess, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	cR, cW, err := os.Pipe()
	if err != nil {
		r.Close()
		w.Close()
		return nil, err
	}
	script := `exec 5<&0
"$@" <&5 3>&- 4<&- &
child=$!
wait "$child"
status=$?
exec 0<&- 1>&- 5<&-
printf '%s\n' "$status" >&3
exec 3>&-
IFS= read -r release <&4
exit "$status"`
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", script, "hive-auth-supervisor", native.Path)
	cmd.Args = append(cmd.Args, native.Args[1:]...)
	cmd.Dir, cmd.Env = native.Dir, native.Env
	cmd.ExtraFiles = []*os.File{w, cR}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return &credentialProcess{cmd: cmd, status: r, statusWriter: w, controlReader: cR, control: cW}, nil
}

func (p *credentialProcess) start() error {
	err := p.cmd.Start()
	p.statusWriter.Close()
	p.controlReader.Close()
	return err
}

func (p *credentialProcess) nativeStatus(ctx context.Context) (int, error) {
	type outcome struct {
		status int
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		data, err := io.ReadAll(io.LimitReader(p.status, 16))
		if err != nil {
			done <- outcome{0, fmt.Errorf("credential completion channel failed")}
			return
		}
		value, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil || value < 0 || value > 255 {
			done <- outcome{0, fmt.Errorf("native credential completion unverified")}
			return
		}
		done <- outcome{value, nil}
	}()
	select {
	case result := <-done:
		return result.status, result.err
	case <-ctx.Done():
		p.status.Close()
		return 0, fmt.Errorf("native credential transfer timed out")
	}
}

func (p *credentialProcess) stop() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.finished {
		return true
	}
	if p.cmd.Process == nil {
		p.close()
		p.finished = true
		return true
	}
	// Do not Wait (and release the leader PID) before terminating descendants.
	if err := syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
		return false
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		// Enumerate explicit group IDs: procps -g selects sessions on Linux.
		cmd := exec.CommandContext(ctx, "ps", "-A", "-o", "pgid=,stat=")
		states, err := cmd.Output()
		cancel()
		if err != nil {
			return false
		}
		live, err := credentialGroupLive(states, p.cmd.Process.Pid)
		if err != nil {
			return false
		}
		if !live {
			_ = p.cmd.Wait()
			p.close()
			p.finished = true
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func credentialGroupLive(data []byte, pgid int) (bool, error) {
	// An all-process enumeration includes ps itself. Empty or malformed output
	// cannot prove absence, even when the command reports success.
	if pgid <= 0 || len(strings.TrimSpace(string(data))) == 0 {
		return false, fmt.Errorf("credential process-group enumeration unverified")
	}
	live := false
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return false, fmt.Errorf("credential process-group enumeration malformed")
		}
		group, err := strconv.Atoi(fields[0])
		state := fields[1]
		if err != nil || group < 0 || strings.Trim(fields[0], "0123456789") != "" {
			return false, fmt.Errorf("credential process-group enumeration malformed")
		}
		// Darwin ps can report '?' during unrelated process churn. Its
		// validated group identity excludes it; only owned states prove stop.
		if group != pgid {
			continue
		}
		if strings.IndexByte("RSDTtZXxIWUP", state[0]) < 0 || strings.Trim(state[1:], "<>ANELSVWXsl+") != "" {
			return false, fmt.Errorf("credential process-group state unverified")
		}
		if state[0] != 'Z' {
			live = true
		}
	}
	return live, nil
}

func (p *credentialProcess) close() {
	p.status.Close()
	p.statusWriter.Close()
	p.controlReader.Close()
	p.control.Close()
}
