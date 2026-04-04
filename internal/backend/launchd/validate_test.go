package launchd

import "testing"

func TestValidateLabel(t *testing.T) {
	t.Parallel()
	if err := validateLabel("com.apple.Safari"); err != nil {
		t.Fatal(err)
	}
	if err := validateLabel("bad label"); err == nil {
		t.Fatal("expected error")
	}
}

func TestPrintTargetParam(t *testing.T) {
	t.Parallel()
	r := PrintTargetParam("target")
	if err := r.Validate("gui/501/com.foo"); err != nil {
		t.Fatal(err)
	}
	if err := r.Validate("system/com.apple.sshd"); err != nil {
		t.Fatal(err)
	}
	if err := r.Validate("../../../etc/passwd"); err == nil {
		t.Fatal("expected error")
	}
}
