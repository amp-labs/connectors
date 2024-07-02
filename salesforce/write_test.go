package salesforce

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/go-test/deep"
)

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseUnknownField := mockutils.DataFromFile(t, "unknown-field.json")
	responseInvalidFieldUpsert := mockutils.DataFromFile(t, "invalid-field-upsert.json")
	responseCreateOK := mockutils.DataFromFile(t, "create-ok.json")
	responseOKWithErrors := mockutils.DataFromFile(t, "success-with-errors.json")

	tests := []struct {
		name         string
		input        common.WriteParams
		server       *httptest.Server
		connector    Connector
		comparator   func(actual, expected *common.WriteResult) bool // custom comparison
		expected     *common.WriteResult
		expectedErrs []error
	}{
		{
			name: "Write object must be included",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{common.ErrMissingObjects},
		},
		{
			name:  "Mime response header expected",
			input: common.WriteParams{ObjectName: "account"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name:  "Error response understood for creating with unknown field",
			input: common.WriteParams{ObjectName: "account", RecordId: "003ak000004dQCUAA2"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseUnknownField)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New("No such column 'AccountNumer' on sobject of type Account"), // nolint:goerr113
			},
		},
		{
			name:  "Error response understood for updating reserved field",
			input: common.WriteParams{ObjectName: "account", RecordId: "003ak000004dQCUAA2"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseInvalidFieldUpsert)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Unable to create/update fields: MasterRecordId"), // nolint:goerr113
			},
		},
		{
			name:  "Write must act as a Create",
			input: common.WriteParams{ObjectName: "account"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					// cannot be update path
					mockutils.RespondToMissingQueryParameters(w, r, []string{"_HttpMethod"}, func() {
						w.WriteHeader(http.StatusOK)
					})
				})
			})),
			expected:     &common.WriteResult{Success: true},
			expectedErrs: nil,
		},
		{
			name:  "Write must act as an Update",
			input: common.WriteParams{ObjectName: "account", RecordId: "003ak000004dQCUAA2"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					mockutils.RespondToQueryParameters(w, r, url.Values{
						"_HttpMethod": []string{"PATCH"},
					}, func() {
						w.WriteHeader(http.StatusOK)
					})
				})
			})),
			expected:     &common.WriteResult{Success: true},
			expectedErrs: nil,
		},
		{
			name:  "Valid creation of account",
			input: common.WriteParams{ObjectName: "accounts"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseCreateOK)
				})
			})),
			expected: &common.WriteResult{
				Success:  true,
				RecordId: "001ak00000OQTieAAH",
				Errors:   []any{},
				Data:     nil,
			},
			expectedErrs: nil,
		},
		{
			name:  "OK Response, but with errors field",
			input: common.WriteParams{ObjectName: "accounts"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseOKWithErrors)
				})
			})),
			expected: &common.WriteResult{
				Success:  false,
				RecordId: "001RM000003oLruYAE",
				Errors: []any{map[string]any{
					"statusCode": "MALFORMED_ID",
					"message":    "malformed id 001RM000003oLrB000",
					"fields":     []any{},
				}},
				Data: nil,
			},
			expectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			ctx := context.Background()

			connector, err := NewConnector(
				WithAuthenticatedClient(http.DefaultClient),
				WithWorkspace("test-workspace"),
			)
			if err != nil {
				t.Fatalf("%s: error in test while constructing connector %v", tt.name, err)
			}

			// for testing we want to redirect calls to our server
			connector.setBaseURL(tt.server.URL)

			// start of tests
			output, err := connector.Write(ctx, tt.input)
			if len(tt.expectedErrs) == 0 && err != nil {
				t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
			}

			if len(tt.expectedErrs) != 0 && err == nil {
				t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
			}

			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
				}
			}

			// compare desired output
			var ok bool
			if tt.comparator == nil {
				// default comparison is concerned about all fields
				ok = reflect.DeepEqual(output, tt.expected)
			} else {
				ok = tt.comparator(output, tt.expected)
			}

			if !ok {
				diff := deep.Equal(output, tt.expected)
				t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)", tt.name, tt.expected, output, diff)
			}
		})
	}
}
