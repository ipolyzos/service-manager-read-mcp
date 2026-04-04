package svcerr_test

import (
	"errors"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestSentinels(t *testing.T) {
	t.Parallel()
	cases := []error{
		svcerr.ErrBackendUnavailable,
		svcerr.ErrUnitNotFound,
		svcerr.ErrUnsupported,
		svcerr.ErrParse,
		svcerr.ErrInvocation,
		svcerr.ErrNotAllowlisted,
		svcerr.ErrInvalidArgument,
		svcerr.ErrOutputLimit,
	}
	for _, err := range cases {
		if err == nil || err.Error() == "" {
			t.Fatalf("expected non-empty error: %v", err)
		}
	}
}

func TestErrorsIs(t *testing.T) {
	t.Parallel()
	inv := &svcerr.InvocationError{Path: "/bin/true", Args: []string{"-h"}, ExitCode: 2, Stderr: []byte("nope")}
	if !errors.Is(inv, svcerr.ErrInvocation) {
		t.Fatalf("invocation should unwrap to ErrInvocation")
	}
	parse := &svcerr.ParseError{Backend: "systemd", Detail: "short read", Cause: svcerr.ErrParse}
	if !errors.Is(parse, svcerr.ErrParse) {
		t.Fatalf("parse should unwrap to ErrParse")
	}
	wrapped := errors.Join(svcerr.ErrUnitNotFound, svcerr.ErrUnsupported)
	if !errors.Is(wrapped, svcerr.ErrUnitNotFound) {
		t.Fatalf("join should preserve Is for first error")
	}
}

func TestInvocationErrorErrorString(t *testing.T) {
	t.Parallel()
	e := &svcerr.InvocationError{
		Path:     "/bin/systemctl",
		Args:     []string{"show", "nginx.service"},
		ExitCode: 1,
		Stderr:   []byte("Access denied\n"),
		Cause:    svcerr.ErrInvocation,
	}
	s := e.Error()
	if s == "" {
		t.Fatal("expected message")
	}
}
