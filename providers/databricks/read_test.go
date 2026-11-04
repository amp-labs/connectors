package databricks

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

const statementsEndpoint = "/api/2.0/sql/statements"

const wellsFirstPage = `{
  "statement_id": "01f0",
  "status": {"state": "SUCCEEDED"},
  "manifest": {"schema": {"columns": [
    {"name": "well_id", "type_name": "LONG"},
    {"name": "name", "type_name": "STRING"},
    {"name": "depth_ft", "type_name": "DOUBLE"},
    {"name": "active", "type_name": "BOOLEAN"}
  ]}},
  "result": {"data_array": [
    ["1", "Eagle Ford 12H", "14250.5", "true"],
    ["2", "Permian 7A", null, "false"]
  ]}
}`

const wellsLastPage = `{
  "statement_id": "01f1",
  "status": {"state": "SUCCEEDED"},
  "manifest": {"schema": {"columns": [
    {"name": "well_id", "type_name": "LONG"}
  ]}},
  "result": {"data_array": [["3"]]}
}`

const statementFailed = `{
  "statement_id": "01f2",
  "status": {"state": "FAILED", "error": {"error_code": "BAD_REQUEST", "message": "TABLE_OR_VIEW_NOT_FOUND"}}
}`

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testconn.TestCaseRead{
		{
			Name: "Full page restores column types and sets the next offset",
			Input: common.ReadParams{
				ObjectName: "wells",
				Fields:     connectors.Fields("name", "depth_ft", "active"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(statementsEndpoint),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, []byte(wellsFirstPage)),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"name": "Eagle Ford 12H", "depth_ft": 14250.5, "active": true},
						Raw: map[string]any{
							"well_id": int64(1), "name": "Eagle Ford 12H", "depth_ft": 14250.5, "active": true,
						},
						Id: "1",
					},
					{
						Fields: map[string]any{"name": "Permian 7A", "depth_ft": nil, "active": false},
						Raw: map[string]any{
							"well_id": int64(2), "name": "Permian 7A", "depth_ft": nil, "active": false,
						},
						Id: "2",
					},
				},
				NextPage: "2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Short page is the last page",
			Input: common.ReadParams{
				ObjectName: "wells",
				Fields:     connectors.Fields("well_id"),
				PageSize:   2,
				NextPage:   "2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(statementsEndpoint),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, []byte(wellsLastPage)),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"well_id": int64(3)},
						Raw:    map[string]any{"well_id": int64(3)},
						Id:     "3",
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Failed statement is an error",
			Input: common.ReadParams{
				ObjectName: "missing",
				Fields:     connectors.Fields("well_id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(statementsEndpoint),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, []byte(statementFailed)),
			}.Server(),
			ExpectedErrs: []error{errStatementNotSucceeded},
		},
		{
			Name: "Invalid next page is rejected",
			Input: common.ReadParams{
				ObjectName: "wells",
				Fields:     connectors.Fields("well_id"),
				NextPage:   "not-a-number",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errInvalidNextPage},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestBuildReadStatement(t *testing.T) {
	t.Parallel()

	connector, err := constructTestConnector("http://localhost")
	if err != nil {
		t.Fatalf("construct connector: %v", err)
	}

	since := time.Date(2026, 11, 1, 8, 30, 0, 0, time.UTC)
	until := time.Date(2026, 11, 2, 8, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		params         common.ReadParams
		wantStatement  string
		wantParameters []statementParameter
	}{
		{
			name: "backfill selects the primary key and orders by timestamp then key",
			params: common.ReadParams{
				ObjectName: "wells",
				Fields:     connectors.Fields("name", "depth_ft"),
			},
			wantStatement: "SELECT `depth_ft`, `name`, `well_id` FROM `main`.`openworks`.`wells` " +
				"ORDER BY `updated_at` ASC, `well_id` ASC LIMIT 1000 OFFSET 0",
		},
		{
			name: "since and until bound the timestamp column",
			params: common.ReadParams{
				ObjectName: "wells",
				Fields:     connectors.Fields("well_id"),
				Since:      since,
				Until:      until,
				PageSize:   50,
				NextPage:   "100",
			},
			wantStatement: "SELECT `well_id` FROM `main`.`openworks`.`wells` " +
				"WHERE `updated_at` >= :since AND `updated_at` <= :until " +
				"ORDER BY `updated_at` ASC, `well_id` ASC LIMIT 50 OFFSET 100",
			wantParameters: []statementParameter{
				{Name: "since", Value: "2026-11-01T08:30:00.000000Z", Type: "TIMESTAMP"},
				{Name: "until", Value: "2026-11-02T08:30:00.000000Z", Type: "TIMESTAMP"},
			},
		},
		{
			name: "identifiers are escaped",
			params: common.ReadParams{
				ObjectName: "we`lls",
				Fields:     connectors.Fields("well_id"),
			},
			wantStatement: "SELECT `well_id` FROM `main`.`openworks`.`we``lls` " +
				"ORDER BY `updated_at` ASC, `well_id` ASC LIMIT 1000 OFFSET 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := connector.buildReadStatement(tt.params)
			if err != nil {
				t.Fatalf("build statement: %v", err)
			}

			if got.Statement != tt.wantStatement {
				t.Fatalf("statement:\n got %s\nwant %s", got.Statement, tt.wantStatement)
			}

			if len(got.Parameters) != len(tt.wantParameters) {
				t.Fatalf("parameters: got %v, want %v", got.Parameters, tt.wantParameters)
			}

			for index, want := range tt.wantParameters {
				if got.Parameters[index] != want {
					t.Fatalf("parameter %d: got %v, want %v", index, got.Parameters[index], want)
				}
			}

			if got.WarehouseID != "wh-123" {
				t.Fatalf("warehouse id: got %q", got.WarehouseID)
			}
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "mock.example.com",
		Metadata: map[string]string{
			"warehouseId":     "wh-123",
			"catalog":         "main",
			"schema":          "openworks",
			"primaryKey":      "well_id",
			"timestampColumn": "updated_at",
		},
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
