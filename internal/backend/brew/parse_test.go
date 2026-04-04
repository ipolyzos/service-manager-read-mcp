package brew

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestParseBrewServicesList(t *testing.T) {
	t.Parallel()
	raw := mustReadFile(t, "brew_services_list.json")
	units, err := ParseBrewServicesList(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 5 {
		t.Fatalf("len=%d", len(units))
	}
	byID := map[string]domain.Unit{}
	for _, u := range units {
		byID[u.Ref.ID] = u
	}
	if byID["nginx"].State != domain.StateRunning {
		t.Fatalf("nginx %#v", byID["nginx"])
	}
	if byID["redis"].State != domain.StateStopped {
		t.Fatalf("redis %#v", byID["redis"])
	}
	if byID["badsvc"].State != domain.StateFailed {
		t.Fatalf("badsvc %#v", byID["badsvc"])
	}
	if byID["weird"].State != domain.StateFailed {
		t.Fatalf("weird %#v (non-zero exit_code should win)", byID["weird"])
	}
	if byID["old"].State != domain.StateFailed {
		t.Fatalf("old %#v", byID["old"])
	}
}

func TestParseBrewServicesListFailed(t *testing.T) {
	t.Parallel()
	raw := mustReadFile(t, "brew_services_list.json")
	f, err := ParseBrewServicesListFailed(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(f) != 3 {
		t.Fatalf("want 3 failed rows (error, exit 1, exit 2), got %#v", f)
	}
}

func TestParseBrewServicesInfo_running(t *testing.T) {
	t.Parallel()
	raw := mustReadFile(t, "brew_services_info_nginx.json")
	st, err := ParseBrewServicesInfo(raw, "nginx")
	if err != nil {
		t.Fatal(err)
	}
	if st.Ref.ID != "nginx" || st.Active != domain.ActiveStateActive {
		t.Fatalf("%#v", st)
	}
	if st.MainPID == nil || *st.MainPID != 4242 {
		t.Fatalf("pid=%v", st.MainPID)
	}
	if st.Load != domain.LoadStateLoaded {
		t.Fatalf("load=%v", st.Load)
	}
}

func TestParseBrewServicesInfo_stopped(t *testing.T) {
	t.Parallel()
	raw := mustReadFile(t, "brew_services_info_dbus.json")
	st, err := ParseBrewServicesInfo(raw, "dbus")
	if err != nil {
		t.Fatal(err)
	}
	if st.Active != domain.ActiveStateInactive || st.SubState != "none" {
		t.Fatalf("%#v", st)
	}
	if st.Load != domain.LoadStateUnknown {
		t.Fatalf("load=%v", st.Load)
	}
}

func TestParseBrewServicesInfo_empty(t *testing.T) {
	t.Parallel()
	_, err := ParseBrewServicesInfo([]byte("[]"), "x")
	if !errors.Is(err, svcerr.ErrUnitNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestFormulaMissingInBrewServicesInfo(t *testing.T) {
	t.Parallel()
	if !FormulaMissingInBrewServicesInfo([]byte("Error: No available formula with the name \"x\".\n"), nil) {
		t.Fatal("expected match")
	}
	if FormulaMissingInBrewServicesInfo([]byte("other"), nil) {
		t.Fatal("expected no match")
	}
}

func mustReadFile(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}
