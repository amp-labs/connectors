package hubspot

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
			Name:  "Successful count query for contacts",
			Input: &common.RecordCountParams{ObjectName: "contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm/v3/objects/contacts/search"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{
					"total": 1234,
					"results": []
				}`),
			}.Server(),
			Expected:     &common.RecordCountResult{Count: 1234},
			ExpectedErrs: nil,
		},
		{
			Name:  "Count with zero results",
			Input: &common.RecordCountParams{ObjectName: "deals"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm/v3/objects/deals/search"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{
					"total": 0,
					"results": []
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
