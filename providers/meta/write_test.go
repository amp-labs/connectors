package meta

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

func TestFacebookWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	writeResponse := testutils.DataFromFile(t, "write_response.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "creating the ad labels",
			Input: common.WriteParams{ObjectName: "adlabels", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v19.0/act_1214321106978726/adlabels"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, writeResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "120228422745080770",
				Errors:   nil,
				Data: map[string]any{
					"id": "120228422745080770",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Creating the business users",
			Input: common.WriteParams{ObjectName: "business_users", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v19.0/1190021932394709/business_users"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, writeResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "120228422745080770",
				Errors:   nil,
				Data: map[string]any{
					"id": "120228422745080770",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestFacebookConnector(tt.Server.URL)
			})
		})
	}
}
