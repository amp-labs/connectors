package naming

import "testing"

func TestSeparateCamelCaseWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"webinarKey", "webinar Key"},
		{"numberOfRegistrants", "number Of Registrants"},
		{"inSession", "in Session"},
		{"sessionId", "session Id"},
		{"already separated", "already separated"},
		{"PascalCase", "Pascal Case"},
		{"URLPath", "URLPath"},
	}

	for _, tc := range tests {
		if got := SeparateCamelCaseWords(tc.in); got != tc.want {
			t.Errorf("SeparateCamelCaseWords(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
