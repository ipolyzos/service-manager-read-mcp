package servicemcp

import (
	"errors"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestValidateServiceIdentifier(t *testing.T) {
	t.Parallel()
	if err := validateServiceIdentifier(domain.ManagerSystemd, "nginx.service"); err != nil {
		t.Fatal(err)
	}
	if err := validateServiceIdentifier(domain.ManagerSystemd, ""); !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("empty id: got %v", err)
	}
	if err := validateServiceIdentifier(domain.ManagerSystemd, "no spaces.service"); err == nil {
		t.Fatal("expected error for invalid unit name")
	} else if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("got %v", err)
	}
	if err := validateServiceIdentifier(domain.ServiceManager("bogus"), "x"); !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("bogus manager: got %v", err)
	}
}
