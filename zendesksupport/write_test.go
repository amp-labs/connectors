package zendesksupport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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

	// server-error.json occurs when trying to Create object without payload name.
	// ex: for tickets payload must have { "ticket": {...} }

	responseMissingParameterError := mockutils.DataFromFile(t, "missing-parameter.json")
	responseDuplicateError := mockutils.DataFromFile(t, "duplicate-error.json")
	responseRecordValidationError := mockutils.DataFromFile(t, "record-validation.json")
	createBrand := mockutils.DataFromFile(t, "create-brand.json")

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
			input: common.WriteParams{ObjectName: "signals"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name:  "Missing write parameter",
			input: common.WriteParams{ObjectName: "brands"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseMissingParameterError)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Parameter brands is required"), // nolint:goerr113
			},
		},
		{
			name:  "Record validation with single detail",
			input: common.WriteParams{ObjectName: "brands"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseDuplicateError)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New("[RecordInvalid]Record validation errors"),               // nolint:goerr113
				errors.New("[DuplicateValue]Subdomain: nk2 has already been taken"), // nolint:goerr113
			},
		},
		{
			name:  "Record validation with multiple details is split into dedicated errors",
			input: common.WriteParams{ObjectName: "brands"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseRecordValidationError)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New("[RecordInvalid]Record validation errors"),        // nolint:goerr113
				errors.New("[InvalidValue]Subdomain: is invalid"),            // nolint:goerr113
				errors.New("[InvalidFormat]Email is not properly formatted"), // nolint:goerr113
				errors.New("[BlankValue]Name: cannot be blank"),              // nolint:goerr113
			},
		},
		{
			name:  "Write must act as a Create",
			input: common.WriteParams{ObjectName: "brands"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
				})
			})),
			expected:     &common.WriteResult{Success: true},
			expectedErrs: nil,
		},
		{
			name:  "Write must act as an Update",
			input: common.WriteParams{ObjectName: "brands", RecordId: "31207417638931"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "PUT", func() {
					w.WriteHeader(http.StatusOK)
				})
			})),
			expected:     &common.WriteResult{Success: true},
			expectedErrs: nil,
		},
		{
			name:  "Valid creation of a brand",
			input: common.WriteParams{ObjectName: "brands"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(createBrand)
				})
			})),
			comparator: func(actual, expected *common.WriteResult) bool {
				return mockutils.WriteResultComparator.SubsetData(actual, expected)
			},
			expected: &common.WriteResult{
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
