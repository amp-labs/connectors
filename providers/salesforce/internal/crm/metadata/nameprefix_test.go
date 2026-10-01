package metadata

import (
	"strings"
	"testing"
)

func TestSanitizeNamePrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Already clean", "acme", "acme"},
		{"Spaces and hyphens become underscores", "Acme Sales-Tools", "Acme_Sales_Tools"},
		{"Dots become underscores", "acme.io", "acme_io"},
		{"Illegal characters are dropped", "Acme! (EU) #1", "Acme_EU_1"},
		{"Consecutive underscores collapse", "acme___sales", "acme_sales"},
		{"Trailing underscores trimmed", "acme__", "acme"},
		{"Digit-leading name is prefixed, not stripped", "11x", "proj_11x"},
		{"Leading digits kept under proj_", "7shifts", "proj_7shifts"},
		{"Digits-only name is prefixed", "123", "proj_123"},
		{"Leading underscore dropped", "_acme", "acme"},
		{"Digits kept when not leading", "acme2", "acme2"},
		{"Punctuation only yields empty", "!!!", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := SanitizeNamePrefix(tt.input); got != tt.expected {
				t.Errorf("SanitizeNamePrefix(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGenerateSubscriptionArtifactName(t *testing.T) {
	t.Parallel()

	longPrefix := strings.Repeat("p", developerNameMaxLength)

	tests := []struct {
		name     string
		prefix   string
		object   string
		suffix   string
		expected string
	}{
		{"Create pair", "acme", "Account", ArtifactSuffixCreate, "acme_Account_Create"},
		{"Update pair", "acme", "Account", ArtifactSuffixUpdate, "acme_Account_Update"},
		{"Custom object collapses before the suffix", "acme", "My_Object__c", ArtifactSuffixCreate,
			"acme_My_Object_c_Create"},
		{"Digit-leading project keeps its digits", "11x", "Lead", "", "proj_11x_Lead"},
		{"Empty prefix falls back to default", "", "Account", "", "amp_Account"},
		{"Prefix truncates to leave room for the suffix", longPrefix, "Account", ArtifactSuffixUpdate,
			longPrefix[:developerNameMaxLength-len("_Account_Update")] + "_Account_Update"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := GenerateSubscriptionArtifactName(tt.prefix, tt.object, tt.suffix)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("GenerateSubscriptionArtifactName(%q, %q, %q) = %q, want %q",
					tt.prefix, tt.object, tt.suffix, got, tt.expected)
			}

			if len(got) > developerNameMaxLength {
				t.Errorf("len = %d, want at most %d", len(got), developerNameMaxLength)
			}
		})
	}
}
