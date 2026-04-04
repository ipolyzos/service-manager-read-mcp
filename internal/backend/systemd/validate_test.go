package systemd

import "testing"

func TestValidateUnitName(t *testing.T) {
	t.Parallel()
	if err := validateUnitName("nginx.service"); err != nil {
		t.Fatal(err)
	}
	if err := validateUnitName("user@1000.service"); err != nil {
		t.Fatal(err)
	}
	if err := validateUnitName(""); err == nil {
		t.Fatal("expected error")
	}
	if err := validateUnitName("bad space.service"); err == nil {
		t.Fatal("expected error")
	}
}

func TestJournalLineCountParam(t *testing.T) {
	t.Parallel()
	r := JournalLineCountParam("n", 100)
	if err := r.Validate("0"); err == nil {
		t.Fatal("expected error")
	}
	if err := r.Validate("101"); err == nil {
		t.Fatal("expected error")
	}
	if err := r.Validate("50"); err != nil {
		t.Fatal(err)
	}
}
