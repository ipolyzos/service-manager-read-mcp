package openrc

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestParseRcServiceList_fixture(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "rc_service_list.txt")
	units, err := ParseRcServiceList(b)
	if err != nil || len(units) != 4 {
		t.Fatalf("err=%v n=%d", err, len(units))
	}
	if units[2].Ref.ID != "nginx" || units[2].State != domain.StateUnknown {
		t.Fatalf("%#v", units[2])
	}
}

func TestParseRcServiceList_empty(t *testing.T) {
	t.Parallel()
	u, err := ParseRcServiceList([]byte("   \n"))
	if err != nil || len(u) != 0 {
		t.Fatalf("err=%v n=%d", err, len(u))
	}
}

func TestParseRcServiceStatus_started(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "rc_service_status_started.txt")
	ref := domain.UnitRef{Manager: domain.ManagerOpenRC, ID: "nginx"}
	st, err := ParseRcServiceStatus(ref, b, nil)
	if err != nil || st.Active != domain.ActiveStateActive {
		t.Fatalf("err=%v %#v", err, st)
	}
}

func TestParseRcServiceStatus_stopped(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "rc_service_status_stopped.txt")
	ref := domain.UnitRef{Manager: domain.ManagerOpenRC, ID: "nginx"}
	st, err := ParseRcServiceStatus(ref, b, nil)
	if err != nil || st.Active != domain.ActiveStateInactive {
		t.Fatalf("err=%v %#v", err, st)
	}
}

func TestParseRcServiceStatus_missing(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "rc_service_status_missing.txt")
	ref := domain.UnitRef{Manager: domain.ManagerOpenRC, ID: "nosuch"}
	_, err := ParseRcServiceStatus(ref, b, nil)
	if !errors.Is(err, svcerr.ErrUnitNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestParseRcStatusCrashed_fixture(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "rc_status_crashed.txt")
	f, err := ParseRcStatusCrashed(b)
	if err != nil || len(f) != 2 {
		t.Fatalf("err=%v n=%d", err, len(f))
	}
	if f[0].Ref.ID != "broken" {
		t.Fatalf("%#v", f[0])
	}
}

func TestParseRcServiceList_garbage(t *testing.T) {
	t.Parallel()
	_, err := ParseRcServiceList([]byte("#\n#\n"))
	if !errors.Is(err, svcerr.ErrParse) {
		t.Fatalf("got %v", err)
	}
}

func mustRead(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join("testdata", name)
	out, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
