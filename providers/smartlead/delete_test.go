package smartlead

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

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseNotFoundErr := testutils.DataFromFile(t, "delete-not-found-resource.json")
	responseCampaign := testutils.DataFromFile(t, "delete-campaign.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "campaigns"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:   "Cannot remove unknown object",
			Input:  common.DeleteParams{ObjectName: "coupons", RecordId: "132"},
			Server: mockserver.Dummy(),
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Cannot remove missing campaign",
			Input: common.DeleteParams{ObjectName: "campaigns", RecordId: "782647"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFoundErr),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError(`Campaign not found - Invalid campaign_id.`),
			},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "campaigns", RecordId: "782647"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodDELETE(),
				Then:  mockserver.Response(http.StatusOK, responseCampaign),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
