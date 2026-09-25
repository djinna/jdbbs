package srv

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// Deadlines for the external converters. A manuscript that makes pandoc,
// typst or a python helper spin for ever must not hold a build slot (there
// are only maxConcurrentBuilds) or a request goroutine indefinitely. These
// are generous: the 314-page image-heavy smoke book compiles in well under a
// minute on the VM.
const (
	toolTimeoutPandoc = 5 * time.Minute
	toolTimeoutTypst  = 10 * time.Minute
	toolTimeoutPython = 5 * time.Minute // preflight, style markers, corrections, template generation
	toolTimeoutImage  = 2 * time.Minute // ImageMagick convert, fontTools subset (per file)
	toolTimeoutQuick  = 60 * time.Second
)

// toolCommand builds an exec.Cmd that is killed — together with any children
// it spawned (pandoc's lua filters, python subprocesses) — when timeout
// elapses or parent is cancelled. The command runs in its own process group
// so the kill cannot miss a grandchild that would otherwise keep the job
// directory busy. Callers must call the returned cancel func once the
// command has finished (defer is fine).
func toolCommand(parent context.Context, timeout time.Duration, name string, args ...string) (*exec.Cmd, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		// Negative pid = the whole process group.
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
			return cmd.Process.Kill()
		}
		return nil
	}
	// If the group somehow survives SIGKILL long enough for pipes to stay
	// open, stop waiting on them after a short grace period.
	cmd.WaitDelay = 5 * time.Second
	return cmd, cancel
}

// toolErr turns a deadline expiry into a message the build log and the
// customer-facing failure note can show plainly, instead of "signal: killed".
func toolErr(cmd *exec.Cmd, err error) error {
	if err == nil {
		return nil
	}
	if cmd != nil && cmd.ProcessState != nil && cmd.ProcessState.Sys() != nil {
		if ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok && ws.Signaled() && ws.Signal() == syscall.SIGKILL {
			return fmt.Errorf("%s: timed out or was killed (the document took too long to process)", filepath.Base(cmd.Path))
		}
	}
	return err
}
