package runit

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// ResolveServiceDir returns the runit scan directory (service names are immediate children).
// Order: non-empty SVDIR env, then /var/service, /service, /etc/service (first existing directory).
func ResolveServiceDir(getenv func(string) string) (string, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	if d := strings.TrimSpace(getenv("SVDIR")); d != "" {
		clean := filepath.Clean(d)
		if fi, err := os.Stat(clean); err == nil && fi.IsDir() {
			return clean, nil
		}
	}
	for _, candidate := range []string{"/var/service", "/service", "/etc/service"} {
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%w: no runit service directory (set SVDIR or install under /var/service, /service, /etc/service)", svcerr.ErrBackendUnavailable)
}

// ListServiceNames returns sorted, validated child names of dir (non-hidden entries only).
func ListServiceNames(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("%w: read service dir: %w", svcerr.ErrBackendUnavailable, err)
	}
	var names []string
	for _, e := range ents {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if err := validateServiceName(name); err != nil {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}
