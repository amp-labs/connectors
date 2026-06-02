package monday

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen
	t.Parallel()

	introspectionResponse := `{
		"data": {
			"__type": {
				"name": "Item",
				"fields": [
					{"name": "id", "type": {"name": "ID", "kind": "SCALAR"}},
					{"name": "name", "type": {"name": "String", "kind": "SCALAR"}}
				]
			}
		}
	}`

	columnsResponse := `{
		"data": {
			"boards": [{
				"columns": [
					{"id": "status", "title": "Status", "type": "status", "settings_str": "{\"labels\":{\"0\":\"Working on it\",\"1\":\"Done\"}}"},
					{"id": "text", "title": "Text", "type": "text", "settings_str": ""}
				]
			}]
		}
	}`

	boardsIntrospectionResponse := `{
		"data": {
			"__type": {
				"name": "Board",
				"fields": [
					{"name": "id", "type": {"name": "ID", "kind": "SCALAR"}},
					{"name": "name", "type": {"name": "String", "kind": "SCALAR"}}
				]
			}
		}
	}`

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Boards metadata does not fetch column definitions",
			Input: []string{"boards"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Check(func(_ http.ResponseWriter, r *http.Request) bool {
					return requestBodyContains(r, "__type")
				}),
				Then: mockserver.Response(http.StatusOK, []byte(boardsIntrospectionResponse)),
				Else: mockserver.Response(http.StatusTeapot, []byte(`{"error":"unexpected request"}`)),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"boards": {
						DisplayName: "Boards",
						Fields: map[string]common.FieldMetadata{
							"id":   {DisplayName: "id", ValueType: common.ValueTypeOther, ProviderType: ""},
							"name": {DisplayName: "name", ValueType: common.ValueTypeOther, ProviderType: ""},
						},
					},
				},
			},
		},
		{
			Name:  "Items metadata merges board column definitions",
			Input: []string{"items@123"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.Check(func(_ http.ResponseWriter, r *http.Request) bool {
							return requestBodyContains(r, "__type")
						}),
						Then: mockserver.Response(http.StatusOK, []byte(introspectionResponse)),
					},
					{
						If: mockcond.Check(func(_ http.ResponseWriter, r *http.Request) bool {
							return requestBodyContains(r, "boards(ids")
						}),
						Then: mockserver.Response(http.StatusOK, []byte(columnsResponse)),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"items@123": {
						DisplayName: "Items",
						Fields: map[string]common.FieldMetadata{
							"id":   {DisplayName: "id", ValueType: common.ValueTypeOther, ProviderType: ""},
							"name": {DisplayName: "name", ValueType: common.ValueTypeOther, ProviderType: ""},
							"cf_status": {
								DisplayName:  "Status",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "status",
								Values: []common.FieldValue{
									{Value: "Done", DisplayValue: "Done"},
									{Value: "Working on it", DisplayValue: "Working on it"},
								},
								IsCustom: new(true),
							},
							"cf_text": {
								DisplayName:  "Text",
								ValueType:    common.ValueTypeString,
								ProviderType: "text",
								IsCustom:     new(true),
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func requestBodyContains(r *http.Request, needle string) bool {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}

	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return strings.Contains(string(body), needle)
}
