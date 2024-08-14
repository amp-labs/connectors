package smartlead

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils"
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
			Name:         "Write object must be included",
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
			Name:         "Mime response header expected",
			Input:        common.DeleteParams{ObjectName: "campaigns", RecordId: "782647"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
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
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(responseNotFoundErr)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New(`Campaign not found - Invalid campaign_id.`), // nolint:goerr113
			},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "campaigns", RecordId: "782647"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "DELETE", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseCampaign)
				})
			})),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
