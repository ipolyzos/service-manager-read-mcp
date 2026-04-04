package servicemcp

import (
	"errors"
	"strings"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestParseRequiredManager(t *testing.T) {
	t.Parallel()
	_, err := parseRequiredManager("")
	if !errors.Is(err, svcerr.ErrInvalidArgument) || !strings.Contains(err.Error(), "manager is required") {
		t.Fatalf("empty: got %v", err)
	}
	_, err = parseRequiredManager("   \t  ")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("whitespace: got %v", err)
	}
	got, err := parseRequiredManager(" systemd ")
	if err != nil || got != domain.ManagerSystemd {
		t.Fatalf("systemd: got %q err=%v", got, err)
	}
	_, err = parseRequiredManager("not-a-manager")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("unknown manager: got %v", err)
	}
	if strings.Contains(err.Error(), "manager is required") {
		t.Fatalf("unknown manager should not use empty-manager wording: %v", err)
	}
}
