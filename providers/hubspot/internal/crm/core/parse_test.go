package core

import (
	"testing"

	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/spyzhov/ajson"
)

func TestGetNextRecordsAfter(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    string
		wantErr error
	}{
		{
			name: "Success",
			json: `{
				"paging": {
					"next": {
						"after": "394",
						"link": "https://api.hubapi.com/crm/v3/objects/contacts?limit=100&properties=listId%2Cname&after=394"
					}
				}
			}`,
			want: "394",
		},
		{
			name: "No paging",
			json: `{"results": []}`,
			want: "",
		},
		{
			name:    "Paging not an object",
			json:    `{"paging": "not-an-object"}`,
			wantErr: jsonquery.ErrNotObject,
		},
		{
			name: "Next missing in paging",
			json: `{"paging": {}}`,
			want: "",
		},
		{
			name:    "Next not an object",
			json:    `{"paging": {"next": "not-an-object"}}`,
			wantErr: jsonquery.ErrNotObject,
		},
		{
			name: "After missing in next",
			json: `{"paging": {"next": {}}}`,
			want: "",
		},
		{
			name:    "After not a string",
			json:    `{"paging": {"next": {"after": 123}}}`,
			wantErr: testutils.StringError("JSON value is not a string"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ajson.Unmarshal([]byte(tt.json))
			if err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}

			got, err := GetNextRecordsAfter(node)
			testutils.CheckOutputWithError(t, tt.name, tt.want, tt.wantErr, got, err)
		})
	}
}

func TestGetNextRecordsURL(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    string
		wantErr error
	}{
		{
			name: "Success",
			json: `{
				"paging": {
					"next": {
						"after": "394",
						"link": "https://api.hubapi.com/crm/v3/objects/contacts?limit=100&properties=listId%2Cname&after=394"
					}
				}
			}`,
			want: "https://api.hubapi.com/crm/v3/objects/contacts?limit=100&properties=listId%2Cname&after=394",
		},
		{
			name: "No paging",
			json: `{"results": []}`,
			want: "",
		},
		{
			name:    "Link not a string",
			json:    `{"paging": {"next": {"link": 123}}}`,
			wantErr: testutils.StringError("JSON value is not a string"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ajson.Unmarshal([]byte(tt.json))
			if err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}

			got, err := GetNextRecordsURL(node)
			testutils.CheckOutputWithError(t, tt.name, tt.want, tt.wantErr, got, err)
		})
	}
}

func TestGetRecords(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    []map[string]any
		wantErr error
	}{
		{
			name: "Success",
			json: `{
				"results": [
					{"id": "1", "name": "test1"},
					{"id": "2", "name": "test2"}
				]
			}`,
			want: []map[string]any{
				{"id": "1", "name": "test1"},
				{"id": "2", "name": "test2"},
			},
		},
		{
			name:    "Results missing",
			json:    `{}`,
			wantErr: jsonquery.ErrKeyNotFound,
		},
		{
			name:    "Results not an array",
			json:    `{"results": "not-an-array"}`,
			wantErr: jsonquery.ErrNotArray,
		},
		{
			name:    "Result item not an object",
			json:    `{"results": [123]}`,
			wantErr: jsonquery.ErrNotObject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ajson.Unmarshal([]byte(tt.json))
			if err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}

			got, err := GetRecords(node)
			testutils.CheckOutputWithError(t, tt.name, tt.want, tt.wantErr, got, err)
		})
	}
}

func TestGetNextRecordsURLCRM(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    string
		wantErr error
	}{
		{
			name: "Success with hasMore true",
			json: `{
				"hasMore": true,
				"offset": 100
			}`,
			want: "100",
		},
		{
			name: "hasMore false",
			json: `{
				"hasMore": false,
				"offset": 100
			}`,
			want: "",
		},
		{
			name: "hasMore missing",
			json: `{
				"offset": 100
			}`,
			want: "",
		},
		{
			name: "offset missing",
			json: `{
				"hasMore": true
			}`,
			want: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ajson.Unmarshal([]byte(tt.json))
			if err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}

			got, err := GetNextRecordsURLCRM(node)
			testutils.CheckOutputWithError(t, tt.name, tt.want, tt.wantErr, got, err)
		})
	}
}
