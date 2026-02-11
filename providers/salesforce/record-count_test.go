package salesforce

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
)

func TestGetRecordCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name         string
		Input        *common.RecordCountParams
		Server       *httptest.Server
		Expected     *common.RecordCountResult
		ExpectedErrs []error
	}{
		{
			Name:  "Successful count query",
			Input: &common.RecordCountParams{ObjectName: "Account"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/query"),
					mockcond.QueryParam("q", "SELECT COUNT() FROM Account"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{
					"totalSize": 42,
					"done": true,
					"records": []
				}`),
			}.Server(),
			Expected:     &common.RecordCountResult{Count: 42},
			ExpectedErrs: nil,
		},
		{
			Name:  "Count with zero results",
			Input: &common.RecordCountParams{ObjectName: "Lead"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/query"),
					mockcond.QueryParam("q", "SELECT COUNT() FROM Lead"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{
					"totalSize": 0,
					"done": true,
					"records": []
				}`),
			}.Server(),
			Expected:     &common.RecordCountResult{Count: 0},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			defer tt.Server.Close()

			connector, err := constructTestConnector(tt.Server.URL)
			if err != nil {
				t.Fatalf("failed to construct connector: %v", err)
			}

			result, err := connector.GetRecordCount(t.Context(), tt.Input)

			if len(tt.ExpectedErrs) > 0 {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				// Just check that an error occurred
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Count != tt.Expected.Count {
				t.Errorf("expected count %d, got %d", tt.Expected.Count, result.Count)
			}
		})
	}
}
