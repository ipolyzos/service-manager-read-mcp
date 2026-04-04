package openrc

import "testing"

func TestValidateServiceName(t *testing.T) {
	t.Parallel()
	if err := validateServiceName("nginx"); err != nil {
		t.Fatal(err)
	}
	if err := validateServiceName("net.eth0"); err != nil {
		t.Fatal(err)
	}
	if err := validateServiceName("../etc/passwd"); err == nil {
		t.Fatal("expected error")
	}
}
