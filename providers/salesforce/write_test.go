package salesforce

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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

	responseUnknownField := testutils.DataFromFile(t, "unknown-field.json")
	responseInvalidFieldUpsert := testutils.DataFromFile(t, "invalid-field-upsert.json")
	responseCreateOK := testutils.DataFromFile(t, "create-ok.json")
	responseOKWithErrors := testutils.DataFromFile(t, "success-with-errors.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "account"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Mime response header expected",
			Input:        common.WriteParams{ObjectName: "account", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Error response understood for creating with unknown field",
			Input: common.WriteParams{ObjectName: "account", RecordId: "003ak000004dQCUAA2", RecordData: "dummy"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseUnknownField)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("No such column 'AccountNumer' on sobject of type Account"), // nolint:goerr113
			},
		},
		{
			Name:  "Error response understood for updating reserved field",
			Input: common.WriteParams{ObjectName: "account", RecordId: "003ak000004dQCUAA2", RecordData: "dummy"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseInvalidFieldUpsert)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Unable to create/update fields: MasterRecordId"), // nolint:goerr113
			},
		},
		{
			Name:  "Write must act as an Update",
			Input: common.WriteParams{ObjectName: "account", RecordId: "003ak000004dQCUAA2", RecordData: "dummy"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					mockutils.RespondToQueryParameters(w, r, url.Values{
						"_HttpMethod": []string{"PATCH"},
					}, func() {
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write(responseCreateOK)
					})
				})
			})),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "001ak00000OQTieAAH",
				Errors:   []any{},
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of account",
			Input: common.WriteParams{ObjectName: "accounts", RecordData: "dummy"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseCreateOK)
				})
			})),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "001ak00000OQTieAAH",
				Errors:   []any{},
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "OK Response, but with errors field",
			Input: common.WriteParams{ObjectName: "accounts", RecordData: "dummy"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseOKWithErrors)
				})
			})),
			Expected: &common.WriteResult{
				Success:  false,
				RecordId: "001RM000003oLruYAE",
				Errors: []any{map[string]any{
					"statusCode": "MALFORMED_ID",
					"message":    "malformed id 001RM000003oLrB000",
					"fields":     []any{},
				}},
				Data: nil,
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
