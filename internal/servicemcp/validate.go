package servicemcp

import (
	"fmt"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/backend/brew"
	"github.com/ipolyzos/service-manager-read-mcp/internal/backend/launchd"
	"github.com/ipolyzos/service-manager-read-mcp/internal/backend/openrc"
	"github.com/ipolyzos/service-manager-read-mcp/internal/backend/runit"
	"github.com/ipolyzos/service-manager-read-mcp/internal/backend/systemd"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// validateServiceIdentifier enforces the same rules as each backend’s service/unit argv validators.
func validateServiceIdentifier(manager domain.ServiceManager, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("%w: id must not be empty", svcerr.ErrInvalidArgument)
	}
	var err error
	switch manager {
	case domain.ManagerSystemd:
		err = systemd.ValidateServiceQueryID(id)
	case domain.ManagerLaunchd:
		err = launchd.ValidateServiceQueryID(id)
	case domain.ManagerOpenRC:
		err = openrc.ValidateServiceQueryID(id)
	case domain.ManagerRunit:
		err = runit.ValidateServiceQueryID(id)
	case domain.ManagerBrew:
		err = brew.ValidateServiceQueryID(id)
	default:
		return fmt.Errorf("%w: unknown manager %q", svcerr.ErrInvalidArgument, manager)
	}
	if err != nil {
		return fmt.Errorf("%w: invalid service id: %v", svcerr.ErrInvalidArgument, err)
	}
	return nil
}
