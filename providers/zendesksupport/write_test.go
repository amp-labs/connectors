package zendesksupport

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

	// server-error.json occurs when trying to Create object without payload name.
	// ex: for tickets payload must have { "ticket": {...} }

	responseMissingParameterError := testutils.DataFromFile(t, "missing-parameter.json")
	responseDuplicateError := testutils.DataFromFile(t, "duplicate-error.json")
	responseRecordValidationError := testutils.DataFromFile(t, "record-validation.json")
	createBrand := testutils.DataFromFile(t, "create-brand.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Mime response header expected",
			Input:        common.WriteParams{ObjectName: "signals"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Missing write parameter",
			Input: common.WriteParams{ObjectName: "brands"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseMissingParameterError)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Parameter brands is required"), // nolint:goerr113
			},
		},
		{
			Name:  "Record validation with single detail",
			Input: common.WriteParams{ObjectName: "brands"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseDuplicateError)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("[RecordInvalid]Record validation errors"),               // nolint:goerr113
				errors.New("[DuplicateValue]Subdomain: nk2 has already been taken"), // nolint:goerr113
			},
		},
		{
			Name:  "Record validation with multiple details is split into dedicated errors",
			Input: common.WriteParams{ObjectName: "brands"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseRecordValidationError)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("[RecordInvalid]Record validation errors"),        // nolint:goerr113
				errors.New("[InvalidValue]Subdomain: is invalid"),            // nolint:goerr113
				errors.New("[InvalidFormat]Email is not properly formatted"), // nolint:goerr113
				errors.New("[BlankValue]Name: cannot be blank"),              // nolint:goerr113
			},
		},
		{
			Name:  "Write must act as a Create",
			Input: common.WriteParams{ObjectName: "brands"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
				})
			})),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Write must act as an Update",
			Input: common.WriteParams{ObjectName: "brands", RecordId: "31207417638931"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "PUT", func() {
					w.WriteHeader(http.StatusOK)
				})
			})),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of a brand",
			Input: common.WriteParams{ObjectName: "brands"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(createBrand)
				})
			})),
			Comparator: func(serverURL string, actual, expected *common.WriteResult) bool {
				return mockutils.WriteResultComparator.SubsetData(actual, expected)
			},
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "31207417638931",
				Errors:   nil,
				Data: map[string]any{
					"id":        float64(31207417638931),
					"name":      "Nike",
					"brand_url": "https://nkn2.zendesk.com",
					"subdomain": "nkn2",
					"active":    true,
					"default":   false,
				},
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
