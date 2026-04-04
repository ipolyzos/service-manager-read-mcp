package runit

import "testing"

func TestValidateServiceName(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in string
		ok bool
	}{
		{"nginx", true},
		{"foo-bar", true},
		{"a1", true},
		{"", false},
		{"../etc", false},
		{"-bad", false},
	} {
		err := validateServiceName(tc.in)
		if tc.ok && err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%q: want error", tc.in)
		}
	}
}
