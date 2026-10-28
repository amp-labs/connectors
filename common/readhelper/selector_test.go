package readhelper

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/google/go-cmp/cmp"
)

func TestSelectFields(t *testing.T) {
	tests := []struct {
		name     string
		record   string
		fields   datautils.StringSet
		expected string
	}{
		{
			name:     "empty record and empty fields",
			record:   `{}`,
			fields:   datautils.NewStringSet(),
			expected: `{}`,
		},
		{
			name:     "top-level fields",
			record:   `{"id": 123, "threadId": "abc", "payload": "ignored"}`,
			fields:   datautils.NewStringSet("id", "threadId"),
			expected: `{"id": 123, "threadid": "abc"}`,
		},
		{
			name: "nested fields using JSONPath",
			record: `{
				"id": "msg1",
				"payload": {
					"body": {
						"data": "hello",
						"size": 100
					},
					"mimeType": "text/plain",
					"labels": ["a"]
				}
			}`,
			fields: datautils.NewStringSet("$['payload']['body']['data']", "$['payload']['mimeType']"),
			expected: `{
				"payload": {
					"body": {
						"data": "hello"
					},
					"mimetype": "text/plain"
				}
			}`,
		},
		{
			name: "path is not JSONPath style but plain",
			record: `{
				"id": "msg1",
				"threadId": "abc",
				"description": "amazing"
			}`,
			fields: datautils.NewStringSet("id", "threadId"),
			expected: `{
				"id": "msg1",
				"threadid": "abc"
			}`,
		},
		{
			name: "multiple nested paths share parent",
			record: `{
				"payload": {
					"body": {
						"data": "abc",
						"size": 123
					},
					"mimeType": "text/html"
				}
			}`,
			fields: datautils.NewStringSet("$['payload']['body']", "$['payload']['body']['data']"),
			expected: `{
				"payload": {
					"body": {
						"data": "abc",
						"size": 123
					}
				}
			}`,
		},
		{
			name: "top-level non-existent field",
			record: `{
				"id": "123"
			}`,
			fields:   datautils.NewStringSet("threadId"),
			expected: `{}`,
		},
		{
			name: "nested non-existent field",
			record: `{
				"payload": {
					"body": {"size": 42}
				}
			}`,
			fields:   datautils.NewStringSet("$['payload']['body']['data']"),
			expected: `{}`,
		},
		{
			name: "complex mixed case",
			record: `{
				"id": "msg1",
				"payload": {
					"body": {"data": "xyz", "size": 77},
					"headers": {
						"From": "alice"
					}
				}
			}`,
			fields: datautils.NewStringSet("id", "$['payload']['body']['data']", "$['payload']['headers']['From']"),
			expected: `{
				"id": "msg1",
				"payload": {
					"body": {
						"data": "xyz"
					},
					"headers": {
						"from": "alice"
					}
				}
			}`,
		},
		{
			name: "nested arrays",
			record: `{
				"id": "msg1",
				"payload": {
					"body": {"data": "xyz", "size": 77},
					"headers": {
						"From": "alice"
					},
					"food": [
						{"id": "001", "value": "Fish"},
						{"id": "002", "value": "Chicken"}
					]
				}
			}`,
			fields: datautils.NewStringSet("$['payload']['food'][*]['value']"),
			expected: `{
				"payload": {
					"food": [
						{"value": "Fish"},
						{"value": "Chicken"}
					]
				}
			}`,
		},
		{
			name: "multiple nested array fields",
			record: `{
				"items": [
					{
						"id": 1,
						"name": "one",
						"tags": [
							{"key": "color", "val": "red"},
							{"key": "size", "val": "large"}
						]
					},
					{
						"id": 2,
						"name": "two",
						"tags": [
							{"key": "color", "val": "blue"}
						]
					}
				]
			}`,
			fields: datautils.NewStringSet("$['items'][*]['id']", "$['items'][*]['tags'][*]['key']"),
			expected: `{
				"items": [
					{
						"id": 1,
						"tags": [
							{"key": "color"},
							{"key": "size"}
						]
					},
					{
						"id": 2,
						"tags": [
							{"key": "color"}
						]
					}
				]
			}`,
		},
		{
			name: "Star can be used at the start and at the end of the path with no effect",
			record: `{
				"items": [
					{
						"id": 1,
						"name": "one",
						"tags": [
							{"key": "color", "val": "red"},
							{"key": "size", "val": "large"}
						]
					},
					{
						"id": 2,
						"name": "two",
						"tags": [
							{"key": "color", "val": "blue"}
						]
					}
				]
			}`,
			fields:   datautils.NewStringSet("$[*]", "$['items'][*]['tags'][*]"),
			expected: `{}`,
		},
		{
			name: "Nested selection does not override a parent",
			record: `{
				"id": "msg1",
				"payload": {
					"body": {"data": "xyz", "size": 77},
					"headers": {
						"From": "alice"
					},
					"food": [
						{"id": "001", "value": "Fish"},
						{"id": "002", "value": "Chicken"}
					]
				}
			}`,
			fields: datautils.NewStringSet("$['payload']", "$['payload']['food'][*]['value']"),
			expected: `{
				"payload": {
					"body": {"data": "xyz", "size": 77},
					"headers": {
						"From": "alice"
					},
					"food": [
						{"id": "001", "value": "Fish"},
						{"id": "002", "value": "Chicken"}
					]
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var record map[string]any
			if err := json.Unmarshal([]byte(tt.record), &record); err != nil {
				t.Fatalf("failed to unmarshal record: %v", err)
			}

			got := SelectFields(record, tt.fields)

			var expected map[string]any
			if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatalf("failed to unmarshal expected: %v", err)
			}

			if diff := cmp.Diff(expected, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
