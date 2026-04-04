package brew

import "testing"

func TestValidateFormulaName(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in string
		ok bool
	}{
		{"nginx", true},
		{"docker-machine", true},
		{"python@3.12", true},
		{"", false},
		{"../foo", false},
		{"-x", false},
	} {
		err := validateFormulaName(tc.in)
		if tc.ok && err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%q: want error", tc.in)
		}
	}
}
