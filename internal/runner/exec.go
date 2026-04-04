package runner

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// Exec runs subprocesses with explicit argv, deadlines, and bounded capture of stdout/stderr.
type Exec struct{}

// Run implements Runner using os/exec. Non-zero exits return Result populated together with
// *svcerr.InvocationError (errors.Is(..., svcerr.ErrInvocation)).
func (Exec) Run(ctx context.Context, s Spec) (Result, error) {
	if s.Path == "" {
		return Result{}, errors.New("runner: empty path")
	}

	maxOut := s.MaxStdoutBytes
	if maxOut <= 0 {
		maxOut = DefaultMaxStreamBytes
	}
	maxErr := s.MaxStderrBytes
	if maxErr <= 0 {
		maxErr = DefaultMaxStreamBytes
	}

	if s.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, s.Path, s.Args...)
	if s.Dir != "" {
		cmd.Dir = s.Dir
	}
	if len(s.Env) > 0 {
		cmd.Env = s.Env
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return Result{}, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return Result{}, err
	}

	if err := cmd.Start(); err != nil {
		return Result{}, err
	}

	var stdout, stderr []byte
	var limErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		b, err := readLimited(stdoutPipe, maxOut)
		stdout = b
		if errors.Is(err, svcerr.ErrOutputLimit) {
			limErr = err
		}
	}()
	go func() {
		defer wg.Done()
		b, err := readLimited(stderrPipe, maxErr)
		stderr = b
		if errors.Is(err, svcerr.ErrOutputLimit) {
			limErr = err
		}
	}()
	wg.Wait()

	if limErr != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return Result{
			Stdout:   stdout,
			Stderr:   stderr,
			ExitCode: -1,
		}, limErr
	}

	waitErr := cmd.Wait()
	res := Result{
		Stdout: stdout,
		Stderr: stderr,
	}
	if waitErr == nil {
		return res, nil
	}

	var exit *exec.ExitError
	if errors.As(waitErr, &exit) {
		res.ExitCode = exit.ExitCode()
		return res, &svcerr.InvocationError{
			Path:     s.Path,
			Args:     s.Args,
			ExitCode: res.ExitCode,
			Stdout:   append([]byte(nil), res.Stdout...),
			Stderr:   append([]byte(nil), res.Stderr...),
			Cause:    waitErr,
		}
	}
	return res, waitErr
}

func readLimited(r io.Reader, limit int) ([]byte, error) {
	if limit <= 0 {
		limit = DefaultMaxStreamBytes
	}
	data, err := io.ReadAll(io.LimitReader(r, int64(limit)+1))
	if len(data) > limit {
		return data[:limit], svcerr.ErrOutputLimit
	}
	return data, err
}

// WithTimeout returns a copy of s with Timeout set if t > 0.
func WithTimeout(s Spec, t time.Duration) Spec {
	if t <= 0 {
		return s
	}
	s.Timeout = t
	return s
}
