package salesfinity

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// nolint:funlen,gocognit,cyclop
func TestRead(t *testing.T) {
	t.Parallel()
	responseReadEmpty := testutils.DataFromFile(t, "read-empty.json")
	responseCallLog := testutils.DataFromFile(t, "read-call-log.json")
	responseCallLogFirstPage := testutils.DataFromFile(t, "call-log.json")
	responseCallLogLastPage := testutils.DataFromFile(t, "read-call-log-last-page.json")
	responseContactListsCsv := testutils.DataFromFile(t, "contact-lists-csv.json")
	tests := []testroutines.Read{

		{
			Name: "Read empty items",
			Input: common.ReadParams{
				ObjectName: "call-log",
				Fields:     connectors.Fields("_id", "updatedAt"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/call-log"),
					mockcond.QueryParam("limit", "10"),
				},
				Then: mockserver.Response(http.StatusOK, responseReadEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read call-log",
			Input: common.ReadParams{
				ObjectName: "call-log",
				Fields:     connectors.Fields("_id", "to", "updatedAt", "from"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/call-log"),
					mockcond.QueryParam("limit", "10"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallLog),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"_id":       "test_call_id_123",
							"to":        "(555) 123-4567",
							"updatedat": "2024-01-15T11:00:00.000Z",
							"from":      "15551234567",
						},
						Raw: map[string]any{
							"_id":       "test_call_id_123",
							"to":        "(555) 123-4567",
							"updatedAt": "2024-01-15T11:00:00.000Z",
							"from":      "15551234567",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read call-log first page with pagination",
			Input: common.ReadParams{
				ObjectName: "call-log",
				Fields:     connectors.Fields("_id", "updatedAt"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/call-log"),
					mockcond.QueryParam("limit", "10"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallLogFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"_id":       "test_call_id_123",
							"updatedat": "2024-01-15T11:00:00.000Z",
						},
						Raw: map[string]any{
							"_id":       "test_call_id_123",
							"updatedAt": "2024-01-15T11:00:00.000Z",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/v1/call-log?limit=10&page=2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read call-log second page using NextPage token",
			Input: common.ReadParams{
				ObjectName: "call-log",
				Fields:     connectors.Fields("_id", "updatedAt", "from"),
				NextPage:   testroutines.URLTestServer + "/v1/call-log?limit=10&page=2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/call-log"),
					mockcond.QueryParam("page", "2"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallLogLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"_id":       "test_call_id_456",
							"updatedat": "2024-01-15T12:30:00.000Z",
							"from":      "15559876543",
						},
						Raw: map[string]any{
							"_id":       "test_call_id_456",
							"updatedAt": "2024-01-15T12:30:00.000Z",
							"from":      "15559876543",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read call-log with PageSize uses limit query param",
			Input: common.ReadParams{
				ObjectName: "call-log",
				Fields:     connectors.Fields("_id", "updatedAt"),
				PageSize:   50,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/call-log"),
					mockcond.QueryParam("limit", "50"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallLog),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"_id":       "test_call_id_123",
							"updatedat": "2024-01-15T11:00:00.000Z",
						},
						Raw: map[string]any{
							"_id":       "test_call_id_123",
							"updatedAt": "2024-01-15T11:00:00.000Z",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contact-lists/csv",
			Input: common.ReadParams{
				ObjectName: "contact-lists/csv",
				Fields:     connectors.Fields("_id", "name", "user"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/contact-lists/csv"),
					mockcond.QueryParam("limit", "10"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactListsCsv),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"_id":  "test_list_id_001",
							"name": "Test Contact List 1",
							"user": "test_user_id_001",
						},
						Raw: map[string]any{
							"_id":  "test_list_id_001",
							"name": "Test Contact List 1",
							"user": "test_user_id_001",
						},
					},
					{
						Fields: map[string]any{
							"_id":  "test_list_id_002",
							"name": "Test Contact List 2",
							"user": "test_user_id_002",
						},
						Raw: map[string]any{
							"_id":  "test_list_id_002",
							"name": "Test Contact List 2",
							"user": "test_user_id_002",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read call-log with Since filters records connector-side",
			Input: common.ReadParams{
				ObjectName: "call-log",
				Fields:     connectors.Fields("_id", "updatedAt"),
				Since:      time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/call-log"),
					mockcond.QueryParam("limit", "10"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallLog),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"_id":       "test_call_id_123",
							"updatedat": "2024-01-15T11:00:00.000Z",
						},
						Raw: map[string]any{
							"_id":       "test_call_id_123",
							"updatedAt": "2024-01-15T11:00:00.000Z",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read call-log with Since after all records returns empty",
			Input: common.ReadParams{
				ObjectName: "call-log",
				Fields:     connectors.Fields("_id", "updatedAt"),
				Since:      time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/call-log"),
					mockcond.QueryParam("limit", "10"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallLogFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contact-lists/csv with Since does not filter (no time-based filtering)",
			Input: common.ReadParams{
				ObjectName: "contact-lists/csv",
				Fields:     connectors.Fields("_id", "name"),
				Since:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/contact-lists/csv"),
					mockcond.QueryParam("limit", "10"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactListsCsv),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"_id": "test_list_id_001", "name": "Test Contact List 1"},
						Raw:    map[string]any{"_id": "test_list_id_001", "name": "Test Contact List 1"},
					},
					{
						Fields: map[string]any{"_id": "test_list_id_002", "name": "Test Contact List 2"},
						Raw:    map[string]any{"_id": "test_list_id_002", "name": "Test Contact List 2"},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
