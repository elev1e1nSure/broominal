package style

import (
	"strings"
	"testing"
)

func TestSprintf(t *testing.T) {
	got := Sprintf(Bold, "hello %s", "world")
	if !strings.Contains(got, "hello world") {
		t.Errorf("Sprintf output missing content: %q", got)
	}
	if !strings.Contains(got, Bold) {
		t.Errorf("Sprintf output missing color code: %q", got)
	}
	if !strings.Contains(got, Reset) {
		t.Errorf("Sprintf output missing reset code: %q", got)
	}
}

func TestColorFuncs(t *testing.T) {
	cases := []struct {
		name string
		fn   func(string, ...any) string
		want string
	}{
		{"Boldf", Boldf, Bold},
		{"Greenf", Greenf, Green},
		{"Yellowf", Yellowf, Yellow},
		{"Redf", Redf, Red},
		{"Cyanf", Cyanf, Cyan},
		{"Grayf", Grayf, Gray},
		{"Passf", Passf, Green + Bold},
		{"Warnf", Warnf, Yellow + Bold},
		{"Failf", Failf, Red + Bold},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn("test")
			if !strings.HasPrefix(got, tc.want) {
				t.Errorf("%s = %q, expected prefix %q", tc.name, got, tc.want)
			}
			if !strings.HasSuffix(got, Reset) {
				t.Errorf("%s missing Reset suffix: %q", tc.name, got)
			}
		})
	}
}
