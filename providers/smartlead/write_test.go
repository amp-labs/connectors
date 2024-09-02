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
			Name:         "Mime response header expected",
			Input:        common.WriteParams{ObjectName: "campaigns", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
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
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(responseAccountMissingErr)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Email account not found!"), // nolint:goerr113
			},
		},
		{
			Name:  "Invalid field when creating campaign",
			Input: common.WriteParams{ObjectName: "campaigns", RecordData: "dummy"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseCampaignInvalidFieldErr)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New(`"clinet_id" is not allowed`), // nolint:goerr113
			},
		},
		{
			Name:  "Create new email campaign",
			Input: common.WriteParams{ObjectName: "campaigns", RecordData: "dummy"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseCampaign)
				})
			})),
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
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write(responseClient)
				})
			})),
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
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseAccount)
				})
			})),
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
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
