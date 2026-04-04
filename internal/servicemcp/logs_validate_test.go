package servicemcp

import (
	"errors"
	"testing"
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestParseRecentLogsBounds_serviceRequiredAndValidated(t *testing.T) {
	t.Parallel()
	_, err := ParseRecentLogsBounds(domain.ManagerSystemd, "", 10, "")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("empty service: got %v", err)
	}
	_, err = ParseRecentLogsBounds(domain.ManagerSystemd, "bad name!", 10, "")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("invalid unit: got %v", err)
	}
}

func TestParseRecentLogsBounds_maxLines(t *testing.T) {
	t.Parallel()
	_, err := ParseRecentLogsBounds(domain.ManagerSystemd, "nginx.service", -1, "")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("negative max_lines: got %v", err)
	}
	_, err = ParseRecentLogsBounds(domain.ManagerSystemd, "nginx.service", 501, "")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("max_lines too large: got %v", err)
	}
	got, err := ParseRecentLogsBounds(domain.ManagerSystemd, "nginx.service", 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.MaxLines != DefaultRecentLogLines {
		t.Fatalf("default max lines: got %d", got.MaxLines)
	}
	got, err = ParseRecentLogsBounds(domain.ManagerSystemd, "nginx.service", 42, "")
	if err != nil || got.MaxLines != 42 {
		t.Fatalf("got %v %#v", err, got)
	}
}

func TestParseRecentLogsBounds_recentWindow(t *testing.T) {
	t.Parallel()
	got, err := ParseRecentLogsBounds(domain.ManagerSystemd, "nginx.service", 10, "")
	if err != nil || got.Window != 0 {
		t.Fatalf("empty window: %#v err=%v", got, err)
	}
	got, err = ParseRecentLogsBounds(domain.ManagerSystemd, "nginx.service", 10, "0")
	if err != nil || got.Window != 0 {
		t.Fatalf("zero window: %#v err=%v", got, err)
	}
	got, err = ParseRecentLogsBounds(domain.ManagerSystemd, "nginx.service", 10, "  1h  ")
	if err != nil || got.Window != time.Hour {
		t.Fatalf("1h: %#v err=%v", got, err)
	}
	_, err = ParseRecentLogsBounds(domain.ManagerSystemd, "nginx.service", 10, "not-a-duration")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("bad duration: got %v", err)
	}
	_, err = ParseRecentLogsBounds(domain.ManagerSystemd, "nginx.service", 10, "-1h")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("negative duration: got %v", err)
	}
	_, err = ParseRecentLogsBounds(domain.ManagerSystemd, "nginx.service", 10, "169h")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("window too large: got %v", err)
	}
}

func TestParseRecentLogsBounds_unknownManager(t *testing.T) {
	t.Parallel()
	_, err := ParseRecentLogsBounds(domain.ServiceManager("weird"), "x", 10, "")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("got %v", err)
	}
}
