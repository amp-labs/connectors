package databricks

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

const wellsColumns = `{
  "statement_id": "01f3",
  "status": {"state": "SUCCEEDED"},
  "manifest": {"schema": {"columns": [
    {"name": "column_name", "type_name": "STRING"},
    {"name": "data_type", "type_name": "STRING"}
  ]}},
  "result": {"data_array": [
    ["well_id", "BIGINT"],
    ["name", "STRING"],
    ["depth_ft", "DOUBLE"],
    ["updated_at", "TIMESTAMP"]
  ]}
}`

const noColumns = `{
  "statement_id": "01f4",
  "status": {"state": "SUCCEEDED"},
  "manifest": {"schema": {"columns": [
    {"name": "column_name", "type_name": "STRING"},
    {"name": "data_type", "type_name": "STRING"}
  ]}},
  "result": {"data_array": []}
}`

func TestListObjectMetadata(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Columns from information_schema become typed fields",
			Input: []string{"wells"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(statementsEndpoint),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, []byte(wellsColumns)),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"wells": {
						DisplayName: "wells",
						Fields: map[string]common.FieldMetadata{
							"well_id":    {DisplayName: "well_id", ValueType: common.ValueTypeInt, ProviderType: "BIGINT"},
							"name":       {DisplayName: "name", ValueType: common.ValueTypeString, ProviderType: "STRING"},
							"depth_ft":   {DisplayName: "depth_ft", ValueType: common.ValueTypeFloat, ProviderType: "DOUBLE"},
							"updated_at": {DisplayName: "updated_at", ValueType: common.ValueTypeDateTime, ProviderType: "TIMESTAMP"},
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Table without columns is reported per object",
			Input: []string{"missing"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(statementsEndpoint),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, []byte(noColumns)),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"missing": mockutils.ExpectedSubsetErrors{errObjectNotFound},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
