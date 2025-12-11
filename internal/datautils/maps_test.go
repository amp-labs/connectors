package datautils

import (
	"encoding/gob"
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestMapDeepCopy(t *testing.T) { // nolint:funlen
	t.Parallel()

	type Sample struct {
		ID   int
		Name string
	}

	// Must register with gob types that would be used with deep copy.
	gob.Register(Sample{})

	tests := []struct {
		name        string
		input       Map[string, any]
		modifyCopy  func(Map[string, any])
		expected    Map[string, any]
		expectError error
	}{
		{
			name:        "Empty map",
			input:       Map[string, any]{},
			modifyCopy:  nil,
			expected:    Map[string, any]{},
			expectError: nil,
		},
		{
			name: "Simple primitive values",
			input: Map[string, any]{
				"a": 1,
				"b": "text",
			},
			modifyCopy: func(c Map[string, any]) {
				c["a"] = 99 // should not affect original
			},
			expected: Map[string, any]{
				"a": 1,
				"b": "text",
			},
			expectError: nil,
		},
		{
			name: "Nested map structure",
			input: Map[string, any]{
				"config": Map[string, any]{
					"enabled": true,
					"retries": 3,
				},
			},
			modifyCopy: func(c Map[string, any]) {
				nested := c["config"].(Map[string, any]) // nolint:forcetypeassert
				nested["retries"] = 10                   // change only copy
			},
			expected: Map[string, any]{
				"config": Map[string, any]{
					"enabled": true,
					"retries": 3,
				},
			},
			expectError: nil,
		},
		{
			name: "Complex object values",
			input: Map[string, any]{
				"user": Sample{ID: 1, Name: "Alice"},
			},
			modifyCopy: func(c Map[string, any]) {
				user := c["user"].(Sample) // nolint:forcetypeassert
				user.Name = "Bob"          // should not affect original
				c["user"] = user
			},
			expected: Map[string, any]{
				"user": Sample{ID: 1, Name: "Alice"},
			},
			expectError: nil,
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			replica, err := tt.input.DeepCopy()

			testutils.CheckOutputWithError(t, tt.name, tt.expected, tt.expectError, replica, err)

			// If no error and we have a modifyCopy function, verify deep copy isolation
			if err == nil && tt.modifyCopy != nil {
				tt.modifyCopy(replica)

				// Recheck that original is unchanged after mutation of copy
				testutils.CheckOutputWithError(t, tt.name+" (post-mutation)", tt.expected, nil, tt.input, nil)
			}
		})
	}
}
