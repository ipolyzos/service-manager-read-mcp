package svcerr

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors for stable handling across backends and MCP layers.
var (
	ErrBackendUnavailable = errors.New("backend unavailable")
	ErrUnitNotFound       = errors.New("unit not found")
	ErrUnsupported        = errors.New("operation unsupported")
	ErrParse              = errors.New("parse error")
	ErrInvocation         = errors.New("command invocation failed")
)

// InvocationError records a completed subprocess that failed or could not start as expected.
type InvocationError struct {
	Path     string
	Args     []string
	ExitCode int
	Stdout   []byte
	Stderr   []byte
	Cause    error
}

func (e *InvocationError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%v: %s", ErrInvocation, e.Path)
	if len(e.Args) > 0 {
		fmt.Fprintf(&b, " %q", e.Args)
	}
	if e.ExitCode != 0 {
		fmt.Fprintf(&b, " (exit %d)", e.ExitCode)
	}
	if len(e.Stderr) > 0 {
		fmt.Fprintf(&b, ": %s", strings.TrimSpace(string(e.Stderr)))
	}
	return b.String()
}

// Unwrap returns ErrInvocation so errors.Is(..., ErrInvocation) matches.
// Inspect Cause directly when the underlying start/exit error is needed.
func (e *InvocationError) Unwrap() error {
	return ErrInvocation
}

// ParseError wraps stderr or contextual detail for parser failures.
type ParseError struct {
	Backend string
	Detail  string
	Cause   error
}

func (e *ParseError) Error() string {
	if e.Backend != "" && e.Detail != "" {
		return fmt.Sprintf("%v (%s): %s", ErrParse, e.Backend, e.Detail)
	}
	if e.Detail != "" {
		return fmt.Sprintf("%v: %s", ErrParse, e.Detail)
	}
	return ErrParse.Error()
}

func (e *ParseError) Unwrap() error {
	if e.Cause != nil {
		return e.Cause
	}
	return ErrParse
}
