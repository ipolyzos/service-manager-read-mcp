package launchd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestParseLaunchctlList_fixture(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "list_sample.txt")
	rows, err := ParseLaunchctlList(b)
	if err != nil || len(rows) != 3 {
		t.Fatalf("err=%v rows=%d", err, len(rows))
	}
	if rows[2].Label != "com.example.broken" || rows[2].LastStatus != 78 {
		t.Fatalf("%#v", rows[2])
	}
	units := ListRowsToUnits(rows)
	if len(units) != 3 {
		t.Fatalf("units %d", len(units))
	}
	if units[0].State != domain.StateRunning || units[1].State != domain.StateStopped {
		t.Fatalf("%#v %#v", units[0], units[1])
	}
	failed := ListRowsToFailedApproximation(rows)
	if len(failed) != 1 || failed[0].Ref.ID != "com.example.broken" {
		t.Fatalf("%#v", failed)
	}
}

func TestParseLaunchctlPrint_running(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "print_running.txt")
	ref := domain.UnitRef{Manager: domain.ManagerLaunchd, ID: "com.example.app"}
	st, err := ParseLaunchctlPrint(ref, b)
	if err != nil {
		t.Fatal(err)
	}
	if st.Active != domain.ActiveStateActive || st.MainPID == nil || *st.MainPID != 4242 {
		t.Fatalf("%#v", st)
	}
}

func TestParseLaunchctlPrint_missing(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "print_missing.txt")
	ref := domain.UnitRef{Manager: domain.ManagerLaunchd, ID: "com.missing"}
	_, err := ParseLaunchctlPrint(ref, b)
	if !errors.Is(err, svcerr.ErrUnitNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestParseLaunchctlList_badLine(t *testing.T) {
	t.Parallel()
	_, err := ParseLaunchctlList([]byte("not enough columns here\n"))
	if err == nil {
		t.Fatal("expected error")
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
