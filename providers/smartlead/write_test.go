package smartlead

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseAccountMissingErr := testutils.DataFromFile(t, "write-account-missing.json")
	responseCampaignInvalidFieldErr := testutils.DataFromFile(t, "write-invalid-field.json")
	responseCampaign := testutils.DataFromFile(t, "write-campaign-new.json")
	responseClient := testutils.DataFromFile(t, "write-client-new.json")
	responseAccount := testutils.DataFromFile(t, "write-account-new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "campaigns"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.WriteParams{ObjectName: "orders", RecordData: "dummy"},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Update non-existent Email Account",
			Input: common.WriteParams{ObjectName: "email-accounts", RecordId: "08037", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseAccountMissingErr),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Email account not found!"),
			},
		},
		{
			Name:  "Invalid field when creating campaign",
			Input: common.WriteParams{ObjectName: "campaigns", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseCampaignInvalidFieldErr),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New(`"clinet_id" is not allowed`),
			},
		},
		{
			Name:  "Create new email campaign",
			Input: common.WriteParams{ObjectName: "campaigns", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseCampaign),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "552906",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create new client",
			Input: common.WriteParams{ObjectName: "client", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseClient),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "18402",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create new email account",
			Input: common.WriteParams{ObjectName: "email-accounts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseAccount),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "2849",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
