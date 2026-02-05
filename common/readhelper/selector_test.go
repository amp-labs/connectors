package readhelper

import (
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/internal/datautils"
)

func TestSelectFields(t *testing.T) {
	tests := []struct {
		name     string
		record   map[string]any
		fields   datautils.StringSet
		expected map[string]any
	}{
		{
			name:     "empty record and empty fields",
			record:   map[string]any{},
			fields:   datautils.NewStringSet(),
			expected: map[string]any{},
		},
		{
			name:   "top-level fields",
			record: map[string]any{"id": 123, "threadId": "abc", "payload": "ignored"},
			fields: datautils.NewStringSet("id", "threadId"),
			expected: map[string]any{
				"id":       123,
				"threadid": "abc",
			},
		},
		{
			name: "nested fields using JSONPath",
			record: map[string]any{
				"id": "msg1",
				"payload": map[string]any{
					"body": map[string]any{
						"data": "hello",
						"size": 100,
					},
					"mimeType": "text/plain",
					"labels":   []string{"a"},
				},
			},
			fields: datautils.NewStringSet("$['payload']['body']['data']", "$['payload']['mimeType']"),
			expected: map[string]any{
				"payload": map[string]any{
					"body": map[string]any{
						"data": "hello",
					},
					"mimetype": "text/plain",
				},
			},
		},
		{
			name: "path is not JSONPath style but plain",
			record: map[string]any{
				"id":          "msg1",
				"threadId":    "abc",
				"description": "amazing",
			},
			fields: datautils.NewStringSet("id", "threadId"),
			expected: map[string]any{
				"id":       "msg1",
				"threadid": "abc",
			},
		},
		{
			name: "multiple nested paths share parent",
			record: map[string]any{
				"payload": map[string]any{
					"body": map[string]any{
						"data": "abc",
						"size": 123,
					},
					"mimeType": "text/html",
				},
			},
			fields: datautils.NewStringSet("$['payload']['body']", "$['payload']['body']['data']"),
			expected: map[string]any{
				"payload": map[string]any{
					"body": map[string]any{
						"data": "abc",
						"size": 123,
					},
				},
			},
		},
		{
			name: "top-level non-existent field",
			record: map[string]any{
				"id": "123",
			},
			fields:   datautils.NewStringSet("threadId"),
			expected: map[string]any{},
		},
		{
			name: "nested non-existent field",
			record: map[string]any{
				"payload": map[string]any{
					"body": map[string]any{"size": 42},
				},
			},
			fields:   datautils.NewStringSet("$['payload']['body']['data']"),
			expected: map[string]any{},
		},
		{
			name: "complex mixed case",
			record: map[string]any{
				"id": "msg1",
				"payload": map[string]any{
					"body": map[string]any{"data": "xyz", "size": 77},
					"headers": map[string]any{
						"From": "alice",
					},
				},
			},
			fields: datautils.NewStringSet("id", "$['payload']['body']['data']", "$['payload']['headers']['From']"),
			expected: map[string]any{
				"id": "msg1",
				"payload": map[string]any{
					"body": map[string]any{
						"data": "xyz",
					},
					"headers": map[string]any{
						"from": "alice",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectFields(tt.record, tt.fields)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("SelectFields() = %#v, want %#v", got, tt.expected)
			}
		})
	}
}
