package runit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestParseSvStatus_run(t *testing.T) {
	t.Parallel()
	raw := mustRead(t, "sv_status_run.txt")
	ref := domain.UnitRef{Manager: domain.ManagerRunit, ID: "sshd"}
	st, err := ParseSvStatus(ref, raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if st.Active != domain.ActiveStateActive || st.SubState != "run" {
		t.Fatalf("%#v", st)
	}
	if st.MainPID == nil || *st.MainPID != 1234 {
		t.Fatalf("pid=%v", st.MainPID)
	}
}

func TestParseSvStatus_downNormal(t *testing.T) {
	t.Parallel()
	raw := mustRead(t, "sv_status_down_normal.txt")
	ref := domain.UnitRef{Manager: domain.ManagerRunit, ID: "cron"}
	st, err := ParseSvStatus(ref, raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if st.Active != domain.ActiveStateInactive || st.SubState != "normally_down" {
		t.Fatalf("%#v", st)
	}
}

func TestParseSvStatus_wantUp(t *testing.T) {
	t.Parallel()
	raw := mustRead(t, "sv_status_want_up.txt")
	ref := domain.UnitRef{Manager: domain.ManagerRunit, ID: "nginx"}
	st, err := ParseSvStatus(ref, raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if st.Active != domain.ActiveStateFailed || st.SubState != "want_up" {
		t.Fatalf("%#v", st)
	}
}

func TestParseSvStatus_missing(t *testing.T) {
	t.Parallel()
	raw := mustRead(t, "sv_status_missing.txt")
	ref := domain.UnitRef{Manager: domain.ManagerRunit, ID: "nginx"}
	_, err := ParseSvStatus(ref, raw, nil)
	if !errors.Is(err, svcerr.ErrUnitNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestSvStatusIndicatesWantUpDown(t *testing.T) {
	t.Parallel()
	raw := mustRead(t, "sv_status_want_up.txt")
	if !SvStatusIndicatesWantUpDown("nginx", raw, nil) {
		t.Fatal("expected true")
	}
	if SvStatusIndicatesWantUpDown("other", raw, nil) {
		t.Fatal("expected false for other service")
	}
}

func mustRead(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}
