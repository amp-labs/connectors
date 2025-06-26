package instantly

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	customTagsResponse := testutils.DataFromFile(t, "write_custom_tags.json")
	accountsResponse := testutils.DataFromFile(t, "write_accounts.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "creating the custom tags",
			Input: common.WriteParams{ObjectName: "custom-tags", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/custom-tags"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, customTagsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "0196837f-d659-7496-854c-84187bb6f708",
				Errors:   nil,
				Data: map[string]any{
					"id":                "0196837f-d659-7496-854c-84187bb6f708",
					"timestamp_created": "2025-04-29T21:41:55.417Z",
					"timestamp_updated": "2025-04-29T21:41:55.417Z",
					"organization_id":   "0196837f-d659-7496-854c-841903f3a321",
					"label":             "Important",
					"description":       "Used for marking important items",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update custom tags as PATCH",
			Input: common.WriteParams{
				ObjectName: "custom-tags",
				RecordId:   "0196837f-d659-7496-854c-84187bb6f708",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/custom-tags/0196837f-d659-7496-854c-84187bb6f708"),
					mockcond.MethodPATCH(),
				},
				Then: mockserver.Response(http.StatusOK, customTagsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "0196837f-d659-7496-854c-84187bb6f708",
				Errors:   nil,
				Data: map[string]any{
					"id":                "0196837f-d659-7496-854c-84187bb6f708",
					"timestamp_created": "2025-04-29T21:41:55.417Z",
					"timestamp_updated": "2025-04-29T21:41:55.417Z",
					"organization_id":   "0196837f-d659-7496-854c-841903f3a321",
					"label":             "Important",
					"description":       "Used for marking important items",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Creating the accounts",
			Input: common.WriteParams{ObjectName: "accounts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/accounts"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, accountsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "",
				Errors:   nil,
				Data: map[string]any{
					"email":             "user@example.com",
					"timestamp_created": "2025-04-29T21:41:55.421Z",
					"timestamp_updated": "2025-04-29T21:41:55.421Z",
					"first_name":        "John",
					"last_name":         "Doe",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
