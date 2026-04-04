package backend_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/backend/stub"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestStubBackend_detectAndList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ref := domain.UnitRef{Manager: domain.ManagerSystemd, ID: "nginx.service"}

	t.Run("unavailable", func(t *testing.T) {
		t.Parallel()
		b := &stub.Backend{Manager: domain.ManagerSystemd, Available: false}
		if _, err := b.DetectServiceManager(ctx); !errors.Is(err, svcerr.ErrBackendUnavailable) {
			t.Fatalf("detect: %v", err)
		}
		if _, err := b.ListUnits(ctx, domain.ListUnitsInput{}); !errors.Is(err, svcerr.ErrBackendUnavailable) {
			t.Fatalf("list: %v", err)
		}
		if _, err := b.GetServiceStatus(ctx, ref); !errors.Is(err, svcerr.ErrBackendUnavailable) {
			t.Fatalf("status: %v", err)
		}
	})

	t.Run("available", func(t *testing.T) {
		t.Parallel()
		b := &stub.Backend{Manager: domain.ManagerSystemd, Available: true}
		m, err := b.DetectServiceManager(ctx)
		if err != nil || m != domain.ManagerSystemd {
			t.Fatalf("detect: %v %v", m, err)
		}
		units, err := b.ListUnits(ctx, domain.ListUnitsInput{Limit: 10})
		if err != nil || len(units) != 0 {
			t.Fatalf("list: %v %#v", err, units)
		}
		if _, err := b.GetServiceStatus(ctx, ref); !errors.Is(err, svcerr.ErrUnsupported) {
			t.Fatalf("status: %v", err)
		}
		if _, err := b.ListFailedServices(ctx); !errors.Is(err, svcerr.ErrUnsupported) {
			t.Fatalf("failed: %v", err)
		}
		if _, err := b.GetRecentServiceLogs(ctx, ref, domain.RecentLogsInput{}); !errors.Is(err, svcerr.ErrUnsupported) {
			t.Fatalf("logs: %v", err)
		}
		caps := b.Capabilities()
		if caps.Manager != domain.ManagerSystemd || !caps.SupportsListUnits {
			t.Fatalf("caps: %#v", caps)
		}
	})
}
