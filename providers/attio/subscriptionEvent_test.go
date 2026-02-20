package attio

import (
	"errors"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
)

func TestGetFieldNameFromObjectMetadata(t *testing.T) {
	t.Parallel()

	metadata := &common.ListObjectMetadataResult{
		Result: map[string]common.ObjectMetadata{
			"obj-uuid-1": {
				DisplayName: "Companies",
				Fields: map[string]common.FieldMetadata{
					"name": {
						DisplayName: "name",
						FieldId:     goutils.Pointer("attr-uuid-1"),
					},
					"domains": {
						DisplayName: "domains",
						FieldId:     goutils.Pointer("attr-uuid-2"),
					},
				},
			},
			"obj-uuid-2": {
				DisplayName: "People",
				Fields: map[string]common.FieldMetadata{
					"email": {
						DisplayName: "email",
						FieldId:     nil,
					},
				},
			},
		},
		Errors: map[string]error{},
	}

	tests := []struct {
		name        string
		objectID    string
		attributeID string
		expected    string
		expectedErr error
	}{
		{
			name:        "Found field by attribute ID",
			objectID:    "obj-uuid-1",
			attributeID: "attr-uuid-1",
			expected:    "name",
		},
		{
			name:        "Found second field by attribute ID",
			objectID:    "obj-uuid-1",
			attributeID: "attr-uuid-2",
			expected:    "domains",
		},
		{
			name:        "Object not found",
			objectID:    "unknown-obj",
			attributeID: "attr-uuid-1",
			expectedErr: common.ErrNotFound,
		},
		{
			name:        "Attribute not found in object",
			objectID:    "obj-uuid-1",
			attributeID: "unknown-attr",
			expectedErr: common.ErrNotFound,
		},
		{
			name:        "Nil FieldId is skipped",
			objectID:    "obj-uuid-2",
			attributeID: "any-attr",
			expectedErr: common.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := GetFieldNameFromObjectMetadata(metadata, tt.objectID, tt.attributeID)
			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}

				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetObjectNameFromObjectMetadata(t *testing.T) {
	t.Parallel()

	metadata := &common.ListObjectMetadataResult{
		Result: map[string]common.ObjectMetadata{
			"obj-uuid-1": {
				DisplayName: "Companies",
			},
			"obj-uuid-2": {
				DisplayName: "People",
			},
		},
		Errors: map[string]error{},
	}

	tests := []struct {
		name        string
		objectID    string
		expected    string
		expectedErr error
	}{
		{
			name:     "Found object display name",
			objectID: "obj-uuid-1",
			expected: "Companies",
		},
		{
			name:     "Found second object display name",
			objectID: "obj-uuid-2",
			expected: "People",
		},
		{
			name:        "Object not found",
			objectID:    "unknown-obj",
			expectedErr: common.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := GetObjectNameFromObjectMetadata(metadata, tt.objectID)
			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}

				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
