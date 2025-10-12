package naming

import (
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic conversions
		{name: "PascalCase to snake_case", input: "UserId", expected: "user_id"},
		{name: "camelCase to snake_case", input: "firstName", expected: "first_name"},
		{name: "Single word", input: "email", expected: "email"},
		{name: "Already snake_case", input: "user_id", expected: "user_id"},
		{name: "Empty string", input: "", expected: ""},

		// Complex cases
		{name: "Multiple words", input: "EmailAddress", expected: "email_address"},
		{name: "Acronym handling", input: "HTTPResponse", expected: "http_response"},
		{name: "Acronym at end", input: "UserID", expected: "user_id"},
		{name: "Mixed case", input: "XMLHttpRequest", expected: "xml_http_request"},

		// Provider-specific examples
		{name: "Attio field", input: "ContentPlaintext", expected: "content_plaintext"},
		{name: "Amplitude field", input: "SessionID", expected: "session_id"},
		{name: "Apollo field", input: "AccountStageId", expected: "account_stage_id"},
		{name: "Avoma field", input: "ExternalId", expected: "external_id"},
		{name: "Bitbucket field", input: "CreatedOn", expected: "created_on"},

		// Edge cases
		{name: "Leading uppercase", input: "ID", expected: "id"},
		{name: "Trailing uppercase", input: "createHTML", expected: "create_html"},
		{name: "Numbers", input: "field123Name", expected: "field123_name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ToSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToSnakeCase(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic conversions
		{name: "snake_case to camelCase", input: "first_name", expected: "firstName"},
		{name: "kebab-case to camelCase", input: "is-archived", expected: "isArchived"},
		{name: "PascalCase to camelCase", input: "FirstName", expected: "firstName"},
		{name: "Already camelCase", input: "firstName", expected: "firstName"},
		{name: "Empty string", input: "", expected: ""},

		// Complex cases
		{name: "Multiple words snake", input: "user_id_value", expected: "userIdValue"},
		{name: "Multiple words kebab", input: "email-address-field", expected: "emailAddressField"},
		{name: "Space separated", input: "first name", expected: "firstName"},

		// Provider-specific examples (Ashby uses camelCase)
		{name: "Ashby field", input: "is_archived", expected: "isArchived"},
		{name: "Ashby compound", input: "primary_email_address", expected: "primaryEmailAddress"},

		// Edge cases
		{name: "Single word", input: "name", expected: "name"},
		{name: "Single uppercase", input: "Name", expected: "name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ToCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToCamelCase(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic conversions
		{name: "snake_case to PascalCase", input: "user_id", expected: "UserId"},
		{name: "kebab-case to PascalCase", input: "email-address", expected: "EmailAddress"},
		{name: "camelCase to PascalCase", input: "firstName", expected: "FirstName"},
		{name: "lowercase to PascalCase", input: "email", expected: "Email"},
		{name: "Empty string", input: "", expected: ""},

		// Complex cases
		{name: "Multiple words", input: "user_id_value", expected: "UserIdValue"},
		{name: "Already PascalCase", input: "FirstName", expected: "FirstName"},

		// Provider-specific examples (AWS, Salesforce use PascalCase)
		{name: "AWS field", input: "display_name", expected: "DisplayName"},
		{name: "Salesforce object", input: "account", expected: "Account"},

		// Edge cases
		{name: "Single letter", input: "a", expected: "A"},
		{name: "Mixed delimiters", input: "first_name-last", expected: "FirstNameLast"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ToPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascalCase(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCapitalizeFirst(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "Lowercase word", input: "hello", expected: "Hello"},
		{name: "Already capitalized", input: "Hello", expected: "Hello"},
		{name: "All caps", input: "HELLO", expected: "Hello"},
		{name: "Empty string", input: "", expected: ""},
		{name: "Single char", input: "a", expected: "A"},
		{name: "Mixed case", input: "hELLO", expected: "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := capitalizeFirst(tt.input)
			if result != tt.expected {
				t.Errorf("capitalizeFirst(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCapitalizeFirstOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "Lowercase word", input: "hello", expected: "Hello"},
		{name: "Already capitalized", input: "Hello", expected: "Hello"},
		{name: "All caps preserves", input: "HELLO", expected: "HELLO"},
		{name: "Empty string", input: "", expected: ""},
		{name: "Mixed case preserved", input: "hELLO", expected: "HELLO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CapitalizeFirstOnly(tt.input)
			if result != tt.expected {
				t.Errorf("CapitalizeFirstOnly(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsSnakeCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "Valid snake_case", input: "user_id", expected: true},
		{name: "Valid single word", input: "email", expected: true},
		{name: "PascalCase", input: "UserId", expected: false},
		{name: "camelCase", input: "userId", expected: false},
		{name: "Empty string", input: "", expected: true},
		{name: "With numbers", input: "field_123", expected: true},
		{name: "With uppercase", input: "User_Id", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("isSnakeCase(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSplitIntoWords(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{name: "snake_case", input: "first_name", expected: []string{"first", "name"}},
		{name: "kebab-case", input: "first-name", expected: []string{"first", "name"}},
		{name: "space separated", input: "first name", expected: []string{"first", "name"}},
		{name: "PascalCase", input: "FirstName", expected: []string{"First", "Name"}},
		{name: "camelCase", input: "firstName", expected: []string{"first", "Name"}},
		{name: "Empty string", input: "", expected: nil},
		{name: "Single word", input: "name", expected: []string{"name"}},
		{name: "Multiple delimiters", input: "first_name-last name", expected: []string{"first", "name", "last", "name"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := splitIntoWords(testCase.input)
			if len(result) != len(testCase.expected) {
				t.Errorf("splitIntoWords(%q) length = %d; want %d", testCase.input, len(result), len(testCase.expected))

				return
			}

			for i := range result {
				if result[i] != testCase.expected[i] {
					t.Errorf("splitIntoWords(%q)[%d] = %q; want %q", testCase.input, i, result[i], testCase.expected[i])
				}
			}
		})
	}
}

func TestIsUpper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    rune
		expected bool
	}{
		{name: "Uppercase A", input: 'A', expected: true},
		{name: "Uppercase Z", input: 'Z', expected: true},
		{name: "Lowercase a", input: 'a', expected: false},
		{name: "Lowercase z", input: 'z', expected: false},
		{name: "Number", input: '1', expected: false},
		{name: "Symbol", input: '_', expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsUpper(tt.input)
			if result != tt.expected {
				t.Errorf("IsUpper(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsLower(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    rune
		expected bool
	}{
		{name: "Lowercase a", input: 'a', expected: true},
		{name: "Lowercase z", input: 'z', expected: true},
		{name: "Uppercase A", input: 'A', expected: false},
		{name: "Uppercase Z", input: 'Z', expected: false},
		{name: "Number", input: '1', expected: false},
		{name: "Symbol", input: '_', expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsLower(tt.input)
			if result != tt.expected {
				t.Errorf("IsLower(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToLowerRune(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    rune
		expected rune
	}{
		{name: "Uppercase to lowercase", input: 'A', expected: 'a'},
		{name: "Already lowercase", input: 'a', expected: 'a'},
		{name: "Number unchanged", input: '1', expected: '1'},
		{name: "Symbol unchanged", input: '_', expected: '_'},
		{name: "Z to z", input: 'Z', expected: 'z'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ToLowerRune(tt.input)
			if result != tt.expected {
				t.Errorf("ToLowerRune(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToUpperRune(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    rune
		expected rune
	}{
		{name: "Lowercase to uppercase", input: 'a', expected: 'A'},
		{name: "Already uppercase", input: 'A', expected: 'A'},
		{name: "Number unchanged", input: '1', expected: '1'},
		{name: "Symbol unchanged", input: '_', expected: '_'},
		{name: "z to Z", input: 'z', expected: 'Z'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ToUpperRune(tt.input)
			if result != tt.expected {
				t.Errorf("ToUpperRune(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
