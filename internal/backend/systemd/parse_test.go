package systemd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestParseListUnitsJSON_fixture(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "list_units.json")
	units, err := ParseListUnitsJSON(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 2 {
		t.Fatalf("len=%d", len(units))
	}
	if units[0].Ref.ID != "nginx.service" || units[0].State != domain.StateRunning {
		t.Fatalf("first: %#v", units[0])
	}
	if units[1].State != domain.StateStopped {
		t.Fatalf("second state: %v", units[1].State)
	}
}

func TestParseFailedServicesJSON_fixture(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "list_failed.json")
	failed, err := ParseFailedServicesJSON(b)
	if err != nil || len(failed) != 1 {
		t.Fatalf("err=%v len=%d", err, len(failed))
	}
	if failed[0].Ref.ID != "broken.service" || failed[0].Reason != "failed" {
		t.Fatalf("%#v", failed[0])
	}
}

func TestParseShowProperties_fixture(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "show_nginx.txt")
	m, err := ParseShowProperties(b)
	if err != nil {
		t.Fatal(err)
	}
	ref := domain.UnitRef{Manager: domain.ManagerSystemd, ID: "nginx.service"}
	st, err := StatusFromShow(ref, m)
	if err != nil {
		t.Fatal(err)
	}
	if st.Active != domain.ActiveStateActive || st.Load != domain.LoadStateLoaded {
		t.Fatalf("%#v", st)
	}
	if st.MainPID == nil || *st.MainPID != 1234 {
		t.Fatalf("pid: %v", st.MainPID)
	}
	if st.Enabled == nil || !*st.Enabled {
		t.Fatalf("enabled: %v", st.Enabled)
	}
}

func TestStatusFromShow_notFound(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "show_notfound.txt")
	m, err := ParseShowProperties(b)
	if err != nil {
		t.Fatal(err)
	}
	ref := domain.UnitRef{Manager: domain.ManagerSystemd, ID: "missing.service"}
	_, err = StatusFromShow(ref, m)
	if !errors.Is(err, svcerr.ErrUnitNotFound) {
		t.Fatalf("expected ErrUnitNotFound, got %v", err)
	}
}

func TestParseJournalShortISO_noEntriesSentinel(t *testing.T) {
	t.Parallel()
	ref := domain.UnitRef{Manager: domain.ManagerSystemd, ID: "sshd.service"}
	lines, err := ParseJournalShortISO([]byte("-- No entries --\n"), ref)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 0 {
		t.Fatalf("expected no lines, got %d", len(lines))
	}
}

func TestParseJournalShortISO_fixture(t *testing.T) {
	t.Parallel()
	b := mustRead(t, "journal_short_iso.txt")
	ref := domain.UnitRef{Manager: domain.ManagerSystemd, ID: "nginx.service"}
	lines, err := ParseJournalShortISO(b, ref)
	if err != nil || len(lines) != 3 {
		t.Fatalf("err=%v n=%d", err, len(lines))
	}
	if lines[0].Line != "nginx.service[123]: ready" {
		t.Fatalf("line0: %q", lines[0].Line)
	}
	if lines[0].Timestamp == nil {
		t.Fatal("missing ts")
	}
	if lines[2].Timestamp != nil {
		t.Fatalf("line2 ts should be nil for -- placeholder, got %v", lines[2].Timestamp)
	}
	if lines[2].Line != "sshd[999]: no journal timestamp placeholder" {
		t.Fatalf("line2: %q", lines[2].Line)
	}
}

func TestParseListUnitsJSON_invalid(t *testing.T) {
	t.Parallel()
	_, err := ParseListUnitsJSON([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func mustRead(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join("testdata", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
